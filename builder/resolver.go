package builder

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/yanglunara/discovery/register"
	"google.golang.org/grpc/attributes"
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

func (r *discoveryResolver) ParseEndpoint(endpoints []string) (string, error) {
	for _, addr := range endpoints {
		if u, err := url.Parse(addr); err == nil && u.Scheme == "grpc" {
			return u.Host, nil
		} else {
			if err != nil {
				return "", err
			}
		}
	}
	return "", nil
}

func (r *discoveryResolver) update(ins []*register.ServiceInstance) {
	var (
		endpoints = make(map[string]struct{})
		filtered  = make([]*register.ServiceInstance, 0, len(ins))
	)
	for _, in := range ins {
		ept, err := r.ParseEndpoint(in.Endpoints)
		if err != nil || ept == "" {
			continue
		}
		if _, ok := endpoints[ept]; ok {
			continue
		}
		filtered = append(filtered, in)
	}
	addrs := make([]resolver.Address, 0, len(filtered))
	parseAttributes := func(metadata map[string]string) *attributes.Attributes {
		a := new(attributes.Attributes)
		for k, v := range metadata {
			a = a.WithValue(k, v)
		}
		return a
	}
	for _, in := range filtered {
		ept, _ := r.ParseEndpoint(in.Endpoints)
		endpoints[ept] = struct{}{}
		addr := resolver.Address{
			ServerName: in.Name,
			Attributes: parseAttributes(in.Metadata).WithValue("rawServiceInstance", in),
			Addr:       ept,
		}
		addrs = append(addrs, addr)
	}
	if len(addrs) == 0 {
		return
	}
	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		fmt.Printf("[resolver] failed to update state: %s \n", err.Error())
	}
}
