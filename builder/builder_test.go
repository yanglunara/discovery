package builder

import (
	"net/url"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/watcher/consul"
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
	cli, _ := api.NewClient(&api.Config{Address: "host.docker.internal:8500"})

	re := consul.NewRegistry(cli)
	b := NewBuilder(re)
	_, err := b.Build(
		resolver.Target{
			URL: url.URL{
				Scheme: resolver.GetDefaultScheme(),
				Path:   "grpc://authority/endpoint",
			},
		},
		&mockConn{},
		resolver.BuildOptions{},
	)
	// time.Sleep(100 * time.Second)
	// if err != nil {
	// 	t.Errorf("expected no error, got %v", err)
	// 	return
	// }
	// sync := sync.WaitGroup{}

	// sync.Wait()
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
