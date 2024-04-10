package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/yunbaifan/discovery/registry"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

func TestNewRegistry(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	clientv3, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5, DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = clientv3.Close()
	}()
	ctx := context.Background()

	s := &registry.Service{
		ID:   "0",
		Name: "helloworld",
	}
	r := NewRegistry(ctx, clientv3)
	w, err := r.Watch(ctx, s.Name)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = w.Stop()
	}()
	go func() {
		for {
			res, err := w.Next()
			if err != nil {
				return
			}
			t.Logf("watch: %d", len(res))
			for _, r := range res {
				t.Logf("next: %+v", r)
			}
		}
	}()
	if err1 := r.Register(ctx, s); err1 != nil {
		t.Fatal(err1)
	}
	res, err := r.Get(ctx, s.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 && res[0].Name != s.Name {
		t.Errorf("not expected: %+v", res)
	}
	if err1 := r.Deregister(ctx, s); err1 != nil {
		t.Fatal(err1)
	}
	res, err = r.Get(ctx, s.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Errorf("not expected empty")
	}
	t.Logf("success")
}
