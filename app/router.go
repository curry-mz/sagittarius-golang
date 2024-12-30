package app

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"code.cd.local/sagittarius/sagittarius-golang/app/config"
	gCtx "code.cd.local/sagittarius/sagittarius-golang/context"
	"code.cd.local/sagittarius/sagittarius-golang/cores/metric"
	"code.cd.local/sagittarius/sagittarius-golang/cores/registry"
	"code.cd.local/sagittarius/sagittarius-golang/cores/server"
	"code.cd.local/sagittarius/sagittarius-golang/cores/tracing"
	"code.cd.local/sagittarius/sagittarius-golang/env"
	"code.cd.local/sagittarius/sagittarius-golang/logger"
	"code.cd.local/sagittarius/sagittarius-golang/nacos"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type router struct {
	baseCtx     context.Context
	cancel      func()
	cfgStr      string
	cfgChangeCh chan struct{}

	info      *registry.Service
	discovery registry.Discovery
	cfg       config.ServiceConfig
	nacosCli  *nacos.Client
	tracer    tracing.Tracer
	metrics   []metric.IMetric
	srvs      []server.Server
}

func Router() *router {
	return r
}

func (r *router) Ctx() context.Context {
	return r.baseCtx
}

func (r *router) Config() *config.ServiceConfig {
	return &r.cfg
}

func (r *router) OnExtraConfigChange() chan struct{} {
	return r.cfgChangeCh
}

func (r *router) ExtraJsonConfig(v interface{}) error {
	return json.Unmarshal([]byte(r.cfgStr), v)
}

func (r *router) ExtraXmlConfig(v interface{}) error {
	return xml.Unmarshal([]byte(r.cfgStr), v)
}

func (r *router) ExtraYamlConfig(v interface{}) error {
	return yaml.Unmarshal([]byte(r.cfgStr), v)
}

func (r *router) Discovery() registry.Discovery {
	return r.discovery
}

func (r *router) Tracer() tracing.Tracer {
	return r.tracer
}

func (r *router) BindServer(srv ...server.Server) {
	r.srvs = append(r.srvs, srv...)
}

func (r *router) Service() *registry.Service {
	return r.info
}

func (r *router) CustomConfig(name string) (*config.Custom, error) {
	return config.New(r.info.Namespace, r.info.Product, r.info.ServiceName, name)
}

var (
	once sync.Once
	r    *router
)

func InitRouter(sd *config.ServiceDefine, opts ...config.Option) {
	once.Do(func() {
		// 设置runtime
		runtime.GOMAXPROCS(runtime.NumCPU())

		if sd.Namespace == "" || sd.Product == "" || sd.ServiceName == "" {
			panic("service undefined")
		}
		r = &router{cfgChangeCh: make(chan struct{})}
		// 读取配置
		cli, cfgStr, err := config.Initialize(sd, &r.cfg, opts...)
		if err != nil {
			panic(err)
		}
		r.nacosCli = cli
		r.cfgStr = cfgStr
		// 确保服务器 GetTime 肯定会成功,因此忽略掉 error
		u, _ := uuid.NewUUID()
		// 初始化服务信息
		hosts := make(map[string]string)
		for _, srv := range r.cfg.Svrs {
			if strings.ToLower(srv.Proto) != env.ProtoHttp &&
				strings.ToLower(srv.Proto) != env.ProtoWebsocket &&
				strings.ToLower(srv.Proto) != env.ProtoRPC &&
				strings.ToLower(srv.Proto) != env.ProtoSocketIO {
				panic("service proto not support")
			}
			hosts[strings.ToLower(srv.Proto)] = fmt.Sprintf("%s:%d", clientIP(), srv.Port)
		}
		r.info = &registry.Service{
			ID:          u.String(),
			Namespace:   sd.Namespace,
			Product:     sd.Product,
			ServiceName: sd.ServiceName,
			Hosts:       hosts,
			Tags:        env.GetRunEnv(),
		}
		// 初始化context信息
		ctx, cancel := context.WithCancel(context.Background())
		ctx = gCtx.NewServerContext(ctx, gCtx.TransData{
			Endpoint:    clientIP(),
			Namespace:   sd.Namespace,
			Product:     sd.Product,
			ServiceName: sd.ServiceName,
		})
		r.baseCtx = ctx
		r.cancel = cancel
		// 生成fullname
		fullName := fmt.Sprintf("%s.%s.%s", sd.Namespace, sd.Product, sd.ServiceName)
		// 初始化日志
		initLogger(r.cfg.Log)
		// 初始化sentry
		initSentry(r.baseCtx, fullName)
		// 初始化链路追踪
		r.tracer = initTracer(fullName)
		// 初始化服务发现
		r.discovery = initDiscovery(ctx, &r.cfg)
		// 初始化监控
		r.metrics = initMetric(r.baseCtx, r.cfg.Svrs)
		// 监听配置变化
		if r.nacosCli != nil {
			go func() {
				for {
					select {
					case <-r.baseCtx.Done():
						return
					case s := <-r.nacosCli.ListenConfig():
						log.Println("nacos config change, new data:", s)
						r.cfgStr = s
						r.cfgChangeCh <- struct{}{}
					}
				}
			}()
		}
		logger.Gen(r.baseCtx, "app %s init over", fullName)
	})
}

