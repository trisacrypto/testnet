package trisads

import (
	"context"
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
	vasp := pb.VASP{
		Entity:           in.Entity,
		Contacts:         in.Contacts,
		TrisaEndpoint:    in.TrisaEndpoint,
		CommonName:       in.CommonName,
		Website:          in.Website,
		BusinessCategory: in.BusinessCategory,
		VaspCategory:     in.VaspCategory,
		EstablishedOn:    in.EstablishedOn,
		Trixo:            in.Trixo,
	}

	// TODO: validate partial VASP record

	if out.Id, err = s.db.Create(vasp); err != nil {
		log.Warn().Err(err).Msg("could not register VASP")
		out.Error = &pb.Error{
			Code:    400,
			Message: err.Error(),
		}
	} else {
		name, _ := vasp.Name()
		log.Info().Str("name", name).Str("id", vasp.Id).Msg("registered VASP")
	}

	// TODO: if verify is true: send verification request
	if err = s.SendVerificationEmail(vasp); err != nil {
		log.Error().Err(err).Msg("could not send verification email")
		out.Error = &pb.Error{
			Code:    500,
			Message: err.Error(),
		}
	} else {
		log.Info().Msg("verification email sent")
	}

	// TODO: populate register reply fields with status
	return out, nil
}

// Lookup a VASP entity by name or ID to get full details including the TRISA certification
// if it exists and the entity has been verified.
func (s *Server) Lookup(ctx context.Context, in *pb.LookupRequest) (out *pb.LookupReply, err error) {
	var vasp pb.VASP
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
		var vasps []pb.VASP
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

	var vasps []pb.VASP
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
	var vasp pb.VASP
	out = &pb.StatusReply{}

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
		var vasps []pb.VASP
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
