package server

import (
	"fmt"
	nethttp "net/http"
	"strings"

	"github.com/curry-mz/sagittarius-golang/app"
	"github.com/curry-mz/sagittarius-golang/cores/server/http"
	"github.com/curry-mz/sagittarius-golang/cores/server/rpc"
	"github.com/curry-mz/sagittarius-golang/cores/server/socketio"
	"github.com/curry-mz/sagittarius-golang/cores/server/websocket"
	"github.com/curry-mz/sagittarius-golang/env"
	"github.com/curry-mz/sagittarius-golang/logger"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func InitRPCServer(opts ...rpc.ServerOption) *rpc.Server {
	cfg := app.Router().Config()
	// 初始化server
	// 找到rpc配置
	for _, svr := range cfg.Svrs {
		if strings.ToLower(svr.Proto) == env.ProtoRPC {
			opts = append(opts, rpc.Address(fmt.Sprintf(":%d", svr.Port)))
			break
		}
	}
	if len(opts) == 0 {
		panic("undefined rpc server port")
	}
	opts = append(opts, rpc.UnaryInterceptor(
		rpc.RecoverServerInterceptor(logger.GetLogger()),
		grpc_prometheus.UnaryServerInterceptor,
		rpc.TracingServerUnaryInterceptor(app.Router().Tracer()),
		rpc.AccessServerUnaryInterceptor(logger.GetAccess(), !cfg.AccessRequestDisable),
	))
	opts = append(opts, rpc.Options([]grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 16),
	}...))
	srv := rpc.NewServer(opts...)
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(srv.Server)

	nethttp.Handle("/metrics", promhttp.Handler())
	return srv
}

func InitWebSocketServer(opts ...websocket.Option) *websocket.Engine {
	cfg := app.Router().Config()
	// 初始化server
	// 找到rpc配置
	for _, svr := range cfg.Svrs {
		if strings.ToLower(svr.Proto) == env.ProtoWebsocket {
			opts = append(opts, websocket.Port(fmt.Sprintf("%d", svr.Port)))
			break
		}
	}
	if len(opts) == 0 {
		panic("undefined websocket server port")
	}
	// 初始化server
	var options []websocket.Option
	path := fmt.Sprintf("/%s/%s/ws", app.Router().Service().Product,
		strings.Join(strings.Split(app.Router().Service().ServiceName, "."), "/"))
	options = append(options, websocket.WsPath(path))
	options = append(options, websocket.Logger(logger.GetLogger()))
	opts = append(options, opts...)
	srv := websocket.NewServer(opts...)
	srv.Use(
		websocket.PanicHandler(logger.GetLogger()),
		websocket.TracingHandler(app.Router().Tracer()),
		websocket.LogHandler(logger.GetAccess(), !cfg.AccessRequestDisable),
	)
	return srv
}

func InitSocketIOServer(opts ...socketio.Option) *socketio.Engine {
	cfg := app.Router().Config()
	// 初始化server
	// 找到rpc配置
	for _, svr := range cfg.Svrs {
		if strings.ToLower(svr.Proto) == env.ProtoSocketIO {
			opts = append(opts, socketio.Port(fmt.Sprintf("%d", svr.Port)))
			break
		}
	}
	if len(opts) == 0 {
		panic("undefined socket.io server port")
	}
	// 初始化server
	srv := socketio.NewServer(opts...)
	srv.Use(
		socketio.PanicHandler(logger.GetLogger()),
		socketio.TracingHandler(app.Router().Tracer()),
		socketio.LogHandler(logger.GetAccess(), !cfg.AccessRequestDisable),
	)
	return srv
}

func InitHttpServer(opts ...http.Option) *http.Engine {
	cfg := app.Router().Config()
	// 初始化server
	// 找到rpc配置
	for _, svr := range cfg.Svrs {
		if strings.ToLower(svr.Proto) == env.ProtoHttp {
			opts = append(opts, http.Addr(fmt.Sprintf(":%d", svr.Port)))
			break
		}
	}
	if len(opts) == 0 {
		panic("undefined http server port")
	}
	srv := http.New(opts...)
	srv.Use(
		http.PanicHandler(logger.GetLogger()),
		http.TracingHandler(app.Router().Tracer()),
		http.LogHandler(logger.GetAccess(), !cfg.AccessRequestDisable),
	)
	return srv
}
