package grpc

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/yanglunara/discovery/lib"
	"github.com/yanglunara/discovery/transport"
	mid "github.com/yanglunara/discovery/transport/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func Middleware(m ...mid.Middleware) ServiceOption {
	return func(s *Service) {
		s.middleware.Use(m...)
	}
}

func UnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServiceOption {
	return func(s *Service) {
		s.unarys = append(s.unarys, in...)
	}
}

func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServiceOption {
	return func(s *Service) {
		s.streams = append(s.streams, in...)
	}
}

func Options(opts ...grpc.ServerOption) ServiceOption {
	return func(s *Service) {
		s.grpcOpts = append(s.grpcOpts, opts...)
	}
}

func NewGrpcServer(opts ...ServiceOption) *Service {
	srv := &Service{
		baseCtx: context.Background(),
		network: "tcp",
		address: ":9090",
		timeout: 1 * time.Second,
		// 开启心跳
		health:     health.NewServer(),
		middleware: transport.NewMatcher(),
	}
	for _, opt := range opts {
		opt(srv)
	}
	unary := []grpc.UnaryServerInterceptor{
		srv.UnaryServerInterceptor(),
	}
	stream := []grpc.StreamServerInterceptor{
		srv.StreamServerInterceptor(),
	}

	if len(unary) > 0 {
		unary = append(unary, srv.unarys...)
	}
	if len(stream) > 0 {
		stream = append(stream, srv.streams...)
	}

	grpcOpts := []grpc.ServerOption{
		grpc.ChainStreamInterceptor(stream...),
		grpc.ChainUnaryInterceptor(unary...),
	}
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}
	if !srv.isOpenHealth {
		grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	}
	srv.Server = grpc.NewServer(grpcOpts...)
	reflection.Register(srv.Server)

	srv.adminClean, _ = admin.Register(srv.Server)
	return srv
}

func (s *Service) Use(selector string, m ...mid.Middleware) {
	s.middleware.Add(selector, m...)
}

func (s *Service) Endpoint() (*url.URL, error) {
	if err := s.listenEndpoint(); err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

func (s *Service) listenEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		addr, err := lib.Extract(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = &url.URL{
			Scheme: "grpc",
			Host:   addr,
		}
	}
	return s.err
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.listenEndpoint(); err != nil {
		return err
	}
	s.baseCtx = ctx
	s.health.Resume()
	return s.Serve(s.lis)
}

func (s *Service) Stop(_ context.Context) error {
	if s.adminClean != nil {
		s.adminClean()
	}
	s.health.Shutdown()
	s.GracefulStop()
	return nil
}
