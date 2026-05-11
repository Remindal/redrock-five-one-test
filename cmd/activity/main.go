package main

import (
	"net"
	"os"
	"seckill-system/internal/activity/conf"
	"seckill-system/internal/activity/dao"
	"seckill-system/internal/activity/handler"
	"seckill-system/kitex_gen/activity/activityservice"
	"seckill-system/pkg/logger"

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
	svr := activityservice.NewServer(
		new(handler.ActivityServiceImpl),
		server.WithRegistry(r),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "activity",
		}),
		server.WithServiceAddr(addr),
	)

	if err := svr.Run(); err != nil {
		panic(err)
	}
}
