package discover

import (
	"context"

	"github.com/yanglunara/discovery/register"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
)

var (
	_ register.Discovery = (*Discovery)(nil)
)

type Discovery struct {
	ctx      context.Context
	client   *clientv3.Client
	kv       clientv3.KV
	watcher  register.Watcher
	registry map[string][]resolver.State
}

func NewDiscovery(ctx context.Context, name string, client *clientv3.Client) *Discovery {
	return &Discovery{
		ctx:     ctx,
		client:  client,
		watcher: local.NewWatcher(ctx, name, client),
	}
}

// GetService 获取服务实例
func (d *Discovery) GetService(ctx context.Context, serviceName string) ([]*register.ServiceInstance, error) {
	return nil, nil
}

// Watch 监听服务变化
func (d *Discovery) Watch(ctx context.Context, serviceName string) (register.Watcher, error) {
	return d.watcher, nil
}

func (d *Discovery) Fetch() (*register.ServiceInstance, bool) {
	return nil, false
}

func (d *Discovery) Close() error {
	_ = d.client.Close()
	_ = d.watcher.Close()
	return nil
}
