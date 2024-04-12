package register

import (
	"context"
)

type ServiceInstance struct {
	// ID is the unique instance ID as registered.
	ID string `json:"id"`
	// Name is the service name as registered.
	Name string `json:"name"`
	// Version is the version of the compiled.
	Version string `json:"version"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string `json:"metadata"`
	// schema:
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []string `json:"endpoints"`
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
}

type Watcher interface {
	Next() ([]*ServiceInstance, error)
	// Stop close the watcher.
	Stop() error
}
