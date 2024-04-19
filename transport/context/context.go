package context

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type mCtx struct {
	p1, p2     context.Context
	done       chan struct{}
	cancelCh   chan struct{}
	err        error
	atomicVal  uint32
	doneOnce   sync.Once
	cancelOnce sync.Once
}

func NewContext(c1, c2 context.Context) (context.Context, context.CancelFunc) {
	m := &mCtx{
		p1:       c1,
		p2:       c2,
		done:     make(chan struct{}),
		cancelCh: make(chan struct{}),
	}
	select {
	case <-m.p1.Done():
		_ = m.finish(m.p1.Err())
	case <-m.p2.Done():
		_ = m.finish(m.p2.Err())
	default:
		go m.wait()
	}
	return m, m.cancel
}

func (m *mCtx) finish(err error) error {
	m.doneOnce.Do(func() {
		m.err = err
		atomic.StoreUint32(&m.atomicVal, 1)
		close(m.done)
	})
	return m.err
}

func (m *mCtx) cancel() {
	m.cancelOnce.Do(func() {
		close(m.cancelCh)
	})
}

func (m *mCtx) wait() {
	var (
		err error
	)
	select {
	case <-m.p1.Done():
		err = m.p1.Err()
	case <-m.p2.Done():
		err = m.p2.Err()
	case <-m.cancelCh:
		err = context.Canceled
	}
	_ = m.finish(err)
}

func (m *mCtx) Done() <-chan struct{} {
	return m.done
}

func (m *mCtx) Err() error {
	if atomic.LoadUint32(&m.atomicVal) != 0 {
		return m.err
	}
	var (
		err error
	)
	select {
	case <-m.p1.Done():
		err = m.p1.Err()
	case <-m.p2.Done():
		err = m.p2.Err()
	case <-m.cancelCh:
		err = context.Canceled
	default:
		return nil
	}
	return m.finish(err)

}

func (m *mCtx) Deadline() (time.Time, bool) {
	d1, ok1 := m.p1.Deadline()
	d2, ok2 := m.p2.Deadline()
	switch {
	case !ok1:
		return d2, ok2
	}
	if !ok2 {
		return d1, ok1
	}
	if d1.Before(d2) {
		return d1, true
	}
	return d2, true
}

func (m *mCtx) Value(key interface{}) interface{} {
	if v := m.p1.Value(key); v != nil {
		return v
	}
	return m.p2.Value(key)
}
