package main

import (
	"flag"

	"code.cd.local/games-go/mini/rummy.busi/conf"
	"code.cd.local/games-go/mini/rummy.busi/server"

	"code.cd.local/sagittarius/sagittarius-golang/app"
	"code.cd.local/sagittarius/sagittarius-golang/app/config"
	"code.cd.local/sagittarius/sagittarius-golang/env"
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