func Run() {
	eg, _ := errgroup.WithContext(r.baseCtx)
	// 开始基础监控
	if len(r.metrics) > 0 {
		for i := 0; i < len(r.metrics); i++ {
			m := r.metrics[i]
			m.Start()
			if m.Reports() == nil {
				continue
			}
			eg.Go(func() error {
				for {
					select {
					case report, ok := <-m.Reports():
						if !ok {
							return nil
						}
						logger.Gen(r.baseCtx, "\n%s", report)
					case <-r.baseCtx.Done():
						return nil
					}
				}
			})
		}
	}
	// 开启服务 & 监听stop
	if len(r.srvs) > 0 {
		for idx := 0; idx < len(r.srvs); idx++ {
			srv := r.srvs[idx]
			eg.Go(func() error {
				logger.Gen(r.baseCtx, "app %s.%s.%s running...",
					r.Service().Namespace, r.Service().Product, r.Service().ServiceName)
				if err := srv.Start(r.baseCtx); err != nil {
					return err
				}
				return nil
			})
		}
		eg.Go(func() error {
			select {
			case <-r.baseCtx.Done():
				logger.Gen(r.baseCtx, "app %s.%s.%s stopping...",
					r.Service().Namespace, r.Service().Product, r.Service().ServiceName)
				var errors []error
				for _, srv := range r.srvs {
					if err := srv.Stop(r.baseCtx); err != nil {
						errors = append(errors, err)
					}
				}
				if len(errors) > 0 {
					return errors[0]
				}
				return nil
			}
		})
	}
	// 开始服务注册
	if r.discovery != nil {
		if err := r.discovery.Register(r.baseCtx, r.info); err != nil {
			panic(err)
		}
		logger.Gen(r.baseCtx, "service %s register, %v", r.info.ServiceName, r.info)
	}
	// 优雅关闭处理
	c := make(chan os.Signal, 1)
	signal.Notify(c, []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}...)
	eg.Go(func() error {
		select {
		case <-r.baseCtx.Done():
			return nil
		case <-c:
			logger.Gen(r.baseCtx, "recv sig, app shutdown beginning...")
			return ShutDown()
		}
	})
	_ = eg.Wait()
}

func Start(stop func()) {
	eg, _ := errgroup.WithContext(r.baseCtx)
	// 开始基础监控
	if len(r.metrics) > 0 {
		for i := 0; i < len(r.metrics); i++ {
			m := r.metrics[i]
			m.Start()
			if m.Reports() == nil {
				continue
			}
			eg.Go(func() error {
				for {
					select {
					case report, ok := <-m.Reports():
						if !ok {
							return nil
						}
						logger.Gen(r.baseCtx, "\n%s", report)
					case <-r.baseCtx.Done():
						return nil
					}
				}
			})
		}
	}
	// 开启服务 & 监听stop
	if len(r.srvs) > 0 {
		for idx := 0; idx < len(r.srvs); idx++ {
			srv := r.srvs[idx]
			eg.Go(func() error {
				logger.Gen(r.baseCtx, "app %s.%s.%s running...",
					r.Service().Namespace, r.Service().Product, r.Service().ServiceName)
				if err := srv.Start(r.baseCtx); err != nil {
					return err
				}
				return nil
			})
		}
		eg.Go(func() error {
			select {
			case <-r.baseCtx.Done():
				logger.Gen(r.baseCtx, "app %s.%s.%s stopping...",
					r.Service().Namespace, r.Service().Product, r.Service().ServiceName)
				var errors []error
				for _, srv := range r.srvs {
					if err := srv.Stop(r.baseCtx); err != nil {
						errors = append(errors, err)
					}
				}
				if len(errors) > 0 {
					return errors[0]
				}
				return nil
			}
		})
	}
	// 开始服务注册
	if r.discovery != nil {
		if err := r.discovery.Register(r.baseCtx, r.info); err != nil {
			panic(err)
		}
		logger.Gen(r.baseCtx, "service %s register, %v", r.info.ServiceName, r.info)
	}
	// 优雅关闭处理
	c := make(chan os.Signal, 1)
	signal.Notify(c, []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}...)
	eg.Go(func() error {
		select {
		case <-r.baseCtx.Done():
			return nil
		case <-c:
			logger.Gen(r.baseCtx, "recv sig, app shutdown beginning...")
			if r.discovery != nil {
				if err := r.discovery.Deregister(r.baseCtx, r.info); err != nil {
					logger.Gen(r.baseCtx, "server shutdown, deregister error:%v", err)
					return err
				}
				logger.Gen(r.baseCtx, "service %s deregister, %v", r.info.ServiceName, r.info)
			}
			stop()
			r.cancel()
			r.baseCtx.Done()
			/*_, cancel := context.WithTimeout(r.baseCtx, 5*time.Second)
			cancel()*/
			//return ShutDown()
			return nil
		}
	})
	_ = eg.Wait()
}

// 改变退出次序。先关闭服务发现，然后结束room。最后整体退出
func ShutDown2() error {
	defer r.cancel()
	if r.discovery != nil {
		ctx, cancel := context.WithTimeout(r.baseCtx, 5*time.Second)
		defer cancel()
		if err := r.discovery.Deregister(ctx, r.info); err != nil {
			logger.Gen(r.baseCtx, "server shutdown, deregister error:%v", err)
			return err
		}
		logger.Gen(r.baseCtx, "service %s deregister, %v", r.info.ServiceName, r.info)
	}
	return nil
}

func ShutDown() error {
	defer r.cancel()

	if r.discovery != nil {
		ctx, cancel := context.WithTimeout(r.baseCtx, 5*time.Second)
		defer cancel()
		if err := r.discovery.Deregister(ctx, r.info); err != nil {
			logger.Gen(r.baseCtx, "server shutdown, deregister error:%v", err)
			return err
		}
		logger.Gen(r.baseCtx, "service %s deregister, %v", r.info.ServiceName, r.info)
	}
	return nil
}

func clientIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	panic("get location ip addr failed")
}
