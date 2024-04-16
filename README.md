# 基于 consul 服务发现与注册

## 已完成
- Consul
- Etcd(半完成 grpc 支持1.26版本 mod 很混乱)

## Docker 环境

```docker
cd ./deployment && docker up -d consul
```

## 示例代码

以下是一个使用Consul作为服务发现工具的示例代码：
```go
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
	_, _ = b.Build(
		resolver.Target{
			URL: url.URL{
				Scheme: resolver.GetDefaultScheme(),
				Path:   "grpc://logic.grpc",
			},
		},
		&mockConn{},
		resolver.BuildOptions{},
	)
}
```

## 服务下线告警