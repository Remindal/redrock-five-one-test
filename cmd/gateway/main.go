package main

import (
	"seckill-system/internal/gateway/conf"
	"seckill-system/internal/gateway/dao"
	"seckill-system/internal/gateway/handler"
	"seckill-system/internal/gateway/middleware"
	"seckill-system/internal/gateway/rpc"
	"seckill-system/pkg/logger"

	"time"

	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

func main() {
	logger.Init()
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
