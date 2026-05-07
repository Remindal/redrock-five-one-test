package main

import (
	"activity/biz/dal"
	"activity/conf"
	"activity/kitex_gen/activity/activityservice"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	cfg := conf.LoadConfig()

	dal.Init(cfg)

	r, err := etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		panic(err)
	}

	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+cfg.Port)
	svr := activityservice.NewServer(
		new(ActivityServiceImpl),
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
