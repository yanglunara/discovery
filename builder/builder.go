package builder

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yanglunara/discovery/register"
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
