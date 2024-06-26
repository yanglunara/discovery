package consul

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
)

var (
	_ register.Registrar = (*Registry)(nil)
	_ register.Stopper   = (*Registry)(nil)
	_ register.Discovery = (*Registry)(nil)
)

func NewRegistry(client *api.Client) *Registry {
	r := &Registry{
		registry: make(map[string]*service),
		timeout:  10 * time.Second,
		cli: &Client{
			cli:                            client,
			dc:                             "SINGLE",
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
	cli      *Client                   // 客户端实例
	service  *register.ServiceInstance // 服务实例
	registry map[string]*service       // 服务注册表，键为服务名，值为服务实例
	lock     sync.RWMutex              // 读写锁，用于保护服务注册表的并发访问
	timeout  time.Duration             // 超时时间
}

func (r *Registry) Register(ctx context.Context, service *register.ServiceInstance) (err error) {
	r.service = service
	return r.cli.Register(ctx, service)
}

func (r *Registry) Deregister(ctx context.Context, service *register.ServiceInstance) error {
	return r.cli.Deregister(ctx, service.ID)
}

func (r *Registry) GetService(tx context.Context, serviceName string) ([]*register.ServiceInstance, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	set := r.registry[serviceName]
	remote := func(cli *Client) []*register.ServiceInstance {
		if service, _, err := cli.Service(tx, serviceName, 0, true); err == nil && len(service) > 0 {
			return service
		}
		return nil
	}
	if set == nil {
		if list := remote(r.cli); len(list) > 0 {
			return list, nil
		}
		return nil, fmt.Errorf("service %s not resolved in registry", serviceName)
	}
	var (
		ss []*register.ServiceInstance
		ok bool
	)
	if ss, ok = set.atoValue.Load().([]*register.ServiceInstance); !ok && ss == nil {
		if list := remote(r.cli); len(list) > 0 {
			return list, nil
		}
		return nil, fmt.Errorf("service %s not resolved in registry", serviceName)
	}
	return ss, nil

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
	}
	if !ok {
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
