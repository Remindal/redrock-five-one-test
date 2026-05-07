package main

import (
	"net"
	"order/biz/dal"
	"order/biz/mq"
	"order/conf"

	"order/kitex_gen/order/orderservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	cfg := conf.LoadConfig()

	dal.Init(cfg)
	dal.InitRabbitMQ(cfg.RabbitMQ.Addr)

	go mq.StartConsumer()

	r, err := etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		panic(err)
	}

	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+cfg.Port)
	svr := orderservice.NewServer(
		new(OrderServiceImpl),
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
