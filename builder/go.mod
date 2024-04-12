module github.com/yanglunara/builder/v2

go 1.20


replace github.com/coreos/bbolt v1.3.9 => go.etcd.io/bbolt v1.3.9

require (
	google.golang.org/grpc v1.63.2
	go.etcd.io/etcd/client/v3 v3.5.13
)
