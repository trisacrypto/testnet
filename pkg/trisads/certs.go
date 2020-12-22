package trisads

import (
	"archive/zip"
	"bytes"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/testnet/pkg/sectigo"
	"github.com/trisacrypto/testnet/pkg/trisads/pb"
	"software.sslmate.com/src/go-pkcs12"
)

// CertManager is a go routine that periodically checks on the status of certificate
// requests and moves them through the request pipeline. Once CertManager detects a
// certificate request that is ready to submit, it submits the request via the Sectigo
// API. If processing, it checks the batch status, and when it detects that the bact
// is done processing it downloads the certs and emails them to the technical conacts.
// If the certificate processing fails for any reason, it sends and error message to
// the TRISA admins since this will prevent the integrator from joining the TestNet.
//
// TODO: move completed certificate requests to archive so that the CertManger routine
// isn't continuosly handling a growing number of requests over time.
//
// TODO: notify admins if cert-manager errors since this will block integration.
func (s *Server) CertManager() {
	// Check certificate download directory
	if certDir, err := s.getCertStorage(); err != nil {
		log.Fatal().Err(err).Msg("cert-manager cannot access certificate storage")
	} else {
		log.Debug().Str("path", certDir).Msg("certificate download directory")
	}

	// Ticker is created in the go routine to prevent backpressure if the cert manager
	// process takes longer than the specified ticker interval.
	ticker := time.NewTicker(s.conf.CertManInterval)
	log.Info().Dur("interval", s.conf.CertManInterval).Msg("cert-manager process started")

	for {
		// Wait for next tick
		<-ticker.C

		// Retrieve all certificate requests from the database
		var err error
		var careqs []*pb.CertificateRequest
		if careqs, err = s.db.ListCertRequests(); err != nil {
			log.Error().Err(err).Msg("cert-manager could not retrieve certificate requests")
		}
		log.Debug().Int("requests", len(careqs)).Msg("cert-manager checking certificate request pipelines")

		for _, req := range careqs {
			logctx := log.With().Str("id", req.Id).Str("domain", req.CommonName).Logger()

			switch req.Status {
			case pb.CertificateRequestState_READY_TO_SUBMIT:
				if err = s.submitCertificateRequest(req); err != nil {
					logctx.Error().Err(err).Msg("cert-manager could not submit certificate request")
				} else {
					logctx.Info().Msg("certificate request submitted")
				}
			case pb.CertificateRequestState_PROCESSING:
				if err = s.checkCertificateRequest(req); err != nil {
					logctx.Error().Err(err).Msg("cert-manager could not process submitted certificate request")
				} else {
					logctx.Debug().Msg("processing certificate request check complete")
				}
			}
		}
	}
}

func (s *Server) submitCertificateRequest(r *pb.CertificateRequest) (err error) {
	// Step 0: mark the VASP status as issuing certificates
	var vasp *pb.VASP
	if vasp, err = s.db.Retrieve(r.Vasp); err != nil {
		return fmt.Errorf("could not fetch VASP to mark as issuing certificate: %s", err)
	}
	vasp.VerificationStatus = pb.VerificationState_ISSUING_CERTIFICATE
	if err = s.db.Update(vasp); err != nil {
		return fmt.Errorf("could not update VASP status: %s", err)
	}

	// Step 1: find an authority with an available balance
	var authority int
	if authority, err = s.findCertAuthority(); err != nil {
		return err
	}

	// Step 2: create the certificate params, decrypting the PKCS 12 password
	params := make(map[string]string)
	params["commonName"] = r.CommonName
	if params["pkcs12Password"], err = s.Decrypt(r.Pkcs12Password, r.Pkcs12Signature); err != nil {
		return fmt.Errorf("could not decrypt pkcs12password: %s", err)
	}

	// Step 3: submit the certificate
	var rep *sectigo.BatchResponse
	batchName := fmt.Sprintf("%s certificate request for %s (id: %s)", s.conf.DirectoryID, r.CommonName, r.Id)
	if rep, err = s.certs.CreateSingleCertBatch(authority, batchName, params); err != nil {
		return fmt.Errorf("could not create single certificate batch: %s", err)
	}

	// Step 4: update the certificate request with the batch details
	r.AuthorityId = int64(authority)
	r.BatchId = int64(rep.BatchID)
	r.BatchName = rep.BatchName
	r.BatchStatus = rep.Status
	r.OrderNumber = int64(rep.OrderNumber)
	r.CreationDate = rep.CreationDate
	r.Profile = rep.Profile
	r.RejectReason = rep.RejectReason

	// Mark the certificate request as processing so downstream status checks occur
	r.Status = pb.CertificateRequestState_PROCESSING
	if err = s.db.SaveCertRequest(r); err != nil {
		return fmt.Errorf("could not update certificate with batch details: %s", err)
	}

	return nil
}

