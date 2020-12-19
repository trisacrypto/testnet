package trisads

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/trisacrypto/testnet/pkg/sectigo"
	"github.com/trisacrypto/testnet/pkg/trisads/pb"
	"github.com/trisacrypto/testnet/pkg/trisads/store"
	"google.golang.org/grpc"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// New creates a TRISA Directory Service with the specified configuration and prepares
// it to listen for and serve GRPC requests.
func New(conf *Settings) (s *Server, err error) {
	// Load the default configuration from the environment
	if conf == nil {
		if conf, err = Config(); err != nil {
			return nil, err
		}
	}

	// Set the global level
	zerolog.SetGlobalLevel(zerolog.Level(conf.LogLevel))

	// Create the server and open the connection to the database
	s = &Server{conf: conf}
	if s.db, err = store.Open(conf.DatabaseDSN); err != nil {
		return nil, err
	}

	// Create the Sectigo API client
	if s.certs, err = sectigo.New(conf.SectigoUsername, conf.SectigoPassword); err != nil {
		return nil, err
	}

	// Create the SendGrid API client
	s.email = sendgrid.NewSendClient(conf.SendGridAPIKey)

	// Configuration complete!
	return s, nil
}

// Server implements the GRPC TRISADirectoryService.
type Server struct {
	pb.UnimplementedTRISADirectoryServer
	db    store.Store
	srv   *grpc.Server
	conf  *Settings
	certs *sectigo.Sectigo
	email *sendgrid.Client
}

// Serve GRPC requests on the specified address.
func (s *Server) Serve() (err error) {
	// Initialize the gRPC server
	s.srv = grpc.NewServer()
	pb.RegisterTRISADirectoryServer(s.srv, s)

	// Catch OS signals for graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		s.Shutdown()
	}()

	// Listen for TCP requests on the specified address and port
	var sock net.Listener
	if sock, err = net.Listen("tcp", s.conf.BindAddr); err != nil {
		return fmt.Errorf("could not listen on %q", s.conf.BindAddr)
	}
	defer sock.Close()

	// Run the server
	log.Info().
		Str("listen", s.conf.BindAddr).
		Str("version", Version()).
		Msg("server started")
	return s.srv.Serve(sock)
}

// Shutdown the TRISA Directory Service gracefully
func (s *Server) Shutdown() (err error) {
	log.Info().Msg("gracefully shutting down")
	s.srv.GracefulStop()
	if err = s.db.Close(); err != nil {
		log.Error().Err(err)
		return err
	}
	return nil
}

// Register a new VASP entity with the directory service. After registration, the new
// entity must go through the verification process to get issued a certificate. The
// status of verification can be obtained by using the lookup RPC call.
func (s *Server) Register(ctx context.Context, in *pb.RegisterRequest) (out *pb.RegisterReply, err error) {
	out = &pb.RegisterReply{}
	vasp := &pb.VASP{
		RegisteredDirectory: s.conf.DirectoryID,
		Entity:              in.Entity,
		Contacts:            in.Contacts,
		TrisaEndpoint:       in.TrisaEndpoint,
		CommonName:          in.CommonName,
		Website:             in.Website,
		BusinessCategory:    in.BusinessCategory,
		VaspCategory:        in.VaspCategory,
		EstablishedOn:       in.EstablishedOn,
		Trixo:               in.Trixo,
		VerificationStatus:  pb.VerificationState_SUBMITTED,
		Version:             1,
	}

	// Compute the common name from the trisa endpoint if not specified
	if vasp.CommonName == "" && vasp.TrisaEndpoint != "" {
		if vasp.CommonName, _, err = net.SplitHostPort(in.TrisaEndpoint); err != nil {
			log.Warn().Err(err).Msg("could not parse common name from endpoint")
			out.Error = &pb.Error{
				Code:    400,
				Message: err.Error(),
			}
			return out, nil
		}
	}

	// Validate partial VASP record to ensure that it can be registered.
	if err = vasp.Validate(true); err != nil {
		log.Warn().Err(err).Msg("invalid or incomplete VASP registration")
		out.Error = &pb.Error{
			Code:    400,
			Message: err.Error(),
		}
		return out, nil
	}

	// TODO: create legal entity hash to detect a repeat registration without ID
	// TODO: add signature to leveldb indices
	if out.Id, err = s.db.Create(vasp); err != nil {
		log.Warn().Err(err).Msg("could not register VASP")
		out.Error = &pb.Error{
			Code:    400,
			Message: err.Error(),
		}
		return out, nil
	}

	name, _ := vasp.Name()
	log.Info().Str("name", name).Str("id", vasp.Id).Msg("registered VASP")

	// Begin verification process by sending emails to all contacts in the VASP record.
	// TODO: add to processing queue to return sooner/parallelize work
	if err = s.VerifyContactEmail(vasp); err != nil {
		log.Error().Err(err).Msg("could not verify contacts")
		out.Error = &pb.Error{
			Code:    500,
			Message: err.Error(),
		}
		return out, nil
	}
	log.Info().Msg("contact email verifications sent")

	out.Id = vasp.Id
	out.RegisteredDirectory = vasp.RegisteredDirectory
	out.CommonName = vasp.CommonName
	out.Status = vasp.VerificationStatus
	out.Message = "verification code sent to contact emails, please check spam folder if not arrived"

	return out, nil
}

