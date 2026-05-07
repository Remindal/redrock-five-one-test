package service

import (
	"context"
	"fmt"
	"order/biz/dal"
	"order/biz/model"
	"order/kitex_gen/order"
	"time"
)

type OrderService struct {
	ctx context.Context
}

func NewOrderService(ctx context.Context) *OrderService {
	return &OrderService{ctx: ctx}
}

func (s *OrderService) CreateOrder(req *order.CreateOrderReq) (*order.CreateOrderResp, error) {

	var existing model.Order
	err := dal.DB.Where("activity_id = ? AND user_id = ?", req.ActivityId, req.UserId).First(&existing).Error
	if err == nil {
		return &order.CreateOrderResp{
			Code:    200,
			Msg:     "订单已存在",
			OrderId: &existing.ID,
		}, nil
	}

	o := model.Order{
		ActivityId: req.ActivityId,
		UserId:     req.UserId,
		Status:     "SUCCESS",
	}
	if err := dal.DB.Create(&o).Error; err != nil {
		//再次查询
		if dal.DB.Where("activity_id = ? AND user_id = ?", req.ActivityId, req.UserId).First(&existing).Error == nil {
			return &order.CreateOrderResp{
				Code:    200,
				Msg:     "订单已存在",
				OrderId: &existing.ID,
			}, nil
		}
		return &order.CreateOrderResp{Code: 5000, Msg: "创建订单失败"}, nil
	}

	// 缓存redis
	orderKey := fmt.Sprintf("seckill:order:%s:%d", req.UserId, req.ActivityId)
	dal.RDB.Set(s.ctx, orderKey, "SUCCESS", 24*time.Hour)

	return &order.CreateOrderResp{
		Code:    200,
		Msg:     "创建成功",
		OrderId: &o.ID,
	}, nil
}

func (s *OrderService) QueryOrder(req *order.QueryOrderReq) (*order.QueryOrderResp, error) {
	if req.OrderId != nil && *req.OrderId > 0 {
		var o model.Order
		if err := dal.DB.First(&o, *req.OrderId).Error; err != nil {
			return &order.QueryOrderResp{Code: 4002, Msg: "订单不存在"}, nil
		}
		return &order.QueryOrderResp{
			Code: 200,
			Msg:  "ok",
			Data: &order.Order{
				Id:         o.ID,
				ActivityId: o.ActivityId,
				UserId:     o.UserId,
				Status:     o.Status,
				CreatedAt:  o.CreatedAt.Format(time.RFC3339),
			},
		}, nil
	}

	if req.ActivityId == nil || req.UserId == nil {
		return &order.QueryOrderResp{Code: 4001, Msg: "参数错误"}, nil
	}

	orderKey := fmt.Sprintf("seckill:order:%s:%d", *req.UserId, *req.ActivityId)

	// 先查 Redis
	_, err := dal.RDB.Get(s.ctx, orderKey).Result()
	if err == nil {
		var o model.Order
		if err := dal.DB.Where("activity_id = ? AND user_id = ?", *req.ActivityId, *req.UserId).First(&o).Error; err != nil {
			return &order.QueryOrderResp{Code: 4002, Msg: "订单不存在"}, nil
		}
		return &order.QueryOrderResp{
			Code: 200,
			Msg:  "ok",
			Data: &order.Order{
				Id:         o.ID,
				ActivityId: o.ActivityId,
				UserId:     o.UserId,
				Status:     o.Status,
				CreatedAt:  o.CreatedAt.Format(time.RFC3339),
			},
		}, nil
	}

	//if Redis 未命中，查 MySQL
	var o model.Order
	if err := dal.DB.Where("activity_id = ? AND user_id = ?", *req.ActivityId, *req.UserId).First(&o).Error; err != nil {
		return &order.QueryOrderResp{Code: 4002, Msg: "订单不存在"}, nil
	}

	//回填 Redis
	dal.RDB.Set(s.ctx, orderKey, o.Status, 24*time.Hour)

	return &order.QueryOrderResp{
		Code: 200,
		Msg:  "ok",
		Data: &order.Order{
			Id:         o.ID,
			ActivityId: o.ActivityId,
			UserId:     o.UserId,
			Status:     o.Status,
			CreatedAt:  o.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}
