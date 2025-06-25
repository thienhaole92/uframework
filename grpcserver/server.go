package grpcserver

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultMaxRecvMsgSize        = 8 * 1024 * 1024 // 8 MB
	defaultKeepaliveEnforcement  = 5 * time.Second
	defaultMaxConnectionIdle     = 15 * time.Second
	defaultMaxConnectionAge      = 600 * time.Second
	defaultMaxConnectionAgeGrace = 5 * time.Second
	defaultKeepaliveTime         = 5 * time.Second
	defaultKeepaliveTimeout      = 1 * time.Second
)

type Option struct {
	Host                  string
	Port                  int
	MaxRecvMsgSize        int           // Maximum message size the server can receive.
	KeepaliveEnforcement  time.Duration // Minimum time between client pings.
	MaxConnectionIdle     time.Duration // Maximum time a connection can be idle.
	MaxConnectionAge      time.Duration // Maximum lifetime of a connection.
	MaxConnectionAgeGrace time.Duration // Grace period for closing connections.
	KeepaliveTime         time.Duration // Time after which a ping is sent if the connection is idle.
	KeepaliveTimeout      time.Duration // Time to wait for a ping acknowledgment.
}

func (o *Option) setDefaults() {
	if o.MaxRecvMsgSize == 0 {
		o.MaxRecvMsgSize = defaultMaxRecvMsgSize
	}

	if o.KeepaliveEnforcement == 0 {
		o.KeepaliveEnforcement = defaultKeepaliveEnforcement
	}

	if o.MaxConnectionIdle == 0 {
		o.MaxConnectionIdle = defaultMaxConnectionIdle
	}

	if o.MaxConnectionAge == 0 {
		o.MaxConnectionAge = defaultMaxConnectionAge
	}

	if o.MaxConnectionAgeGrace == 0 {
		o.MaxConnectionAgeGrace = defaultMaxConnectionAgeGrace
	}

	if o.KeepaliveTime == 0 {
		o.KeepaliveTime = defaultKeepaliveTime
	}

	if o.KeepaliveTimeout == 0 {
		o.KeepaliveTimeout = defaultKeepaliveTimeout
	}
}

type Server struct {
	Server  *grpc.Server
	address string
}

func New(opts *Option) *Server {
	// Set default values for any missing configuration fields.
	opts.setDefaults()

	// Set up logging interceptors.
	unaryInterceptor, streamInterceptor := setupLogging()

	options := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(opts.MaxRecvMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             opts.KeepaliveEnforcement,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     opts.MaxConnectionIdle,
			MaxConnectionAge:      opts.MaxConnectionAge,
			MaxConnectionAgeGrace: opts.MaxConnectionAgeGrace,
			Time:                  opts.KeepaliveTime,
			Timeout:               opts.KeepaliveTimeout,
		}),
		grpc.UnaryInterceptor(unaryInterceptor),   // Add the unary interceptor.
		grpc.StreamInterceptor(streamInterceptor), // Add the stream interceptor.
	}

	// Create a new gRPC server with the configured options.
	grpcServer := grpc.NewServer(options...)

	return &Server{
		Server:  grpcServer,
		address: net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)),
	}
}

func setupLogging() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	// Unary interceptor for logging unary RPCs.
	unaryInterceptor := func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Log the incoming request.
		log.Info().
			Str("method", info.FullMethod).
			Interface("request", req).
			Msg("Incoming gRPC request")

		// Call the handler to process the request.
		resp, err := handler(ctx, req)

		// Log the response or error.
		if err != nil {
			log.Error().
				Err(err).
				Str("method", info.FullMethod).
				Msg("gRPC request failed")
		} else {
			log.Info().
				Str("method", info.FullMethod).
				Msg("gRPC request completed successfully")
		}

		return resp, err
	}

	// Stream interceptor for logging streaming RPCs.
	streamInterceptor := func(
		srv any,
		sss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Log the incoming stream request.
		log.Info().
			Str("method", info.FullMethod).
			Msg("Incoming gRPC stream request")

		// Call the handler to process the stream.
		err := handler(srv, sss)

		// Log the outcome of the stream.
		if err != nil {
			log.Error().
				Err(err).
				Str("method", info.FullMethod).
				Msg("gRPC stream request failed")
		} else {
			log.Info().
				Str("method", info.FullMethod).
				Msg("gRPC stream request completed successfully")
		}

		return err
	}

	return unaryInterceptor, streamInterceptor
}

func (s *Server) Run() {
	go func() {
		log.Info().Str("address", s.address).Msg("Starting gRPC server")

		// Create a TCP listener on the specified address.
		listener, err := net.Listen("tcp", s.address)
		if err != nil {
			log.Panic().Err(err).Str("address", s.address).Msg("Failed to create gRPC server listener")
		}

		// Start serving incoming connections.
		log.Info().Str("address", s.address).Msg("gRPC server is now listening")

		if err := s.Server.Serve(listener); err != nil {
			log.Panic().Err(err).Str("address", s.address).Msg("Failed to start gRPC server")
		}
	}()
}

func (s *Server) Close() {
	log.Info().Msg("Shutting down gRPC server gracefully")
	s.Server.GracefulStop()
	log.Info().Msg("gRPC server has been stopped")
}