// Lookup a VASP entity by name or ID to get full details including the TRISA certification
// if it exists and the entity has been verified.
func (s *Server) Lookup(ctx context.Context, in *pb.LookupRequest) (out *pb.LookupReply, err error) {
	var vasp *pb.VASP
	out = &pb.LookupReply{}

	if in.Id != "" {
		// TODO: add registered directory to lookup
		if vasp, err = s.db.Retrieve(in.Id); err != nil {
			out.Error = &pb.Error{
				Code:    404,
				Message: err.Error(),
			}
		}

	} else if in.CommonName != "" {
		// TODO: change lookup to unique common name lookup
		var vasps []*pb.VASP
		if vasps, err = s.db.Search(map[string]interface{}{"common_name": in.CommonName}); err != nil {
			out.Error = &pb.Error{
				Code:    404,
				Message: err.Error(),
			}
		}

		if len(vasps) == 1 {
			vasp = vasps[0]
		} else {
			out.Error = &pb.Error{
				Code:    404,
				Message: "not found",
			}
		}
	} else {
		out.Error = &pb.Error{
			Code:    400,
			Message: "no lookup query provided",
		}
		return out, nil
	}

	if out.Error == nil {
		// TODO: should lookups only return verified peers?
		out.Id = vasp.Id
		out.RegisteredDirectory = vasp.RegisteredDirectory
		out.CommonName = vasp.CommonName
		out.Endpoint = vasp.TrisaEndpoint
		out.Certificate = vasp.Certificate
		out.Name, _ = vasp.Name()
		out.Country = vasp.Entity.CountryOfRegistration
		out.VerifiedOn = vasp.VerifiedOn
		log.Info().Str("id", vasp.Id).Msg("VASP lookup succeeded")
	} else {
		log.Warn().Err(out.Error).Msg("could not lookup VASP")
	}
	return out, nil
}

// Search for VASP entity records by name or by country in order to perform more detailed
// Lookup requests. The search process is purposefully simplistic at the moment.
func (s *Server) Search(ctx context.Context, in *pb.SearchRequest) (out *pb.SearchReply, err error) {
	out = &pb.SearchReply{}
	query := make(map[string]interface{})
	query["name"] = in.Name
	query["country"] = in.Country

	var vasps []*pb.VASP
	if vasps, err = s.db.Search(query); err != nil {
		out.Error = &pb.Error{
			Code:    400,
			Message: err.Error(),
		}
	}

	out.Results = make([]*pb.SearchResult, 0, len(vasps))
	for _, vasp := range vasps {
		out.Results = append(out.Results, &pb.SearchResult{
			Id:                  vasp.Id,
			RegisteredDirectory: vasp.RegisteredDirectory,
			CommonName:          vasp.CommonName,
		})
	}

	entry := log.With().
		Strs("name", in.Name).
		Strs("country", in.Country).
		Int("results", len(out.Results)).
		Logger()

	if out.Error != nil {
		entry.Warn().Err(out.Error).Msg("unsuccessful search")
	} else {
		entry.Info().Msg("search succeeded")
	}
	return out, nil
}

// Status returns the status of a VASP including its verification and service status if
// the directory service is performing health check monitoring.
func (s *Server) Status(ctx context.Context, in *pb.StatusRequest) (out *pb.StatusReply, err error) {
	var vasp *pb.VASP
	out = &pb.StatusReply{}

	if in.Id != "" {
		// TODO: add registered directory to lookup
		if vasp, err = s.db.Retrieve(in.Id); err != nil {
			log.Error().Err(err).Str("id", in.Id).Msg("could not retrieve vasp")
			out.Error = &pb.Error{
				Code:    404,
				Message: err.Error(),
			}
		}

	} else if in.CommonName != "" {
		// TODO: change lookup to unique common name lookup
		var vasps []*pb.VASP
		if vasps, err = s.db.Search(map[string]interface{}{"common_name": in.CommonName}); err != nil {
			log.Error().Err(err).Str("common_name", in.CommonName).Msg("could not retrieve vasp")
			out.Error = &pb.Error{
				Code:    404,
				Message: err.Error(),
			}
		}

		if len(vasps) == 1 {
			vasp = vasps[0]
		} else {
			out.Error = &pb.Error{
				Code:    404,
				Message: "not found",
			}
		}
	} else {
		out.Error = &pb.Error{
			Code:    400,
			Message: "no lookup query provided",
		}
		return out, nil
	}

	if out.Error == nil {
		// TODO: should lookups only return verified peers?
		out.VerificationStatus = vasp.VerificationStatus
		out.ServiceStatus = vasp.ServiceStatus
		out.VerifiedOn = vasp.VerifiedOn
		out.FirstListed = vasp.FirstListed
		out.LastUpdated = vasp.LastUpdated
		log.Info().Str("id", vasp.Id).Msg("VASP status succeeded")
	} else {
		log.Warn().Err(out.Error).Msg("could not lookup VASP for status")
	}
	return out, nil
}

