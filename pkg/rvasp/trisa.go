package rvasp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/testnet/pkg/trisa/crypto/aesgcm"
	"github.com/trisacrypto/testnet/pkg/trisa/crypto/rsaoeap"
	"github.com/trisacrypto/testnet/pkg/trisa/mtls"
	protocol "github.com/trisacrypto/testnet/pkg/trisa/protocol/v1alpha1"
	"github.com/trisacrypto/testnet/pkg/trust"
	"google.golang.org/grpc"
)

// TRISA implements the GRPC TRISANetwork and TRISAHealth services.
type TRISA struct {
	protocol.UnimplementedTRISANetworkServer
	protocol.UnimplementedTRISAHealthServer
	parent *Server
	srv    *grpc.Server
	certs  *trust.Provider
	chain  trust.ProviderPool
	sign   *rsa.PrivateKey
}

// NewTRISA from a parent server.
func NewTRISA(parent *Server) (svc *TRISA, err error) {
	conf := parent.conf
	if conf.CertPath == "" || conf.TrustChainPath == "" {
		return nil, errors.New("certificate path or trust chain path missing")
	}

	var sz *trust.Serializer
	svc = &TRISA{parent: parent}

	// Load the TRISA certificates for server-side TLS
	if sz, err = trust.NewSerializer(false); err != nil {
		return nil, err
	}

	if svc.certs, err = sz.ReadFile(conf.CertPath); err != nil {
		return nil, err
	}

	// Load the TRISA public pool from disk
	if sz, err = trust.NewSerializer(false); err != nil {
		return nil, err
	}

	if svc.chain, err = sz.ReadPoolFile(conf.TrustChainPath); err != nil {
		return nil, err
	}

	// Extract the signing key from the TRISA certificate
	// TODO: use separate signing key from mTLS certs
	if svc.sign, err = svc.certs.GetRSAKeys(); err != nil {
		return nil, err
	}
	return svc, nil
}

// Serve initializes the GRPC server and returns any errors during intitialization, it
// then kicks off a go routine to handle requests. Not thread safe, should not be called
// multiple times.
func (s *TRISA) Serve() (err error) {
	// Create TLS Credentials for the server
	// NOTE: the mtls package specifies TRISA-specific TLS configuration.
	var creds grpc.ServerOption
	if creds, err = mtls.ServerCreds(s.certs, s.chain); err != nil {
		return err
	}

	s.srv = grpc.NewServer(creds)
	protocol.RegisterTRISANetworkServer(s.srv, s)
	protocol.RegisterTRISAHealthServer(s.srv, s)

	var sock net.Listener
	if sock, err = net.Listen("tcp", s.parent.conf.TRISABindAddr); err != nil {
		return fmt.Errorf("trisa service could not listen on %q", s.parent.conf.TRISABindAddr)
	}

	go func() {
		defer sock.Close()

		log.Info().
			Str("listen", s.parent.conf.TRISABindAddr).
			Msg("trisa server started")

		if err := s.srv.Serve(sock); err != nil {
			s.parent.echan <- err
		}
	}()

	return nil
}

// Shutdown the TRISA server gracefully
func (s *TRISA) Shutdown() (err error) {
	log.Info().Msg("trisa server gracefully shutting down")
	s.srv.GracefulStop()
	log.Debug().Msg("successful trisa server shutdown")
	return nil
}

// Transfer enables a quick one-off transaction between peers.
func (s *TRISA) Transfer(ctx context.Context, in *protocol.SecureEnvelope) (out *protocol.SecureEnvelope, err error) {
	var peer *Peer
	if peer, err = s.parent.peerFromContext(ctx); err != nil {
		log.Error().Err(err).Msg("could not verify peer from context")
		return nil, &protocol.Error{
			Code:    protocol.Unverified,
			Message: err.Error(),
		}
	}
	log.Info().Str("peer", peer.CommonName).Msg("unary transfer request received")

	// Check the algorithms to make sure they're supported
	if in.EncryptionAlgorithm != "AES256-GCM" || in.HmacAlgorithm != "HMAC-SHA256" {
		log.Warn().
			Str("encryption", in.EncryptionAlgorithm).
			Str("hmac", in.HmacAlgorithm).
			Msg("unsupported cryptographic algorithms")
		return nil, protocol.Errorf(protocol.UnhandledAlgorithm, "server only supports AES256-GCM and HMAC-SHA256")
	}

	// Decrypt the encryption key and HMAC secret with private signing keys
	var cipher *rsaoeap.RSA
	if cipher, err = rsaoeap.New(&s.sign.PublicKey, s.sign); err != nil {
		log.Error().Err(err).Msg("could not create RSA cipher for asymmetric decryption")
		return nil, protocol.Errorf(protocol.InternalError, "unable to decrypt keys")
	}

	var encryptionKey, hmacSecret []byte
	if encryptionKey, err = cipher.Decrypt(in.EncryptionKey); err != nil {
		log.Error().Err(err).Msg("could not decrypt encryption key")
		return nil, &protocol.Error{
			Code:    protocol.InvalidKey,
			Message: "encryption key signed incorrectly",
			Retry:   true,
		}
	}
	if hmacSecret, err = cipher.Decrypt(in.HmacSecret); err != nil {
		log.Error().Err(err).Msg("could not decrypt hmac secret")
		return nil, &protocol.Error{
			Code:    protocol.InvalidKey,
			Message: "hmac secret signed incorrectly",
			Retry:   true,
		}
	}

	// Decrypt the message and verify its signature
	var payloadData []byte
	var payloadCipher *aesgcm.AESGCM
	if payloadCipher, err = aesgcm.New(encryptionKey, hmacSecret); err != nil {
		log.Error().Err(err).Msg("could not create AES-GCM cipher for symmetric decryption")
		return nil, protocol.Errorf(protocol.InternalError, "unable to decrypt payload")
	}

	if err = payloadCipher.Verify(in.Payload, in.Hmac); err != nil {
		log.Error().Err(err).Msg("could not verify hmac signature")
		return nil, protocol.Errorf(protocol.InvalidSignature, "could not verify HMAC signature")
	}

	if payloadData, err = payloadCipher.Decrypt(in.Payload); err != nil {
		log.Error().Err(err).Msg("could not decrypt payload")
		return nil, protocol.Errorf(protocol.InvalidKey, "could not decrypt payload with key")
	}

	// Parse the payload into rVASP-appropriate data
	payload := &protocol.Payload{}
	if err = proto.Unmarshal(payloadData, payload); err != nil {
		log.Error().Err(err).Msg("could not unmarshal payload")
		return nil, protocol.Errorf(protocol.EnvelopeDecodeFail, "could not unmarshal payload")
	}

	if payload.Identity.TypeUrl != "type.googleapis.com/ivms101.IdentityPayload" {
		log.Warn().Str("type", payload.Identity.TypeUrl).Msg("unsupported identity type")
		return nil, protocol.Errorf(protocol.UnparseableIdentity, "rVASP requires ivms101.IdentityPayload payload identity type")
	}

	if payload.Transaction.TypeUrl != "type.googleapis.com/rvasp.v1.Transaction" {
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unsupported transaction type")
		return nil, protocol.Errorf(protocol.UnparseableTransaction, "rVASP requires rvasp.v1.Transaction payload transaction type")
	}

	// Check signing key is available to send an encrypted response
	if peer.SigningKey == nil {
		log.Warn().Str("peer", peer.CommonName).Msg("no remote signing key available")
		return nil, &protocol.Error{
			Code:    protocol.NoSigningKey,
			Message: "please retry transfer after key exchange",
			Retry:   true,
		}
	}

	return nil, &protocol.Error{
		Code:    protocol.Unimplemented,
		Message: "rVASP has not implemented transfers yet",
		Retry:   false,
	}
}

