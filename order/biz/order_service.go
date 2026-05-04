package service

import (
	"context"
	"order/kitex_gen/order"
)

type OrderService struct {
	ctx context.Context
}

func NewOrderService(ctx context.Context) *OrderService {
	return &OrderService{ctx: ctx}
}

func (s *OrderService) CreateOrder(req *order.CreateOrderReq) (*order.CreateOrderResp, error) {
	// TODO: Day3 写 MQ 消费逻辑
	return &order.CreateOrderResp{Code: 200, Msg: "ok"}, nil
}

func (s *OrderService) QueryOrder(req *order.QueryOrderReq) (*order.QueryOrderResp, error) {
	// TODO: Day3 写查询逻辑
	return &order.QueryOrderResp{Code: 200, Msg: "ok"}, nil
}
