package service

import (
	"context"
	"fmt"
	"seckill/biz/dal"
	"seckill/kitex_gen/seckill"
	"time"
)

type SeckillService struct {
	ctx context.Context
}

func NewSeckillService(ctx context.Context) *SeckillService {
	return &SeckillService{ctx: ctx}
}

func (s *SeckillService) DoSeckill(req *seckill.DoSeckillReq) (*seckill.DoSeckillResp, error) {

	infoKey := fmt.Sprintf("seckill:info:%d", req.ActivityId)
	info, err := dal.RDB.HGetAll(s.ctx, infoKey).Result()

	//直接查mysql
	if err != nil || len(info) == 0 {

		var act struct {
			ID          int64     `gorm:"column:id"`
			Name        string    `gorm:"column:name"`
			RemainStock int32     `gorm:"column:remain_stock"`
			StartTime   time.Time `gorm:"column:start_time"`
			EndTime     time.Time `gorm:"column:end_time"`
			Status      int32     `gorm:"column:status"`
		}

		if dbErr := dal.DB.Table("activity").
			Where("id = ?", req.ActivityId).
			First(&act).Error; dbErr != nil {
			return &seckill.DoSeckillResp{Code: 4002, Msg: "活动不存在"}, nil
		}

		// 回填Redis
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

		info = map[string]string{
			"name":       act.Name,
			"start_time": act.StartTime.Format(time.RFC3339),
			"end_time":   act.EndTime.Format(time.RFC3339),
			"status":     fmt.Sprintf("%d", act.Status),
		}
	}

	//活动时间校验
	now := time.Now()
	startTime, _ := time.Parse(time.RFC3339, info["start_time"])
	endTime, _ := time.Parse(time.RFC3339, info["end_time"])

	if now.Before(startTime) {
		return &seckill.DoSeckillResp{Code: 4003, Msg: "活动未开始"}, nil
	}
	if now.After(endTime) {
		return &seckill.DoSeckillResp{Code: 4003, Msg: "活动已结束"}, nil
	}

	//Lua原子扣减
	luaResult, err := dal.EvalLuaStockDeduct(s.ctx, req.ActivityId, req.UserId)
	if err != nil {
		return &seckill.DoSeckillResp{Code: 5000, Msg: "系统错误"}, nil
	}

	switch luaResult {
	case -1:
		return &seckill.DoSeckillResp{Code: 4006, Msg: "请勿重复抢购"}, nil
	case 0:
		return &seckill.DoSeckillResp{Code: 4005, Msg: "库存不足"}, nil
	}

	//发MQ
	if err := dal.SendSeckillMessage(s.ctx, req.ActivityId, req.UserId); err != nil {
		stockKey := fmt.Sprintf("seckill:stock:%d", req.ActivityId)
		usersKey := fmt.Sprintf("seckill:users:%d", req.ActivityId)

		dal.RDB.Incr(s.ctx, stockKey)
		dal.RDB.SRem(s.ctx, usersKey, req.UserId)

		return &seckill.DoSeckillResp{Code: 4007, Msg: "系统繁忙，请稍后再试"}, nil
	}

	return &seckill.DoSeckillResp{
		Code:   200,
		Msg:    "抢购成功",
		Status: strPtr("PROCESSING"),
	}, nil
}

func strPtr(s string) *string {
	return &s
}
