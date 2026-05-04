package service

import (
	"context"
	"fmt"
	"time"

	"activity/biz/dal"
	"activity/biz/model"
	"activity/kitex_gen/activity"
)

type ActivityService struct {
	ctx context.Context
}

func NewActivityService(ctx context.Context) *ActivityService {
	return &ActivityService{ctx: ctx}
}

func (s *ActivityService) CreateActivity(req *activity.CreateActivityReq) (*activity.CreateActivityResp, error) {
	if req.Stock <= 0 {
		return &activity.CreateActivityResp{Code: 4001, Msg: "库存必须大于0"}, nil
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return &activity.CreateActivityResp{Code: 4001, Msg: "开始时间格式错误"}, nil
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return &activity.CreateActivityResp{Code: 4001, Msg: "结束时间格式错误"}, nil
	}
	if endTime.Before(startTime) {
		return &activity.CreateActivityResp{Code: 4001, Msg: "结束时间必须晚于开始时间"}, nil
	}

	act := model.Activity{
		Name:        req.Name,
		RemainStock: req.Stock,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      1,
	}
	if err := dal.DB.Create(&act).Error; err != nil {
		return &activity.CreateActivityResp{Code: 5000, Msg: "数据库错误"}, nil
	}

	stockKey := fmt.Sprintf("seckill:stock:%d", act.ID)
	infoKey := fmt.Sprintf("seckill:info:%d", act.ID)

	pipe := dal.RDB.Pipeline()
	pipe.Set(s.ctx, stockKey, req.Stock, 0)
	pipe.HSet(s.ctx, infoKey, map[string]interface{}{
		"name":       req.Name,
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
		"status":     "1",
	})
	if _, err := pipe.Exec(s.ctx); err != nil {
		return &activity.CreateActivityResp{
			Code: 5000,
			Msg:  "Redis 预热失败: " + err.Error(),
		}, nil
	}

	return &activity.CreateActivityResp{
		Code:       200,
		Msg:        "创建成功",
		ActivityId: &act.ID,
	}, nil
}

func (s *ActivityService) GetActivity(req *activity.GetActivityReq) (*activity.GetActivityResp, error) {
	infoKey := fmt.Sprintf("seckill:info:%d", req.ActivityId)

	// 1. 先查 Redis
	info, err := dal.RDB.HGetAll(s.ctx, infoKey).Result()
	if err == nil && len(info) > 0 {
		stockStr, _ := dal.RDB.Get(s.ctx, fmt.Sprintf("seckill:stock:%d", req.ActivityId)).Result()
		stock := int32(0)
		fmt.Sscanf(stockStr, "%d", &stock)

		id := req.ActivityId
		return &activity.GetActivityResp{
			Code: 200,
			Msg:  "ok",
			Data: &activity.Activity{
				Id:          id,
				Name:        info["name"],
				Stock:       stock,
				RemainStock: stock,
				StartTime:   info["start_time"],
				EndTime:     info["end_time"],
				Status:      1,
			},
		}, nil
	}

	// 2. Redis 没命中，查 MySQL
	var act model.Activity
	if err := dal.DB.First(&act, req.ActivityId).Error; err != nil {
		return &activity.GetActivityResp{Code: 4002, Msg: "活动不存在"}, nil
	}

	// 3. 回填 Redis
	stockKey := fmt.Sprintf("seckill:stock:%d", act.ID)
	pipe := dal.RDB.Pipeline()
	pipe.Set(s.ctx, stockKey, act.RemainStock, 0)
	pipe.HSet(s.ctx, infoKey, map[string]interface{}{
		"name":       act.Name,
		"start_time": act.StartTime.Format(time.RFC3339),
		"end_time":   act.EndTime.Format(time.RFC3339),
		"status":     fmt.Sprintf("%d", act.Status),
	})
	pipe.Exec(s.ctx)

	return &activity.GetActivityResp{
		Code: 200,
		Msg:  "ok",
		Data: &activity.Activity{
			Id:          act.ID,
			Name:        act.Name,
			Stock:       act.Stock,
			RemainStock: act.RemainStock,
			StartTime:   act.StartTime.Format(time.RFC3339),
			EndTime:     act.EndTime.Format(time.RFC3339),
			Status:      act.Status,
		},
	}, nil
}
