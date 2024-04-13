package consul

import (
	"context"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
)

var (
	_ register.Watcher = (*watcher)(nil)
)

type watcher struct {
	event  chan struct{}
	set    *service
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWatcher(ctx context.Context, client *api.Client) register.Watcher {
	return &watcher{}
}

func (w *watcher) Next() (arr []*register.ServiceInstance, err error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case <-w.event:
	}
	if ss, ok := w.set.atoValue.Load().([]*register.ServiceInstance); ok {
		arr = append(arr, ss...)
	}
	return
}

func (w *watcher) Close() error {
	w.cancel()
	w.set.lock.Lock()
	delete(w.set.wathcer, w)
	if len(w.set.wathcer) == 0 {
		w.set.cancel()
	}
	w.set.lock.Unlock()
	return nil
}
