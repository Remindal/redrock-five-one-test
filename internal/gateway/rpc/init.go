package rpc

import (
	"seckill-system/internal/gateway/conf"
	"time"

	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"

	"seckill-system/kitex_gen/activity/activityservice"
	"seckill-system/kitex_gen/order/orderservice"
	"seckill-system/kitex_gen/seckill/seckillservice"
)

var (
	ActivityClient activityservice.Client
	SeckillClient  seckillservice.Client
	OrderClient    orderservice.Client
)

func Init(cfg *conf.Config) {
	r, err := etcd.NewEtcdResolver(cfg.Etcd.Endpoints)
	if err != nil {
		panic(err)
	}

	ActivityClient = activityservice.MustNewClient(
		cfg.Rpc.Activity,
		client.WithResolver(r),
		client.WithRPCTimeout(3*time.Second),
		client.WithConnectTimeout(3*time.Second),
	)
	SeckillClient = seckillservice.MustNewClient(
		cfg.Rpc.Seckill,
		client.WithResolver(r),
		client.WithRPCTimeout(3*time.Second),
		client.WithConnectTimeout(3*time.Second),
	)
	OrderClient = orderservice.MustNewClient(
		cfg.Rpc.Order,
		client.WithResolver(r),
		client.WithRPCTimeout(3*time.Second),
		client.WithConnectTimeout(3*time.Second),
	)
}
