package resolver

import (
	"github.com/yanglunara/discovery/naming"
	"google.golang.org/grpc/resolver"
	"net/url"
	"os"
)

var (
	_ resolver.Builder  = &Builder{}
	_ resolver.Resolver = &Resolver{}
)

type Builder struct {
	naming.Builder
}

func Register(nb naming.Builder) {
	resolver.Register(&Builder{nb})
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	zone := os.Getenv("ZONE")
	clusters := map[string]struct{}{}
	if u, err := url.ParseQuery(target.URL.RawQuery); err == nil {
		if zones, ok := u["zone"]; ok && len(zones) > 0 {
			zone = zones[0]
		}
		for _, c := range u["cluster"] {
			clusters[c] = struct{}{}
		}
	}
	r := &Resolver{
		cc:       cc,
		nr:       b.Builder.Build(target.URL.Host),
		quit:     make(chan struct{}, 1),
		zone:     zone,
		clusters: clusters,
	}

	go r.watcher()

	return r, nil
}
