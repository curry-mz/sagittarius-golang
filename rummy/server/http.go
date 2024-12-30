package server

import (
	"context"

	"code.cd.local/games-go/mini/rummy.busi/service/manager"

	"code.cd.local/sagittarius/sagittarius-golang/app/server"
	"code.cd.local/sagittarius/sagittarius-golang/cores/server/http"
)

func NewServer() *http.Engine {
	service, err := manager.New(context.TODO())
	if err != nil {
		panic(err)
	}
	srv := server.InitHttpServer(http.OnStop(service.Stop))
	{
		srv.POST("/server_api/v1/proxy/websocket", service.ProxyMessageHandler)
		//srv.POST("/server_api/v1/conn/disconnect", service.DisconnectHandler)
	}
	return srv
}
