package handler

import (
	"context"
	"seckill-system/internal/seckill/service"
	"seckill-system/kitex_gen/seckill"
)

type SeckillServiceImpl struct{}

func (s *SeckillServiceImpl) DoSeckill(ctx context.Context, req *seckill.DoSeckillReq) (resp *seckill.DoSeckillResp, err error) {
	return service.NewSeckillService(ctx).DoSeckill(req)
}