// VerifyEmail checks the contact tokens for the specified VASP and registers the
// contact email verification. If successful, this method then sends the verification
// request to the TRISA Admins for review and generates a PKCS12 password in the RPC
// response to decrypt the certificate private keys when they're emailed.
func (s *Server) VerifyEmail(ctx context.Context, in *pb.VerifyEmailRequest) (out *pb.VerifyEmailReply, err error) {
	out = &pb.VerifyEmailReply{}

	var vasp *pb.VASP
	if vasp, err = s.db.Retrieve(in.Id); err != nil {
		log.Error().Err(err).Str("id", in.Id).Msg("could not retrieve vasp")
		out.Error = &pb.Error{
			Code:    404,
			Message: err.Error(),
		}
		return out, nil
	}

	verified := 0
	found := false
	contacts := []*pb.Contact{
		vasp.Contacts.Technical,
		vasp.Contacts.Administrative,
		vasp.Contacts.Billing,
		vasp.Contacts.Legal,
	}
	for _, contact := range contacts {
		if contact == nil {
			continue
		}
		if contact.Token == in.Token {
			log.Info().Str("email", contact.Email).Msg("contact email verified")
			contact.Verified = true
			contact.Token = ""
			found = true
		}
		if contact.Verified {
			verified++
		}
	}

	if !found || verified == 0 {
		log.Error().Err(err).Str("token", in.Token).Msg("could not find contact with token")
		out.Error = &pb.Error{
			Code:    404,
			Message: "could not find contact with specified token",
		}
		return out, nil
	}

	// Ensures that we only send the verification email to the admins once.
	if verified > 1 {
		// Save the updated contact
		if err = s.db.Update(vasp); err != nil {
			log.Error().Err(err).Msg("could not update vasp after contact verification")
			out.Error = &pb.Error{
				Code:    500,
				Message: err.Error(),
			}
			return out, nil
		}

		out.Status = vasp.VerificationStatus
		out.Message = "email successfully verified; verification review already sent to TRISA admins"
		return out, nil
	}

	// Note that this status will get updated in the review request email
	vasp.VerificationStatus = pb.VerificationState_EMAIL_VERIFIED

	// If this is the first verification, generate the PKCS12 password and send verification review email
	// TODO: make this better
	if err = s.ReviewRequestEmail(vasp); err != nil {
		log.Error().Err(err).Msg("could not send verification review email")
		out.Error = &pb.Error{
			Code:    500,
			Message: "could not send verification review email",
		}
		return out, nil
	}

	// Now that the email has been sent out the vasp is pending review
	vasp.VerificationStatus = pb.VerificationState_PENDING_REVIEW

	// Create and encrypt PKCS12 password
	password := CreateToken(16)
	vasp.CertificateRequest = &pb.CertificateRequest{}

	if vasp.CertificateRequest.Pkcs12Password, vasp.CertificateRequest.Pkcs12Signature, err = s.Encrypt(password); err != nil {
		log.Error().Err(err).Msg("could not encrypt password to store in database")
		out.Error = &pb.Error{
			Code:    500,
			Message: "could not create certificate request password",
		}
	}

	// Save the VASP and newly created certificate request
	if err = s.db.Update(vasp); err != nil {
		log.Error().Err(err).Msg("could not update vasp after contact verification and pkcs12 password generation")
		out.Error = &pb.Error{
			Code:    500,
			Message: err.Error(),
		}
		return out, nil
	}

	// Perform a check
	check, err := s.db.Retrieve(vasp.Id)
	if err != nil {
		log.Error().Err(err).Msg("could not retrieve VASP to check it")
	}
	data, _ := json.MarshalIndent(check, "", "  ")
	fmt.Println(string(data))

	out.Status = vasp.VerificationStatus
	out.Message = "email successfully verified and verification review sent to TRISA admins; pkcs12 password to decrypt your emailed certificates attached - do not lose!"
	out.Pkcs12Password = password
	return out, nil
}