func (s *Server) checkCertificateRequest(r *pb.CertificateRequest) (err error) {
	if r.BatchId == 0 {
		return errors.New("missing batch ID - cannot retrieve status")
	}

	// Step 1: refresh batch info from the sectigo service
	var info *sectigo.BatchResponse
	if info, err = s.certs.BatchDetail(int(r.BatchId)); err != nil {
		return fmt.Errorf("could not fetch batch info for id %d: %s", r.BatchId, err)
	}

	// Step 1b: update  certificate request with fetched info
	r.BatchStatus = info.Status
	r.RejectReason = info.RejectReason

	// Step 2: get the processing info for the batch
	var proc *sectigo.ProcessingInfoResponse
	if proc, err = s.certs.ProcessingInfo(int(r.BatchId)); err != nil {
		return fmt.Errorf("could not fetch batch processing info for id %d: %s", r.BatchId, err)
	}

	// TODO: make a debug message rather than an info message
	log.Info().
		Str("status", r.BatchStatus).
		Str("reject", r.RejectReason).
		Int("active", proc.Active).
		Int("failed", proc.Failed).
		Int("success", proc.Success).
		Msg("batch processing status")

	// Step 3: check active - if there is still an active batch then delay
	if proc.Active > 0 {
		r.Status = pb.CertificateRequestState_PROCESSING
		if err = s.db.SaveCertRequest(r); err != nil {
			return fmt.Errorf("could not save updated cert request: %s", err)
		}
		return nil
	}

	// Step 4: check failures -- determine if certificate request has been rejected
	if proc.Failed > 0 {
		logctx := log.With().
			Int("batch_id", int(r.BatchId)).
			Int("failed", proc.Failed).
			Int("success", proc.Success).
			Str("status", r.BatchStatus).
			Str("name", r.BatchName).
			Logger()

		if proc.Success > 0 || r.BatchStatus == sectigo.BatchStatusReadyForDownload {
			// This may mean that some certificates can be downloaded, so just log
			// errors and continue with download processing
			logctx.Warn().Msg("certificate request mixed success/failure")
		} else {
			// In this case there were no successes, so set certificate request status accordingly
			// and do not continue processing the certificate request
			if r.RejectReason != "" || r.BatchStatus == sectigo.BatchStatusRejected {
				// Assume the certificate was rejected
				r.Status = pb.CertificateRequestState_CR_REJECTED
				logctx.Warn().Msg("certificate request rejected")
			} else {
				// Assume the certificate errored and wasn't rejected
				r.Status = pb.CertificateRequestState_CR_ERRORED
				logctx.Warn().Msg("certificate request errored")
			}

			if err = s.db.SaveCertRequest(r); err != nil {
				return fmt.Errorf("could not save updated cert request: %s", err)
			}
			return nil
		}
	}

	// Step 5: Check to make sure we can download certificates
	if proc.Success == 0 || info.Status != sectigo.BatchStatusReadyForDownload {
		// We should not be in this state, it should have been handled in Step 4
		// so this is a developer error on our part, or a change in the Sectigo API
		log.Error().Str("status", info.Status).Msg("unhandled sectigo state")
		r.Status = pb.CertificateRequestState_PROCESSING
		if err = s.db.SaveCertRequest(r); err != nil {
			return fmt.Errorf("could not save updated cert request: %s", err)
		}
		return nil
	}

	// Step 6: Mark the status as ready for download!
	r.Status = pb.CertificateRequestState_DOWNLOADING
	if err = s.db.SaveCertRequest(r); err != nil {
		return fmt.Errorf("could not save updated cert request: %s", err)
	}

	// Send off downloader go routine to fetch the certs and notify the user
	go s.downloadCertificateRequest(r)
	return nil
}

