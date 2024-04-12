package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yanglunara/discovery/register"
	"go.etcd.io/etcd/clientv3"
)

var (
	_ register.Registrar = (*Registry)(nil)

	_ register.Discovery = (*Registry)(nil)
)

type Options struct {
	Ctx       context.Context
	Namespace string // 命名空间
	Ttl       time.Duration
	MaxRetry  int //最大重试次数
}

type Registry struct {
	*Options
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	ctxMap map[*register.ServiceInstance]context.CancelFunc
}

func NewRegistry(client *clientv3.Client, opts *Options) *Registry {
	return &Registry{
		Options: opts,
		client:  client,
		kv:      clientv3.NewKV(client),
		lease:   clientv3.NewLease(client),
		ctxMap:  make(map[*register.ServiceInstance]context.CancelFunc),
	}
}

// withKV 注册数据到etcd
func (r *Registry) withKV(ctx context.Context, key string, value string) (id clientv3.LeaseID, err error) {
	var (
		grant *clientv3.LeaseGrantResponse
	)
	if grant, err = r.lease.Grant(ctx, int64(r.Ttl.Seconds())); err != nil {
		return
	}
	if _, err = r.kv.Put(ctx, key, value, clientv3.WithLease(grant.ID)); err != nil {
		return
	}
	return grant.ID, nil
}

// Register 注册服务实例
func (r *Registry) Register(ctx context.Context, service *register.ServiceInstance) (err error) {
	var (
		value []byte
	)
	if value, err = json.Marshal(service); err != nil {
		return
	}
	if r.lease != nil {
		defer func() {
			_ = r.lease.Close()
		}()
	}
	// 创建租约
	r.lease = clientv3.NewLease(r.client)
	var (
		leaseID clientv3.LeaseID
		key     string = r.Namespace + "/" + service.Name + "/" + service.ID
	)
	// 注册服务
	if leaseID, err = r.withKV(ctx,
		key,
		string(value)); err != nil {
		return
	}
	hCtx, cancel := context.WithCancel(ctx)
	// 根据服务实例创建一个上下文, 用于取消续租
	r.ctxMap[service] = cancel

	go r.heartBeat(hCtx, leaseID, key, string(value))
	return nil
}

func (r *Registry) GetService(ctx context.Context, name string) ([]*register.ServiceInstance, error) {
	resp, err := r.kv.Get(ctx, fmt.Sprintf("%s/%s", r.Namespace, name), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	items := make([]*register.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		si := new(register.ServiceInstance)
		if err := json.Unmarshal(kv.Value, si); err != nil {
			return nil, err
		}
		if si.Name != name {
			continue
		}
		items = append(items, si)
	}
	return items, nil
}

// Watch creates a watcher according to the service name.
func (r *Registry) Watch(ctx context.Context, name string) (register.Watcher, error) {
	return newWatcher(ctx, fmt.Sprintf("%s/%s", r.Namespace, name), name, r.client)
}

func (r *Registry) Deregister(ctx context.Context, service *register.ServiceInstance) (err error) {
	defer func() {
		if r.lease != nil {
			_ = r.lease.Close()
		}
	}()
	// 取消续租
	if cancel, ok := r.ctxMap[service]; ok {
		cancel()
		delete(r.ctxMap, service)
	}
	if _, err = r.client.Delete(
		ctx,
		r.Namespace+"/"+service.Name+"/"+service.ID); err != nil {
		return
	}
	return nil
}

// heartBeat 续租
func (r *Registry) heartBeat(ctx context.Context, leaseID clientv3.LeaseID, key string, value string) {
	curLeaseID := leaseID
	// 续租
	kac, err := r.lease.KeepAlive(ctx, curLeaseID)
	if err != nil {
		curLeaseID = 0
	}
	for {
		if curLeaseID == 0 {
			var arr []int
			for cnt := 0; cnt < r.MaxRetry; cnt++ {
				// 上下文出错 退出当前续租
				if ctx.Err() != nil {
					return
				}
				idChan := make(chan clientv3.LeaseID, 1)
				errChan := make(chan error, 1)
				cancelCtx, cancel := context.WithCancel(ctx)
				go func() {
					defer cancel()
					if id, err := r.withKV(cancelCtx, key, value); err != nil {
						errChan <- err
					} else {
						idChan <- id
					}
				}()
				// 等待续租结果
				select {
				// 3秒超时
				case <-time.After(time.Second * 3):
					cancel()
					continue
				case <-errChan:
					continue
				case curLeaseID = <-idChan:
				}
				kac, err = r.client.KeepAlive(ctx, curLeaseID)
				if err == nil {
					break
				}
				arr = append(arr, 1<<cnt)
				// 重试间隔
				time.Sleep(time.Second * time.Duration(arr[cnt]))
			}
			if _, ok := <-kac; !ok {
				return
			}
		}
		select {
		case _, ok := <-kac:
			if !ok {
				if ctx.Err() != nil {
					return
				}
				curLeaseID = 0
				continue
			}
		case <-r.Ctx.Done():
			return
		}
	}
}
