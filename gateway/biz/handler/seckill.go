package handler

import (
	"context"
	"gateway/rpc"

	"seckill/kitex_gen/seckill"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type DoSeckillHTTPReq struct {
	ActivityId int64  `json:"activity_id"`
	UserId     string `json:"user_id"`
}

func DoSeckill(ctx context.Context, c *app.RequestContext) {
	var req DoSeckillHTTPReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 4001, "msg": "参数错误"})
		return
	}

	resp, err := rpc.SeckillClient.DoSeckill(ctx, &seckill.DoSeckillReq{
		ActivityId: req.ActivityId,
		UserId:     req.UserId,
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 5000, "msg": "系统错误"})
		return
	}

	c.JSON(consts.StatusOK, resp)
}
