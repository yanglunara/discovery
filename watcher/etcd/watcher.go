package etcd

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"github.com/yanglunara/discovery/register"
// 	"go.etcd.io/etcd/clientv3"
// )

// var (
// 	_ register.Watcher = (*watcher)(nil)
// )

// type watcher struct {
// 	key         string
// 	ctx         context.Context
// 	cancel      context.CancelFunc
// 	client      *clientv3.Client
// 	watchChan   clientv3.WatchChan
// 	watcher     clientv3.Watcher
// 	kv          clientv3.KV
// 	first       bool
// 	serviceName string
// }

// func NewWatcher(ctx context.Context, name string, client *clientv3.Client) register.Watcher {
// 	key := fmt.Sprintf("/%s/%s", string(register.NamespaceDefault), name)
// 	w := &watcher{
// 		key:         key,
// 		client:      client,
// 		watcher:     clientv3.NewWatcher(client),
// 		kv:          clientv3.NewKV(client),
// 		first:       true,
// 		serviceName: name,
// 	}

// 	w.ctx, w.cancel = context.WithCancel(ctx)
// 	//启动监听
// 	w.watchChan = w.watcher.Watch(w.ctx, key,
// 		clientv3.WithPrefix(),
// 		clientv3.WithRev(0),
// 		clientv3.WithKeysOnly(),
// 	)
// 	if err := w.watcher.RequestProgress(w.ctx); err != nil {
// 		panic(err)
// 	}
// 	return w
// }

// func (w *watcher) Next() ([]*register.ServiceInstance, error) {
// 	if w.first {
// 		item, err := w.getInstance()
// 		w.first = false
// 		return item, err
// 	}
// 	select {
// 	case <-w.ctx.Done():
// 		return nil, w.ctx.Err()
// 	case watchResp, ok := <-w.watchChan:
// 		if !ok || watchResp.Err() != nil {
// 			time.Sleep(time.Second)
// 			if err := w.reWatch(); err != nil {
// 				return nil, err
// 			}
// 		}
// 		for _, event := range watchResp.Events {
// 			fmt.Printf("event: %v\n", event)
// 		}
// 		return w.getInstance()
// 	}
// }

// func (w *watcher) Close() error {
// 	w.cancel()
// 	_ = w.client.Close()
// 	return w.watcher.Close()
// }

// func (w *watcher) getInstance() ([]*register.ServiceInstance, error) {
// 	resp, err := w.kv.Get(w.ctx, w.key, clientv3.WithPrefix())
// 	if err != nil {
// 		return nil, err
// 	}
// 	items := make([]*register.ServiceInstance, 0, len(resp.Kvs))
// 	for _, kv := range resp.Kvs {
// 		si := new(register.ServiceInstance)
// 		if err := json.Unmarshal(kv.Value, si); err != nil {
// 			return nil, err
// 		}
// 		if si.Name != w.serviceName {
// 			continue
// 		}
// 		items = append(items, si)
// 	}
// 	return items, nil
// }

// func (w *watcher) reWatch() error {
// 	_ = w.watcher.Close()
// 	w.watcher = clientv3.NewWatcher(w.client)
// 	w.watchChan = w.watcher.Watch(w.ctx, w.key, clientv3.WithPrefix(), clientv3.WithRev(0), clientv3.WithKeysOnly())
// 	return w.watcher.RequestProgress(w.ctx)
// }
