package register

import (
	"context"

	consulApi "github.com/hashicorp/consul/api"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID        string            `json:"id"`        // 注册的唯一实例ID
	Name      string            `json:"name"`      // 注册的服务名称
	Version   string            `json:"version"`   // 编译的版本
	Metadata  map[string]string `json:"metadata"`  // 与服务实例关联的键值对元数据
	Endpoints []string          `json:"endpoints"` // 端点
	//服务实例最后更新的时间戳
	LastTs int64 `json:"latest_timestamp"`
}

// Registrar 注册器接口
type Registrar interface {
	// 注册服务
	Register(ctx context.Context, service *ServiceInstance) error
	// 注销服务
	Deregister(ctx context.Context, service *ServiceInstance) error
}

// Discovery 服务发现接口
type Discovery interface {
	// 根据服务名称返回内存中的服务实例
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// 根据服务名称创建观察者
	Watch(ctx context.Context, serviceName string) (Watcher, error)
	// 关闭服务发现
	Close() error
}

// Watcher  观察者接口
type Watcher interface {
	Next() ([]*ServiceInstance, error) // 获取下一个服务实例
	Stopper                            // 停止观察者
}

// Builder 构建器接口
type Builder interface {
	Build(id string) Resolver // 构建解析器
	Scheme() string           // 获取方案
}

// Stopper 停止器接口
type Stopper interface {
	Close() error // 关闭停止器
}

// Resolver 解析器接口
type Resolver interface {
	ServiceResolver(ctx context.Context, entries []*consulApi.ServiceEntry) []*ServiceInstance // 服务解析
}

// 数据中心类型
const (
	SingleDataCenter = "SINGLE" // 单数据中心
	MultiDataCenter  = "MULTI"  // 多数据中心
)

// EntriesOption 条目选项
type EntriesOption struct {
	Resolver
	Service, Tag string
	Index        uint64
	PassingOnly  bool
	Opts         *consulApi.QueryOptions
}

// Entries 条目接口
type Entries interface {
	MultiDCService(ctx context.Context, en *EntriesOption) ([]*ServiceInstance, uint64, error)  // 多数据中心服务
	SingleDCEntries(ctx context.Context, en *EntriesOption) ([]*ServiceInstance, uint64, error) // 单数据中心条目
}
