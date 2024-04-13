package consul

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/yanglunara/discovery/register"
)

var (
	_ register.Stopper = (*service)(nil)
)

type service struct {
	serviceName string
	wathcer     map[*watcher]struct{}
	atoValue    *atomic.Value
	lock        sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func (s *service) Close() error {
	s.cancel()
	return nil
}

func (s *service) broadcast(ss []*register.ServiceInstance) {
	s.atoValue.Store(ss)
	s.lock.RLock()
	defer s.lock.RUnlock()
	for w := range s.wathcer {
		select {
		case <-s.ctx.Done():
			return
		case w.event <- struct{}{}:
		default:
		}
	}
}
