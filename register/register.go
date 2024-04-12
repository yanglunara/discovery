package register

import (
	"context"

	"go.etcd.io/etcd/clientv3"
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
	Close() error
}

type Watcher interface {
	Next() ([]*ServiceInstance, error)
	Watcher() clientv3.WatchChan
	// Stop close the watcher.
	Stopper
}

type Stopper interface {
	Close() error
}

type Instance struct {
	Region   string            `json:"region"`           // Region 是地区
	Zone     string            `json:"zone"`             // Zone 是数据中心
	Env      string            `json:"env"`              // Env 是环境，如生产环境/预发布环境、用户验收测试环境/功能测试环境
	AppID    string            `json:"appid"`            // AppID 是映射服务树的应用ID
	Hostname string            `json:"hostname"`         // Hostname 是来自 Docker 的主机名
	Addrs    []string          `json:"addrs"`            // Addrs 是应用实例的地址，格式为：scheme://host
	Version  string            `json:"version"`          // Version 是发布版本
	LastTs   int64             `json:"latest_timestamp"` // LastTs 是实例最新更新的时间戳
	Metadata map[string]string `json:"metadata"`         // Metadata 是与 Addr 关联的信息，可能会被用于做负载均衡决策
}
type Namespace string

const (
	NamespaceDefault Namespace = "microservices"
)

type Instances struct {
	Instances map[string][]*Instance `json:"instances"`
	LastTs    int64                  `json:"latest_timestamp"`
	Scheduler []Zone                 `json:"scheduler"`
}

type Zone struct {
	Src string           `json:"src"`
	Dst map[string]int64 `json:"dst"`
}

type Resolver interface {
	Fetch() (*Instances, bool)
	Watch() <-chan struct{}
	Close() error
}

type Builder interface {
	Build(id string) Resolver
	Scheme() string
}
