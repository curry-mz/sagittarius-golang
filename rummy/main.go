package main

import (
	"flag"

	"code.cd.local/games-go/mini/rummy.busi/conf"
	"code.cd.local/games-go/mini/rummy.busi/server"

	"github.com/curry-mz/sagittarius-golang/app"
	"github.com/curry-mz/sagittarius-golang/app/config"
	"github.com/curry-mz/sagittarius-golang/env"
)

const (
	Namespace   = "games-cd"
	Product     = "mini"
	ServiceName = "rummy.busi"
)

var (
	confPath = flag.String("confPath", "", "confPath is none")
)

func init() {
	flag.Parse()

	var opts []config.Option
	if *confPath != "" && env.IsTesting() {
		opts = append(opts, config.WithPath(*confPath))
	}
	app.InitRouter(&config.ServiceDefine{
		Namespace:   Namespace,
		Product:     Product,
		ServiceName: ServiceName,
	}, opts...)
}

func main() {
	conf.InitAppConfig()
	conf.InitKafka()
	srv := server.NewServer()
	app.Router().BindServer(srv)

	app.Run()
}
