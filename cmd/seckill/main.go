package main

import (
	"net"
	"os"
	"seckill-system/internal/seckill/conf"
	"seckill-system/internal/seckill/dao"
	"seckill-system/internal/seckill/handler"
	"seckill-system/pkg/logger"

	"seckill-system/kitex_gen/seckill/seckillservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	logger.Init()
	cfg := conf.LoadConfig()

	dal.Init(cfg)

	r, err := etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		panic(err)
	}

	host := "127.0.0.1"
	if h := os.Getenv("ADVERTISE_HOST"); h != "" {
		host = h
	}
	addr, _ := net.ResolveTCPAddr("tcp", host+":"+cfg.Port)
	svr := seckillservice.NewServer(
		new(handler.SeckillServiceImpl),
		server.WithRegistry(r),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "seckill",
		}),
		server.WithServiceAddr(addr),
	)

	if err := svr.Run(); err != nil {
		panic(err)
	}
}
