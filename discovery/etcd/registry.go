package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/yunbaifan/discovery/registry"
	"go.etcd.io/etcd/clientv3"
)

var (
	_ registry.Registrar = (*Registry)(nil)
)

type KV func(ctx context.Context, client *clientv3.Client, lease clientv3.Lease, ttl time.Duration, key, value string) (leaseID clientv3.LeaseID, err error)

type Get interface {
	Get(ctx context.Context, key, name string, kv clientv3.KV) ([]*registry.Service, error)
}

var (
	_ Get = (*get)(nil)
)

type get struct {
}

func (g *get) Get(ctx context.Context, key, name string, kv clientv3.KV) ([]*registry.Service, error) {
	resp, err := kv.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	items := make([]*registry.Service, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var (
			s registry.Service
		)
		if err := json.Unmarshal(kv.Value, &s); err != nil {
			return nil, err
		}
		if s.Name != name {
			continue
		}
		items = append(items, &s)
	}
	return items, nil
}

type options struct {
	// 上下文
	ctx context.Context
	// 命名空间
	namespace string
	// 过期时间
	ttl time.Duration
	// 最大重试次数
	maxRetry int
}
type Registry struct {
	client *clientv3.Client
	kv     clientv3.KV
	//租约
	lease clientv3.Lease
	//上下文取消函数
	maps   map[*registry.Service]context.CancelFunc
	opt    *options
	withKV KV
	Inter  Get
}

func NewRegistry(ctx context.Context, client *clientv3.Client) *Registry {
	return &Registry{
		client: client,
		kv:     clientv3.NewKV(client),
		opt: &options{
			ctx:       ctx,
			namespace: "/microservices",
			ttl:       time.Second * 15,
			maxRetry:  5,
		},
		maps:  make(map[*registry.Service]context.CancelFunc),
		Inter: &get{},
	}
}

// Register registers a new service instance.
func (r *Registry) Register(ctx context.Context, service *registry.Service) (err error) {
	// 服务注册路径
	key := fmt.Sprintf("%s/%s/%s", r.opt.namespace, service.Name, service.ID)
	// 服务注册值
	var (
		buf     []byte
		leaseID clientv3.LeaseID
	)
	if buf, err = json.Marshal(service); err != nil {
		return
	}
	if r.lease != nil {
		defer func() {
			_ = r.lease.Close()
		}()
	}
	//用于创建一个新的租约客户端
	r.lease = clientv3.NewLease(r.client)
	//
	withKV := func(ctx context.Context, client *clientv3.Client, lease clientv3.Lease, ttl time.Duration, key, value string) (leaseID clientv3.LeaseID, err error) {
		var (
			grant *clientv3.LeaseGrantResponse
		)
		if grant, err = lease.Grant(ctx, int64(ttl.Seconds())); err != nil {
			return
		}
		if _, err = r.client.Put(ctx, key, value, clientv3.WithLease(grant.ID)); err != nil {
			return 0, err
		}
		return grant.ID, nil
	}
	// 注册服务
	if leaseID, err = withKV(ctx, r.client, r.lease, r.opt.ttl, key, string(buf)); err != nil {
		return
	}
	// 上下文 取消函数
	newCtx, cancel := context.WithCancel(r.opt.ctx)
	r.maps[service] = cancel
	// 心跳
	r.withKV = withKV

	go r.heartbeat(newCtx, leaseID, key, string(buf))

	return nil
}

func (r *Registry) heartbeat(ctx context.Context, leaseID clientv3.LeaseID, key, value string) {
	curLeaseID := leaseID
	ka, err := r.client.KeepAlive(ctx, leaseID)
	if err != nil {
		curLeaseID = 0
	}
	for {
		// 如果租约ID为0
		if curLeaseID == 0 {
			var (
				arr []int
			)
			// 重试机制
			for cnt := 0; cnt < r.opt.maxRetry; cnt++ {
				// 上下文错误时停止操作
				if ctx.Err() != nil {
					return
				}
				// 带缓冲的通道
				idChan, errChan := make(chan clientv3.LeaseID, 1), make(chan error, 1)
				cancelCtx, cancel := context.WithCancel(ctx)
				go func() {
					defer cancel()
					if id, err := r.withKV(cancelCtx, r.client, r.lease, r.opt.ttl, key, value); err != nil {
						errChan <- err
					} else {
						idChan <- id
					}
				}()
				select {
				case <-time.After(3 * time.Second):
					cancel()
					continue
				case <-errChan:
					return
				case curLeaseID = <-idChan:
				}
				if ka, err = r.client.KeepAlive(ctx, curLeaseID); err == nil {
					break
				}
				// 数字1 左移cnt位   如果 5  reat值 为 1 2 4 8 16
				// 如果 3 reat值 为 1 2 4
				arr = append(arr, 1<<cnt)
				// 随机等待
				time.Sleep(time.Duration(arr[rand.Intn(len(arr))]) * time.Second)
			}
		}
		select {
		case _, ok := <-ka:
			if !ok {
				if ctx.Err() != nil {
					return
				}
			}
			curLeaseID = 0
			continue
		case <-r.opt.ctx.Done():
			return
		}
	}
}

func (r *Registry) Deregister(ctx context.Context, service *registry.Service) error {
	defer func() {
		if r.lease != nil {
			_ = r.lease.Close()
		}
	}()
	// cancel heartbeat
	if cancel, ok := r.maps[service]; ok {
		cancel()
		delete(r.maps, service)
	}
	key := fmt.Sprintf("%s/%s/%s", r.opt.namespace, service.Name, service.ID)
	_, err := r.client.Delete(ctx, key)
	return err
}

func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newWatcher(ctx, &watchConf{
		Key:    fmt.Sprintf("%s/%s", r.opt.namespace, serviceName),
		Name:   serviceName,
		Client: r.client,
	})
}

func (r *Registry) Get(ctx context.Context, name string) ([]*registry.Service, error) {
	key := fmt.Sprintf("%s/%s", r.opt.namespace, name)
	return r.Inter.Get(ctx, key, name, r.kv)
}
