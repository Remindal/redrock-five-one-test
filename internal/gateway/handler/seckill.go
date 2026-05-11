package handler

import (
	"context"
	"seckill-system/internal/gateway/rpc"
	"seckill-system/kitex_gen/seckill"
	"seckill-system/pkg/errno"

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
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": errno.ErrParam.Code, "msg": errno.ErrParam.Msg})
		return
	}

	// JWT 鉴权：优先从 token 解析的上下文取 user_id，防止伪造
	if uid, exists := c.Get("user_id"); exists {
		req.UserId = uid.(string)
	}

	resp, err := rpc.SeckillClient.DoSeckill(ctx, &seckill.DoSeckillReq{
		ActivityId: req.ActivityId,
		UserId:     req.UserId,
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": errno.ErrInternal.Code, "msg": errno.ErrInternal.Msg})
		return
	}

	c.JSON(consts.StatusOK, resp)
}
