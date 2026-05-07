package main

import (
	"gateway/biz/dal"
	"gateway/biz/handler"
	"gateway/biz/middleware"
	"gateway/conf"
	"gateway/rpc"

	"time"

	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

func main() {
	cfg := conf.LoadConfig()

	dal.InitRedis(cfg)
	rpc.Init(cfg)

	h := server.Default(
		server.WithHostPorts(cfg.Host + ":" + cfg.Port),
	)

	h.Use(recovery.Recovery())
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	h.Use(middleware.RateLimit(10000, 1))

	h.POST("/api/auth/login", handler.Login)

	h.Use(middleware.JWTAuth())

	h.POST("/api/activity/create", handler.CreateActivity)
	h.POST("/api/seckill/do", handler.DoSeckill)
	h.GET("/api/order/query", handler.QueryOrder)

	h.Spin()
}
