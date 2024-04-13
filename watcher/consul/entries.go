package consul

import (
	"context"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
)

var (
	_ register.Entries = (*entries)(nil)
)

type entries struct {
	resolver register.Resolver
	cli      *api.Client
}

func NewEntries(resolver register.Resolver, cli *api.Client) register.Entries {
	return &entries{
		resolver,
		cli,
	}
}

func (e *entries) MultiDCService(ctx context.Context, en *register.EntriesOption) ([]*register.ServiceInstance, uint64, error) {
	var (
		services []*register.ServiceInstance
	)
	dcs, err := e.cli.Catalog().Datacenters()
	if err != nil {
		return nil, 0, err
	}
	resolver := e.resolver
	for _, dc := range dcs {
		en.Opts.Datacenter = dc
		e, m, err := e.singleDCEntries(en.Service, "", en.PassingOnly, en.Opts)
		if err != nil {
			return nil, 0, err
		}
		ins := resolver.ServiceResolver(ctx, e)
		for _, in := range ins {
			if in.Metadata == nil {
				in.Metadata = make(map[string]string, 1)
			}
			in.Metadata["dc"] = dc
		}
		services = append(services, ins...)
		en.Opts.WaitIndex = m.LastIndex
	}
	return services, en.Opts.WaitIndex, nil
}
func (e *entries) SingleDCEntries(ctx context.Context, en *register.EntriesOption) ([]*register.ServiceInstance, uint64, error) {
	entries, meta, err := e.singleDCEntries(en.Service, "", en.PassingOnly, en.Opts)
	if err != nil {
		return nil, 0, err
	}
	return e.resolver.ServiceResolver(ctx, entries), meta.LastIndex, nil
}

func (e *entries) singleDCEntries(service, tag string, passingOnly bool, opts *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	return e.cli.Health().Service(service, tag, passingOnly, opts)
}
