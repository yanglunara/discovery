package builder

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
	"github.com/yanglunara/discovery/watcher/consul"
	"google.golang.org/grpc/resolver"
)

var (
	_ resolver.Builder = (*Builder)(nil)
	_ register.Stopper = (*Builder)(nil)
)

type Builder struct {
	discoverer register.Discovery
	tiemout    time.Duration
	cancel     context.CancelFunc
	resolver   resolver.Resolver
}

func NewConsulDiscovery(endpoint string) register.Discovery {
	cli, err := api.NewClient(&api.Config{Address: endpoint})
	if err != nil {
		panic(errors.New("init consul error"))
	}
	return consul.NewRegistry(cli)
}

func NewBuilder(b register.Discovery) resolver.Builder {
	return &Builder{
		discoverer: b,
		tiemout:    time.Second * 10,
	}
}

// Build  grpc 驱动
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	watchRes := &struct {
		err error
		w   register.Watcher
	}{}
	done := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	go func() {
		w, err := b.discoverer.Watch(ctx, strings.TrimPrefix(target.URL.Path, "/"))
		watchRes.w = w
		watchRes.err = err
		close(done)
	}()
	var (
		err error
	)

	select {
	case <-done:
		err = watchRes.err
	case <-time.After(b.tiemout):
		err = errors.New("discovery create watcher overtime")
	}
	if err != nil {
		// 关闭所有资源
		cancel()
		_ = b.discoverer.Close()
		return nil, err
	}

	r := &discoveryResolver{
		w:      watchRes.w,
		d:      b.discoverer,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
	}
	b.resolver = r
	go r.watch()

	return r, nil
}

func (b *Builder) Scheme() string {
	return "discovery"
}

func (b *Builder) Close() error {
	b.cancel()
	b.resolver.Close()
	return nil
}
