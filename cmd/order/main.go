package main

import (
	"net"
	"os"
	"seckill-system/internal/order/conf"
	"seckill-system/internal/order/dao"
	"seckill-system/internal/order/handler"
	"seckill-system/internal/order/mq"
	"seckill-system/pkg/logger"

	"seckill-system/kitex_gen/order/orderservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	logger.Init()
	cfg := conf.LoadConfig()

	dal.Init(cfg)
	dal.InitRabbitMQ(cfg.RabbitMQ.Addr)

	go mq.StartConsumer()

	r, err := etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		panic(err)
	}

	host := "127.0.0.1"
	if h := os.Getenv("ADVERTISE_HOST"); h != "" {
		host = h
	}
	addr, _ := net.ResolveTCPAddr("tcp", host+":"+cfg.Port)
	svr := orderservice.NewServer(
		new(handler.OrderServiceImpl),
		server.WithRegistry(r),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "order",
		}),
		server.WithServiceAddr(addr),
	)

	if err := svr.Run(); err != nil {
		panic(err)
	}
}
