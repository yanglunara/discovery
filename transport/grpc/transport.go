package grpc

import (
	"github.com/yanglunara/discovery/transport"
	"google.golang.org/grpc/metadata"
)

var (
	_ transport.Transport = (*Transport)(nil)
)

type Transport struct {
	endpoint   string
	operation  string
	reqHeader  headerMetadata
	respHeader headerMetadata
}

func (tr *Transport) Scheme() transport.Scheme {
	return transport.SchemeGRPC
}

func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

func (tr *Transport) Operation() string {
	return tr.operation
}

func (tr *Transport) RequestHeader() transport.Header {
	return tr.reqHeader
}
func (tr *Transport) ResponseHeader() transport.Header {
	return tr.respHeader
}

type headerMetadata metadata.MD

func (hm headerMetadata) Get(key string) string {
	if val := metadata.MD(hm).Get(key); len(val) > 0 {
		return val[0]
	}
	return ""
}

func (hm headerMetadata) Set(key, value string) {
	metadata.MD(hm).Set(key, value)
}

func (hm headerMetadata) Add(key, value string) {
	metadata.MD(hm).Append(key, value)
}

func (hm headerMetadata) Keys() []string {
	keys := make([]string, 0, len(hm))
	for k := range hm {
		keys = append(keys, k)
	}
	return keys
}

func (hm headerMetadata) Values(key string) []string {
	return metadata.MD(hm).Get(key)
}
