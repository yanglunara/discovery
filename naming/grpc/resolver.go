package resolver

import (
	"log"
	"net/url"

	"github.com/yanglunara/discovery/naming"
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	nr       naming.Resolver
	cc       resolver.ClientConn
	quit     chan struct{}
	zone     string
	clusters map[string]struct{}
}

func (r *Resolver) Close() {
	select {
	case r.quit <- struct{}{}:
		_ = r.nr.Close()
	default:
	}
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *Resolver) watcher() {
	event := r.nr.Watch()
	for {
		select {
		case <-r.quit:
			return
		case _, ok := <-event:
			if !ok {
				return
			}
		}
		if ins, ok := r.nr.Fetch(); ok {
			instances, ok := ins.Instances[r.zone]
			if !ok {
				for _, value := range ins.Instances {
					instances = append(instances, value...)
				}
			}
			if len(instances) > 0 {
				r.newAddress(instances)
			}
		}
	}
}

func (r *Resolver) newAddress(instances []*naming.Instance) {
	var (
		addrs = make([]resolver.Address, 0, len(instances))
	)
	for _, ins := range instances {
		if len(r.clusters) > 0 {
			if _, ok := r.clusters[ins.Metadata["cluster"]]; !ok {
				continue
			}
		}
		var (
			grpcAddr string
		)
		for _, a := range ins.Addrs {
			u, err := url.Parse(a)
			if err == nil && u.Scheme == "grpc" {
				grpcAddr = u.Host
			}
		}
		addr := resolver.Address{
			Addr:       grpcAddr,
			ServerName: ins.AppID,
		}
		addrs = append(addrs, addr)
	}
	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		log.Fatalf("grpc resolver: UpdateState failed! error: %v", err)
	}
}
