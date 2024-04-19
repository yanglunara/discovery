package grpc

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/yanglunara/discovery/transport"
	ct "github.com/yanglunara/discovery/transport/context"
	"github.com/yanglunara/discovery/transport/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpcmd "google.golang.org/grpc/metadata"
)

var (
	_ transport.GrpcService = (*Service)(nil)
	_ transport.EndPointer  = (*Service)(nil)
)

type ServiceOption func(s *Service)

type Service struct {
	*grpc.Server
	baseCtx      context.Context
	lis          net.Listener
	middleware   transport.Matcher
	grpcOpts     []grpc.ServerOption
	health       *health.Server
	isOpenHealth bool
	endpoint     *url.URL
	timeout      time.Duration
	network      string
	address      string
	unarys       []grpc.UnaryServerInterceptor
	streams      []grpc.StreamServerInterceptor
	adminClean   func()
	err          error
}

func (s *Service) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := ct.NewContext(ctx, s.baseCtx)
		defer cancel()
		md, _ := grpcmd.FromIncomingContext(ctx)
		respHeader := grpcmd.MD{}
		tr := &Transport{
			operation:  info.FullMethod,
			reqHeader:  headerMetadata(md),
			respHeader: headerMetadata(respHeader),
		}
		if s.endpoint != nil {
			tr.endpoint = s.endpoint.String()
		}
		ctx = transport.NewServiceContext(ctx, tr)
		// 设置超时
		if s.timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, s.timeout)
			defer cancel()
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
		if next := s.middleware.Match(tr.Operation()); len(next) > 0 {
			h = middleware.Next(next...)(h)
		}
		resp, err := h(ctx, req)
		// 发送header 头
		if len(respHeader) > 0 {
			_ = grpc.SendHeader(ctx, respHeader)
		}
		return resp, err
	}
}

type gsStram struct {
	grpc.ServerStream
	ctx context.Context
}

func NewStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	return &gsStram{
		ServerStream: stream,
		ctx:          ctx,
	}
}

func (g *gsStram) Context() context.Context {
	return g.ctx
}

func (s *Service) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, cancel := ct.NewContext(stream.Context(), s.baseCtx)
		defer cancel()
		md, _ := grpcmd.FromIncomingContext(ctx)
		respHeader := grpcmd.MD{}
		ctx = transport.NewServiceContext(ctx, &Transport{
			operation:  info.FullMethod,
			endpoint:   s.endpoint.String(),
			reqHeader:  headerMetadata(md),
			respHeader: headerMetadata(respHeader),
		})
		ws := NewStream(ctx, stream)
		err := handler(srv, ws)
		if len(respHeader) > 0 {
			_ = grpc.SendHeader(ctx, respHeader)
		}
		return err
	}
}
