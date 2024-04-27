package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yanglunara/discovery/builder"
	"github.com/yanglunara/discovery/register"
	"google.golang.org/grpc"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
)

type (
	Conn interface {
		GetConn(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	}
)
type ClientOption func(o *rpcClient)

func WithDiscoveryAddress(addres string) ClientOption {
	return func(o *rpcClient) {
		o.address = addres
	}
}
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *rpcClient) {
		o.grpcOpts = opts
	}
}
func WithEendpoint(endpoint string) ClientOption {
	return func(o *rpcClient) {
		o.endpoint = fmt.Sprintf("discovery:///%s", endpoint)
	}
}

func WithInsecure(insecure bool) ClientOption {
	return func(o *rpcClient) {
		o.insecure = true
	}
}

var (
	mu        sync.Mutex
	RpcClient Conn
	_         Conn = (*rpcClient)(nil)
)

type rpcClient struct {
	timeout                time.Duration
	balancerName           string
	subsetSize             int
	printDiscoveryDebugLog bool
	healthCheckConfig      string
	discovery              register.Discovery // 服务发现
	address                string
	grpcOpts               []grpc.DialOption
	endpoint               string
	WindowSize             int32
	aliveTime              time.Duration
	insecure               bool
	localCache             map[string]*grpc.ClientConn
}

// SetRPCClient 设置RPC客户端 单列模式
func SetRPCClient(opt ...ClientOption) Conn {
	if RpcClient == nil {
		mu.Lock()
		defer mu.Unlock()
		if RpcClient == nil {
			gcs := rpcClient{
				timeout:                3 * time.Second,
				balancerName:           "round_robin",
				subsetSize:             25,
				printDiscoveryDebugLog: true,
				healthCheckConfig:      `,"healthCheckConfig":{"serviceName":""}`,
				WindowSize:             1 << 24,
				aliveTime:              10 * time.Second,
				localCache:             make(map[string]*grpc.ClientConn),
			}
			for _, o := range opt {
				o(&gcs)
			}
			gcs.discovery = builder.NewConsulDiscovery(gcs.address)
		}
	}
	return RpcClient
}

func (g *rpcClient) setGrpcOpts() []grpc.DialOption {
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]%s}`,
			g.balancerName, g.healthCheckConfig)),
		grpc.WithInitialWindowSize(g.WindowSize),
		grpc.WithInitialConnWindowSize(g.WindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(g.WindowSize))),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(int(g.WindowSize))),
	}
	if g.insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	}
	if g.discovery != nil {
		grpcOpts = append(grpcOpts, grpc.WithResolvers(
			builder.NewBuilder(g.discovery),
		))
	}
	if len(g.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, g.grpcOpts...)
	}
	return grpcOpts
}

func (g *rpcClient) GetConn(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, fmt.Sprintf("discovery:///%s", serviceName), g.setGrpcOpts()...)
}
