package builder

import (
	"context"
	"github.com/yanglunara/discovery/discover"
	"net/url"
	"sync"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

type mockConn struct{}

func (m *mockConn) UpdateState(resolver.State) error {
	return nil
}

func (m *mockConn) ReportError(error) {}

func (m *mockConn) NewAddress(_ []resolver.Address) {}

func (m *mockConn) NewServiceConfig(_ string) {}

func (m *mockConn) ParseServiceConfig(_ string) *serviceconfig.ParseResult {
	return nil
}

func TestBuilder_Build(t *testing.T) {
	ctx := context.Background()
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second, DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	re := discover.NewDiscovery(ctx, "logic", client)
	b := NewBuilder(re)
	_, err = b.Build(
		resolver.Target{
			URL: url.URL{
				Scheme: resolver.GetDefaultScheme(),
				Path:   "grpc://authority/endpoint",
			},
		},
		&mockConn{},
		resolver.BuildOptions{},
	)
	time.Sleep(100 * time.Second)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return
	}
	sync := sync.WaitGroup{}

	sync.Wait()
	//timeoutBuilder := NewBuilder(&mockDiscovery{}, WithTimeout(0))
	//_, err = timeoutBuilder.Build(
	//	resolver.Target{
	//		URL: url.URL{
	//			Scheme: resolver.GetDefaultScheme(),
	//			Path:   "grpc://authority/endpoint",
	//		},
	//	},
	//	&mockConn{},
	//	resolver.BuildOptions{},
	//)
	//if err == nil {
	//	t.Errorf("expected error, got %v", err)
	//}
}
