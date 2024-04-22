package consul

import (
	"context"
	"fmt"
	"strings"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
)

var (
	_ register.Resolver = (*resolver)(nil)
)

type resolver struct {
	ctx context.Context
}

func NewResolver(ctx context.Context) register.Resolver {
	return &resolver{
		ctx: ctx,
	}
}

func (r *resolver) ServiceResolver(ctx context.Context, entries []*consulApi.ServiceEntry) []*register.ServiceInstance {
	services := make([]*register.ServiceInstance, 0, len(entries))
	for _, entry := range entries {
		var version string
		for _, tag := range entry.Service.Tags {
			if ss := strings.SplitN(tag, "=", 2); len(ss) == 2 && ss[0] == "version" {
				version = ss[1]
			}
		}
		endpoints := make([]string, 0)
		for scheme, addr := range entry.Service.TaggedAddresses {
			// 检查是否是合法的scheme
			if scheme == "lan_ipv4" || scheme == "wan_ipv4" || scheme == "lan_ipv6" || scheme == "wan_ipv6" {
				continue
			}
			endpoints = append(endpoints, addr.Address)
		}
		if len(endpoints) == 0 && entry.Service.Address != "" && entry.Service.Port > 0 {
			addres := fmt.Sprintf("http://%s:%d", entry.Service.Address, entry.Service.Port)
			endpoints = append(endpoints, addres)
		}
		services = append(services, &register.ServiceInstance{
			ID:        entry.Service.ID,
			Name:      entry.Service.Service,
			Metadata:  entry.Service.Meta,
			Version:   version,
			Endpoints: endpoints,
		})
	}
	return services
}
