package main

import (
	"gateway/biz/handler"
	"gateway/conf"
	"gateway/rpc"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	cfg := conf.LoadConfig()

	rpc.Init(cfg)

	h := server.Default(
		server.WithHostPorts(cfg.Host + ":" + cfg.Port),
	)

	h.POST("/api/activity/create", handler.CreateActivity)
	h.POST("/api/seckill/do", handler.DoSeckill)
	h.GET("/api/order/query", handler.QueryOrder)

	h.Spin()
}
