package naming

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
