package app

import (
	"context"
	"os"
	"strings"

	"github.com/curry-mz/sagittarius-golang/app/config"
	"github.com/curry-mz/sagittarius-golang/consul"
	"github.com/curry-mz/sagittarius-golang/cores/logger"
	"github.com/curry-mz/sagittarius-golang/cores/metric"
	"github.com/curry-mz/sagittarius-golang/cores/metric/local"
	"github.com/curry-mz/sagittarius-golang/cores/metric/pprof"
	"github.com/curry-mz/sagittarius-golang/cores/registry"
	cConsul "github.com/curry-mz/sagittarius-golang/cores/registry/consul"
	cEtcd "github.com/curry-mz/sagittarius-golang/cores/registry/etcd"
	"github.com/curry-mz/sagittarius-golang/cores/tracing"
	"github.com/curry-mz/sagittarius-golang/cores/tracing/jaeger"
	"github.com/curry-mz/sagittarius-golang/env"
	"github.com/curry-mz/sagittarius-golang/etcd"
	gLog "github.com/curry-mz/sagittarius-golang/logger"

	"github.com/getsentry/sentry-go"
)

func initLogger(cfg *config.LogConfig) {
	if cfg != nil {
		// 日志配置
		var opts []logger.Option
		// 设置日志路径
		if env.GetLogPath() != "" {
			opts = append(opts, logger.SetPath(env.GetLogPath()))
		}
		// 设置日志有效期
		if cfg.SaveDays > 0 {
			opts = append(opts, logger.SetSaveDays(cfg.SaveDays))
		}
		// 设置日志分割
		if cfg.Rotation != "" {
			rot := strings.ToLower(cfg.Rotation)
			rotation := logger.RotationDay
			if rot == "hour" {
				rotation = logger.RotationHour
			}
			opts = append(opts, logger.SetRotation(rotation))
		}
		// 设置格式
		if cfg.Format == logger.JsonFormat || cfg.Format == logger.ConsoleFormat {
			opts = append(opts, logger.SetFormat(cfg.Format))
		}
		// 初始化日志
		gLog.InitLogger(cfg.Level, opts...)
	} else {
		// 初始化日志
		gLog.InitLogger("")
	}
}

func initSentry(ctx context.Context, fullName string) {
	dns := env.GetSentryDNS()
	if dns == "" {
		return
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dns,
		AttachStacktrace: true,
		ServerName:       fullName,
	}); err != nil {
		gLog.Gen(ctx, "init sentry error:%v", err)
	}
}

func initTracer(fullName string) tracing.Tracer {
	addr := env.GetJaegerAddr()
	// tracer配置
	var opts []jaeger.Option
	// jaeger收集器配置
	if addr != "" {
		opts = append(opts, jaeger.WithAddr(addr))
	}
	// 初始化tracer
	return jaeger.NewTracer(fullName, opts...)
}

func initDiscovery(ctx context.Context, cfg *config.ServiceConfig) registry.Discovery {
	if cfg.Discovery == nil {
		return nil
	}
	switch cfg.Discovery.Used {
	case "consul":
		addr := env.GetConsulAddr()
		if addr == "" {
			return nil
		}
		addr = strings.TrimRight(addr, "/")

		var clientOpts []consul.Option
		c := consul.NewConsulClient(addr, clientOpts...)
		// 生成服务发现
		var discoveryOpts []cConsul.Option
		discoveryOpts = append(discoveryOpts, cConsul.Context(ctx))
		register := cConsul.NewDiscovery(c, discoveryOpts...)
		return register
	default:
		// 默认支持etcd
		// 创建etcd客户端
		eps := env.GetEtcdEndpoints()
		if len(eps) == 0 {
			return nil
		}
		var clientOpts []etcd.Option
		c := etcd.NewEtcdClient(eps, clientOpts...)
		// 生成服务发现
		var discoveryOpts []cEtcd.Option
		discoveryOpts = append(discoveryOpts,
			cEtcd.Context(ctx),
			cEtcd.Namespace(r.info.Namespace),
			cEtcd.Product(r.info.Product))
		register := cEtcd.NewDiscovery(c, discoveryOpts...)
		return register
	}
}

func initMetric(ctx context.Context, cfg []*config.ServerConfig) []metric.IMetric {
	if strings.ToLower(os.Getenv(env.MetricDisable)) == env.TRUE {
		return nil
	}
	// 找到最大配置端口
	port := 0
	for _, c := range cfg {
		if c.Port > port {
			port = c.Port
		}
	}
	if port == 0 {
		port = 8801
	} else {
		port += 1
	}
	return []metric.IMetric{local.InitMetric(ctx), pprof.InitMetric(ctx, pprof.SetPort(port))}
}
