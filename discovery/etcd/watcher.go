package etcd

import (
	"context"
	"time"

	"github.com/yunbaifan/discovery/registry"
	"go.etcd.io/etcd/clientv3"
)

var (
	_ registry.Watcher = (*watcher)(nil)
)

type watchConf struct {
	Key, Name string
	Client    *clientv3.Client
}

type watcher struct {
	ctx    context.Context
	cancel context.CancelFunc
	*watchConf
	watchChan clientv3.WatchChan
	watcher   clientv3.Watcher
	kv        clientv3.KV
	first     bool
	Inter     Get
}

func newWatcher(ctx context.Context, conf *watchConf) (*watcher, error) {
	w := &watcher{
		watchConf: conf,
		kv:        clientv3.NewKV(conf.Client),
		watcher:   clientv3.NewWatcher(conf.Client),
		first:     true,
		Inter:     &get{},
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.watchChan = w.watcher.Watch(
		w.ctx, conf.Key,
		clientv3.WithPrefix(),
		clientv3.WithRev(0),
		clientv3.WithKeysOnly(),
	)
	if err := w.watcher.RequestProgress(w.ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *watcher) Next() ([]*registry.Service, error) {
	// 第一次调用
	if w.first {
		w.first = false
		return w.Inter.Get(w.ctx, w.Key, w.Name, w.kv)
	}
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case watchResp, ok := <-w.watchChan:
		if !ok || watchResp.Err() != nil {
			time.Sleep(time.Second)
			if err := w.restWatch(); err != nil {
				return nil, err
			}
		}
		return w.Inter.Get(w.ctx, w.Key, w.Name, w.kv)
	}
}

func (w *watcher) Stop() error {
	w.cancel()
	return w.watcher.Close()
}

func (w *watcher) restWatch() error {
	// 先关闭调用
	_ = w.watcher.Close()
	w.watcher = clientv3.NewWatcher(w.Client)
	w.watchChan = w.watcher.Watch(w.ctx, w.watchConf.Key, clientv3.WithPrefix(), clientv3.WithRev(0), clientv3.WithKeysOnly())
	return w.watcher.RequestProgress(w.ctx)
}
