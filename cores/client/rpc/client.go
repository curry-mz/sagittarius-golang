package rpc

import (
	"context"
	"crypto/tls"
	"fmt"

	"code.cd.local/sagittarius/sagittarius-golang/cores/client/rpc/resolver/direct"
	"code.cd.local/sagittarius/sagittarius-golang/cores/client/rpc/resolver/discovery"
	"code.cd.local/sagittarius/sagittarius-golang/cores/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type ClientOption func(o *clientOptions)

// WithEps 兜底endpoints
func WithEps(eps ...string) ClientOption {
	return func(o *clientOptions) {
		o.eps = eps
	}
}

// WithWatcher 服务发现监听
func WithWatcher(watcher registry.Watcher) ClientOption {
	return func(o *clientOptions) {
		o.watcher = watcher
	}
}

// WithTLS 加密传输设置
func WithTLS(tlsCfg *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsCfg = tlsCfg
	}
}

// WithUnaryInterceptor 拦截器
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.ints = in
	}
}

// WithOptions grpc option
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.grpcOpts = opts
	}
}

// WithBalancerName 负载均衡策略
func WithBalancerName(balancerName string) ClientOption {
	return func(o *clientOptions) {
		o.balancerName = balancerName
	}
}

type clientOptions struct {
	eps          []string
	watcher      registry.Watcher
	tlsCfg       *tls.Config
	ints         []grpc.UnaryClientInterceptor
	grpcOpts     []grpc.DialOption
	balancerName string
}

func DialContext(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, opts...)
}

func dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{
		balancerName: roundrobin.Name,
	}
	for _, o := range opts {
		o(&options)
	}
	if len(options.eps) == 0 && options.watcher == nil {
		return nil, fmt.Errorf("default endpoints is nil and service discovery is nil")
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, options.balancerName)),
		grpc.WithChainUnaryInterceptor(options.ints...),
	}
	if options.tlsCfg != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(options.tlsCfg)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var builder resolver.Builder
	if options.watcher != nil {
		builder = discovery.NewBuilder(
			options.watcher,
			discovery.WithEps(options.eps...),
		)
	} else {
		builder = direct.NewBuilder(direct.WithEps(options.eps...))
	}
	grpcOpts = append(grpcOpts, grpc.WithResolvers(builder))
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return grpc.DialContext(ctx, fmt.Sprintf("%s:///", builder.Scheme()), grpcOpts...)
}
