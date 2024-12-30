package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type Option func(*option)

type option struct {
	dialTimeout string
}

func DialTimeout(dialTimeout string) Option {
	return func(o *option) {
		o.dialTimeout = dialTimeout
	}
}

func NewEtcdClient(eps []string, opts ...Option) *clientv3.Client {
	o := option{
		dialTimeout: "10s",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	td, err := time.ParseDuration(o.dialTimeout)
	if err != nil {
		panic(err)
	}
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   eps,
		DialTimeout: td,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		panic(err)
	}
	return c
}
