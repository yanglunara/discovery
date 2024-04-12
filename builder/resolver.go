package builder

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/yanglunara/discovery/register"
	"google.golang.org/grpc/resolver"
)

type discoveryResolver struct {
	w      register.Watcher
	cc     resolver.ClientConn
	d      register.Discovery
	ctx    context.Context
	cancel context.CancelFunc
}

func (r *discoveryResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *discoveryResolver) Close() {
	r.cancel()
	_ = r.d.Close()
}

func (r *discoveryResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		var (
			ins []*register.ServiceInstance
			err error
		)
		if ins, err = r.w.Next(); err != nil {
			if errors.Is(err, context.Canceled) {
				time.Sleep(time.Second)
				continue
			}
		}
		r.update(ins)
	}
}

func (r *discoveryResolver) update(ins []*register.ServiceInstance) {
	var addrs []resolver.Address
	for _, in := range ins {
		for _, addr := range in.Endpoints {
			u, err := url.Parse(addr)
			if err == nil && u.Scheme == "grpc" {
				addrs = append(addrs, resolver.Address{Addr: u.Host, ServerName: in.Name})
			}
		}
	}
	fmt.Printf("log.Logger: %#v %d \n ", addrs, len(addrs))
	//r.cc.UpdateState(resolver.State{Addresses: addrs})
}
