package registry

import (
	"context"
	"fmt"
	"sort"
)

type Registrar interface {
	// Register registers a new service instance.
	Register(ctx context.Context, service *Service) error
	// Deregister removes a previously registered service instance.
	Deregister(ctx context.Context, service *Service) error
}

type Discovery interface {
	GetService(ctx context.Context, serviceName string) ([]*Service, error)

	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

// Watcher watches for changes to service instances.
type Watcher interface {
	Next() ([]*Service, error)
	Stop() error
}

type Service struct {
	ID        string            `json:"id" description:"unique service id"`
	Name      string            `json:"name" description:"service name"`
	Version   string            `json:"version" description:"service version"`
	Address   string            `json:"address" description:"service address"`
	Endpoints []string          `json:"endpoints" description:"service endpoints"`
	Metadata  map[string]string `json:"metadata" description:"service metadata"`
}

func (s *Service) String() string {
	return fmt.Sprintf("%s-%s", s.Name, s.ID)
}

func (s *Service) check(t *Service) bool {
	return s.ID == t.ID && s.Name == t.Name && s.Version == t.Version
}

func (s *Service) Equal(o interface{}) bool {
	if s == nil && o == nil {
		return true
	}
	if s == nil || o == nil {
		return false
	}
	var (
		t  *Service
		ok bool
	)
	if t, ok = o.(*Service); !ok {
		return false
	}
	fn := func(a, b *Service) bool {
		if len(a.Endpoints) != len(b.Endpoints) {
			return false
		}
		sort.Strings(a.Endpoints)
		sort.Strings(b.Endpoints)
		for i := 0; i < len(a.Endpoints); i++ {
			if a.Endpoints[i] != b.Endpoints[i] {
				return false
			}
		}
		if len(a.Metadata) != len(b.Metadata) {
			return false
		}
		for k, v := range a.Metadata {
			if b.Metadata[k] != v {
				return false
			}
		}
		return a.ID == b.ID && a.Name == b.Name && a.Version == b.Version
	}
	return fn(s, t)
}
