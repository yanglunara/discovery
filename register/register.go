package register

import (
	"context"

	consulApi "github.com/hashicorp/consul/api"
)

type ServiceInstance struct {
	// ID is the unique instance ID as registered.
	ID string `json:"id"`
	// Name is the service name as registered.
	Name string `json:"name"`
	// Version is the version of the compiled.
	Version string `json:"version"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata  map[string]string `json:"metadata"`
	Endpoints []string          `json:"endpoints"`
}

type Registrar interface {
	// Register the registration.
	Register(ctx context.Context, service *ServiceInstance) error
	// Deregister the registration.
	Deregister(ctx context.Context, service *ServiceInstance) error
}

type Discovery interface {
	// GetService return the service instances in memory according to the service name.
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// Watch creates a watcher according to the service name.
	Watch(ctx context.Context, serviceName string) (Watcher, error)
	Close() error
}

type Watcher interface {
	Next() ([]*ServiceInstance, error)
	// Stop close the watcher.
	Stopper
}

type Builder interface {
	Build(id string) Resolver
	Scheme() string
}

type Stopper interface {
	Close() error
}

type Resolver interface {
	ServiceResolver(ctx context.Context, entries []*consulApi.ServiceEntry) []*ServiceInstance
}

const (
	SingleDataCenter = "SINGLE"
	MultiDataCenter  = "MULTI"
)

type EntriesOption struct {
	Resolver
	Service, Tag string
	Index        uint64
	PassingOnly  bool
	Opts         *consulApi.QueryOptions
}

type Entries interface {
	MultiDCService(ctx context.Context, en *EntriesOption) ([]*ServiceInstance, uint64, error)
	SingleDCEntries(ctx context.Context, en *EntriesOption) ([]*ServiceInstance, uint64, error)
}

type Namespace string

const (
	NamespaceDefault Namespace = "microservices"
)

type DebugOption struct {
	IsDebug bool
}
