package etcd

// import (
// 	"context"
// 	"github.com/yanglunara/discovery/register"
// 	"go.etcd.io/etcd/clientv3"
// 	"google.golang.org/grpc"
// 	"os"
// 	"testing"
// 	"time"
// )

// func TestHeartBeat(t *testing.T) {
// 	client, err := clientv3.New(clientv3.Config{
// 		Endpoints:   []string{"127.0.0.1:2379"},
// 		DialTimeout: time.Second, DialOptions: []grpc.DialOption{grpc.WithBlock()},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer func() {
// 		_ = client.Close()
// 	}()

// 	ctx := context.Background()
// 	id, _ := os.Hostname()
// 	s := &register.ServiceInstance{
// 		ID:      id,
// 		Name:    "helloworld",
// 		Version: "v1.0",
// 	}
// 	opts := &Options{
// 		Ctx:       ctx,
// 		MaxRetry:  5,
// 		Ttl:       2 * time.Second,
// 		Namespace: "microservices",
// 	}
// 	go func() {
// 		//r := NewRegistry(client, opts)
// 		//w, err1 := r.Watch(ctx, s.Name)
// 		//if err1 != nil {
// 		//	return
// 		//}
// 		//defer func() {
// 		//	_ = w.Stop()
// 		//}()
// 		//for {
// 		//	res, err2 := w.Next()
// 		//	if err2 != nil {
// 		//		return
// 		//	}
// 		//	t.Logf("watch: %d", len(res))
// 		//	for _, r := range res {
// 		//		t.Logf("next: %+v", r)
// 		//	}
// 		//}
// 	}()
// 	time.Sleep(time.Second)

// 	// new a server
// 	r := NewRegistry(client, opts)

// 	//key := fmt.Sprintf("%s/%s/%s", r.Namespace, s.Name, s.ID)
// 	//value, _ := json.Marshal(s)
// 	//r.lease = clientv3.NewLease(r.client)
// 	_ = r.Register(ctx, s)
// 	//_, _ = r.withKV(ctx, key, string(value))
// 	//if err != nil {
// 	//	t.Fatal(err)
// 	//}
// 	//go r.heartBeat(ctx, leaseID, key, string(value))
// 	// wait for lease expired
// 	time.Sleep(10 * time.Second)

// 	//res, err := r.GetService(ctx, s.Name)
// 	//if err != nil {
// 	//	t.Fatal(err)
// 	//}
// 	//if len(res) != 0 {
// 	//	t.Errorf("not expected empty")
// 	//}

// 	//time.Sleep(time.Second * 30)
// 	//res, err = r.GetService(ctx, s.Name)
// 	//if err != nil {
// 	//	t.Fatal(err)
// 	//}
// 	//if len(res) == 0 {
// 	//	t.Errorf("reconnect failed")
// 	//}
// }