// finds the first authority with an available balance greater than 0.
func (s *Server) findCertAuthority() (id int, err error) {
	var authorities []*sectigo.AuthorityResponse
	if authorities, err = s.certs.UserAuthorities(); err != nil {
		return 0, fmt.Errorf("could not fetch user authorities: %s", err)
	}

	for _, authority := range authorities {
		var balance int
		if balance, err = s.certs.AuthorityAvailableBalance(authority.ID); err != nil {
			log.Error().Err(err).Int("authority", authority.ID).Msg("could not fetch authority balance")
		}
		if balance > 0 {
			return authority.ID, nil
		}
	}

	return 0, fmt.Errorf("could not find authority with available balance out of %d available authorities", len(authorities))
}

// a go routine that downloads the certificate in the background, then sends the certs
// as an attachment to the technical contact if available.
func (s *Server) downloadCertificateRequest(r *pb.CertificateRequest) {
	var (
		err     error
		path    string
		certDir string
	)

	// Get the cert storage directory to download certs to
	if certDir, err = s.getCertStorage(); err != nil {
		log.Error().Err(err).Msg("could not find cert storage directory")
		return
	}

	// Download the certificates as a zip file to the cert storage directory
	if path, err = s.certs.Download(int(r.BatchId), certDir); err != nil {
		log.Error().Err(err).Int("batch", int(r.BatchId)).Msg("could not download certificates")
		return
	}

	// Mark as downloaded.
	r.Status = pb.CertificateRequestState_DOWNLOADED
	if err = s.db.SaveCertRequest(r); err != nil {
		log.Error().Err(err).Msg("could not save updated cert request")
		return
	}
	log.Info().Str(path, path).Msg("certificates downloaded")

	// Fetch the VASP to get contact info and store certificate data
	var vasp *pb.VASP
	if vasp, err = s.db.Retrieve(r.Vasp); err != nil {
		log.Error().Err(err).Msg("could not get VASP to store certificates")
		return
	}

	// Unpack the certificate and fetch the public key material
	var pkcs12password string
	if pkcs12password, err = s.Decrypt(r.Pkcs12Password, r.Pkcs12Signature); err != nil {
		log.Error().Err(err).Msg("could not decrypt pkcs12password to extract public key")
		return
	}

	if vasp.Certificate, err = extractCertificate(path, pkcs12password); err != nil {
		log.Error().Err(err).Msg("could not extract certificate")
		return
	}

	// Update the VASP status as verified/certificate issued
	vasp.VerificationStatus = pb.VerificationState_VERIFIED
	if err = s.db.Update(vasp); err != nil {
		log.Error().Err(err).Msg("could not update VASP status as verified")
		return
	}

	// Email the certificates to the technical contacts
	if err = s.DeliverCertificatesEmail(vasp, path); err != nil {
		log.Error().Err(err).Msg("could not deliver certificates to technical contact")
	}

	// Mark certficate request as complete.
	r.Status = pb.CertificateRequestState_COMPLETED
	if err = s.db.SaveCertRequest(r); err != nil {
		log.Error().Err(err).Msg("could not save updated cert request")
		return
	}

	log.Info().
		Str("serial_number", hex.EncodeToString(vasp.Certificate.SerialNumber)).
		Msg("certificates extracted and delivered")
}

