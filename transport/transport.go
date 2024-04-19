package transport

import (
	"context"
	"net/url"
)

type GrpcService interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type EndPointer interface {
	Endpoint() (*url.URL, error)
}

type Header interface {
	Get(key string) string
	Set(key, value string)
	Add(key, value string)
	Keys() []string
	Values(key string) []string
}

type Transport interface {
	Scheme() Scheme

	Endpoint() string

	Operation() string
	RequestHeader() Header
	ResponseHeader() Header
}
type Scheme string

const ShenmeHTTP Scheme = "http"
const SchemeGRPC Scheme = "grpc"

func (s Scheme) String() string {
	return string(s)
}

type (
	serviceTransportKey struct{}
	clientTransportKey  struct{}
)

func NewServiceContext(ctx context.Context, t Transport) context.Context {
	return context.WithValue(ctx, serviceTransportKey{}, t)
}

func FromServiceContext(ctx context.Context) (Transport, bool) {
	t, ok := ctx.Value(serviceTransportKey{}).(Transport)
	return t, ok
}

func NewClientContext(ctx context.Context, t Transport) context.Context {
	return context.WithValue(ctx, clientTransportKey{}, t)
}

func FromClientContext(ctx context.Context) (Transport, bool) {
	t, ok := ctx.Value(clientTransportKey{}).(Transport)
	return t, ok
}
