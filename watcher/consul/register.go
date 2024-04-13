package consul

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
)

var (
	_ register.Registrar = (*Registry)(nil)
	_ register.Stopper   = (*Registry)(nil)
)

func NewRegistry(client *api.Client) *Registry {
	r := &Registry{
		registry: make(map[string]*service),
		timeout:  10 * time.Second,
		cli: &Client{
			cli:                            client,
			dc:                             register.SingleDataCenter,
			healthCheckInterval:            10 * time.Second,
			heartBeat:                      true,
			deregisterCriticalServiceAfter: 600 * time.Second,
			enableHealthCheck:              true,
			maxTry:                         5,
		},
	}
	// 初始化上下文
	r.cli.ctx, r.cli.cancel = context.WithCancel(context.Background())
	// 初始化 entries
	r.cli.entries = NewEntries(NewResolver(r.cli.ctx), r.cli.cli)
	r.cli.timeout = r.timeout

	return r
}

type Registry struct {
	cli      *Client
	service  *register.ServiceInstance
	registry map[string]*service
	lock     sync.RWMutex
	timeout  time.Duration
}

func (r *Registry) Register(ctx context.Context, service *register.ServiceInstance) (err error) {
	r.service = service
	return r.cli.Register(ctx, service)
}

func (r *Registry) Deregister(ctx context.Context, service *register.ServiceInstance) error {
	return r.cli.Deregister(ctx, service.ID)
}

func (r *Registry) Watch(ctx context.Context, name string) (register.Watcher, error) {
	// 执行加锁
	r.lock.Lock()
	defer r.lock.Unlock()
	var (
		set *service
		ok  bool
	)
	if set, ok = r.registry[name]; !ok {
		set = &service{
			serviceName: name,
			atoValue:    new(atomic.Value),
			wathcer:     make(map[*watcher]struct{}),
		}
		set.ctx, set.cancel = context.WithCancel(context.Background())
		r.registry[name] = set
	}
	w := &watcher{
		event: make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.set = set
	set.lock.Lock()
	set.wathcer[w] = struct{}{}
	set.lock.Unlock()
	if ss, ok := set.atoValue.Load().([]*register.ServiceInstance); ok && len(ss) > 0 {
		w.event <- struct{}{}
	} else {
		if err := r.resolve(set.ctx, set); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func (r *Registry) resolve(ctx context.Context, ss *service) (err error) {
	outCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	var (
		entries []*register.ServiceInstance
		idx     uint64
	)
	if entries, idx, err = r.cli.Service(outCtx, ss.serviceName, 0, true); err != nil {
		return
	}
	// 进行广播
	if len(entries) > 0 {
		ss.broadcast(entries)
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				timeOutCtx, cancel := context.WithTimeout(context.Background(), r.timeout)
				tmp, tmpIdx, err := r.cli.Service(timeOutCtx, ss.serviceName, idx, true)
				if err != nil {
					cancel()
					time.Sleep(time.Second)
					continue
				}
				if len(tmp) != 0 && tmpIdx != idx {
					entries = tmp
					ss.broadcast(entries)
				}
				idx = tmpIdx
			case <-ctx.Done():
				r.lock.Lock()
				delete(r.registry, ss.serviceName)
				r.lock.Unlock()
			}
		}
	}()
	return nil
}

func (r *Registry) Close() error {
	r.registry = nil
	r.cli.Close()
	// 注销服务
	if r.service != nil {
		_ = r.Deregister(context.Background(), r.service)
	}
	return nil
}