func extractCertificate(path, pkcs12password string) (pub *pb.Certificate, err error) {
	var archive *zip.ReadCloser
	if archive, err = zip.OpenReader(path); err != nil {
		return nil, err
	}
	defer archive.Close()

	// Validate archive
	if len(archive.File) == 0 {
		return nil, fmt.Errorf("downloaded ZIP file contained no certs")
	}
	if len(archive.File) > 1 {
		return nil, fmt.Errorf("unhandled case, too many files in ZIP - cannot match with VASP")
	}

	// Open and parse the certificate file
	var contents io.ReadCloser
	if contents, err = archive.File[0].Open(); err != nil {
		return nil, fmt.Errorf("could not read certificate from archive: %s", err)
	}
	defer contents.Close()

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, contents); err != nil {
		return nil, fmt.Errorf("could not read certificate from archive: %s", err)
	}

	// TODO: do we extract the entire trust chain or just the certificate?
	var cert *x509.Certificate
	if _, cert, _, err = pkcs12.DecodeChain(buf.Bytes(), pkcs12password); err != nil {
		return nil, fmt.Errorf("could not decode pkcs12 chain: %s", err)
	}

	pub = &pb.Certificate{
		Version:            int64(cert.Version),
		SerialNumber:       cert.SerialNumber.Bytes(),
		Signature:          cert.Signature,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		Subject: &pb.Name{
			CommonName:         cert.Subject.CommonName,
			SerialNumber:       cert.Subject.SerialNumber,
			Organization:       cert.Subject.Organization,
			OrganizationalUnit: cert.Subject.OrganizationalUnit,
			StreetAddress:      cert.Subject.StreetAddress,
			Locality:           cert.Subject.Locality,
			Province:           cert.Subject.Province,
			PostalCode:         cert.Subject.PostalCode,
			Country:            cert.Subject.Country,
		},
		Issuer: &pb.Name{
			CommonName:         cert.Issuer.CommonName,
			SerialNumber:       cert.Issuer.SerialNumber,
			Organization:       cert.Issuer.Organization,
			OrganizationalUnit: cert.Issuer.OrganizationalUnit,
			StreetAddress:      cert.Issuer.StreetAddress,
			Locality:           cert.Issuer.Locality,
			Province:           cert.Issuer.Province,
			PostalCode:         cert.Issuer.PostalCode,
			Country:            cert.Issuer.Country,
		},
		NotBefore: cert.NotBefore.Format(time.RFC3339),
		NotAfter:  cert.NotAfter.Format(time.RFC3339),
		Revoked:   false,
	}

	buf = bytes.NewBuffer(nil)
	if err = pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
		return nil, fmt.Errorf("could not PEM encode certificate: %s", err)
	}
	pub.Data = buf.Bytes()

	return pub, nil
}

// get the configured cert storage directory or return a temporary directory/
func (s *Server) getCertStorage() (path string, err error) {
	if s.conf.CertManStorage != "" {
		var stat os.FileInfo
		if stat, err = os.Stat(s.conf.CertManStorage); err != nil {
			if os.IsNotExist(err) {
				// Create the directory if it does not exist and return
				if err = os.MkdirAll(path, 0755); err != nil {
					return "", err
				}
				return s.conf.CertManStorage, nil
			}

			// Other permissions error, cannot access cert storage
			return "", err
		}

		if !stat.IsDir() {
			return "", errors.New("not a directory")
		}
		return s.conf.CertManStorage, nil
	}

	// Create a temporary directory
	if path, err = ioutil.TempDir("", "trisads_certs"); err != nil {
		return "", err
	}
	log.Warn().Str("certs", path).Msg("using a temporary directory for cert downloads")
	return path, err
}