// TransferStream allows for high-throughput transactions.
func (s *TRISA) TransferStream(stream protocol.TRISANetwork_TransferStreamServer) (err error) {
	return &protocol.Error{
		Code:    protocol.Unimplemented,
		Message: "rVASP has not implemented the transfer stream yet",
		Retry:   false,
	}
}

// ConfirmAddress allows the rVASP to respond to proof-of-control requests.
func (s *TRISA) ConfirmAddress(ctx context.Context, in *protocol.Address) (out *protocol.AddressConfirmation, err error) {
	return nil, &protocol.Error{
		Code:    protocol.Unimplemented,
		Message: "rVASP has not implemented address confirmation yet",
		Retry:   false,
	}
}

// KeyExchange facilitates signing key exchange between VASPs.
func (s *TRISA) KeyExchange(ctx context.Context, in *protocol.SigningKey) (out *protocol.SigningKey, err error) {
	var peer *Peer
	if peer, err = s.parent.peerFromContext(ctx); err != nil {
		log.Error().Err(err).Msg("could not verify peer from context")
		return nil, &protocol.Error{
			Code:    protocol.Unverified,
			Message: err.Error(),
		}
	}
	log.Info().Str("peer", peer.CommonName).Msg("key exchange request received")

	// Cache key inside of the in-memory Peer map
	var pub interface{}
	if pub, err = x509.ParsePKIXPublicKey(in.Data); err != nil {
		log.Error().
			Err(err).
			Int64("version", in.Version).
			Str("algorithm", in.PublicKeyAlgorithm).
			Msg("could not parse incoming PKIX public key")
		return nil, protocol.Errorf(protocol.NoSigningKey, "could not parse signing key")
	}

	var ok bool
	if peer.SigningKey, ok = pub.(*rsa.PublicKey); !ok {
		return nil, protocol.Errorf(protocol.UnhandledAlgorithm, "unsuported signing algorithm")
	}

	// TODO: check not before and not after constraints

	// TODO: Kick off a go routine to store the key in the database

	// Return the public signing-key of the service
	// TODO: use separate signing key insead of using public key of mTLS certs
	var key *x509.Certificate
	if key, err = s.certs.GetLeafCertificate(); err != nil {
		log.Error().Err(err).Msg("could not extract leaf certificate")
		return nil, protocol.Errorf(protocol.InternalError, "could not return signing keys")
	}

	out = &protocol.SigningKey{
		Version:            int64(key.Version),
		Signature:          key.Signature,
		SignatureAlgorithm: key.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: key.PublicKeyAlgorithm.String(),
		NotBefore:          key.NotBefore.Format(time.RFC3339),
		NotAfter:           key.NotAfter.Format(time.RFC3339),
	}

	if out.Data, err = x509.MarshalPKIXPublicKey(key.PublicKey); err != nil {
		log.Error().Err(err).Msg("could not marshal PKIX public key")
		return nil, protocol.Errorf(protocol.InternalError, "could not marshal public key")
	}

	return out, nil
}

// Status returns a directory health check status as online and requests half an hour checks.
func (s *TRISA) Status(ctx context.Context, in *protocol.HealthCheck) (out *protocol.ServiceState, err error) {
	log.Info().
		Uint32("attempts", in.Attempts).
		Str("last checked at", in.LastCheckedAt).
		Msg("status check")

	// Request another health check between 30 minutes and an hour from now.
	now := time.Now()
	return &protocol.ServiceState{
		Status:    protocol.ServiceState_HEALTHY,
		NotBefore: now.Add(30 * time.Minute).Format(time.RFC3339),
		NotAfter:  now.Add(1 * time.Hour).Format(time.RFC3339),
	}, nil
}
