package handler

import (
	"context"
	"seckill-system/internal/order/service"
	"seckill-system/kitex_gen/order"
)

type OrderServiceImpl struct{}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, req *order.CreateOrderReq) (resp *order.CreateOrderResp, err error) {
	return service.NewOrderService(ctx).CreateOrder(req)
}

func (s *OrderServiceImpl) QueryOrder(ctx context.Context, req *order.QueryOrderReq) (resp *order.QueryOrderResp, err error) {
	return service.NewOrderService(ctx).QueryOrder(req)
}
