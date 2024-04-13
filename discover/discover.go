package discover

import (
	"context"
	"github.com/yanglunara/discovery/register"
	"google.golang.org/grpc/resolver"
)

var (
	_ register.Discovery = (*Discovery)(nil)
)

type Discovery struct {
	ctx      context.Context
	watcher  register.Watcher
	registry map[string][]resolver.State
}

func NewDiscovery(ctx context.Context, watcher register.Watcher) *Discovery {
	return &Discovery{
		ctx:     ctx,
		watcher: watcher,
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

func (d *Discovery) Close() error {
	_ = d.watcher.Close()
	return nil
}
