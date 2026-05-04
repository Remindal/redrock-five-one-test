package main

import (
	"context"
	"seckill/biz/service"
	"seckill/kitex_gen/seckill"
)

type SeckillServiceImpl struct{}

func (s *SeckillServiceImpl) DoSeckill(ctx context.Context, req *seckill.DoSeckillReq) (resp *seckill.DoSeckillResp, err error) {
	return service.NewSeckillService(ctx).DoSeckill(req)
}
