package handler

import (
	"context"
	"seckill-system/internal/gateway/rpc"
	"seckill-system/kitex_gen/activity"
	"seckill-system/pkg/errno"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type CreateActivityHTTPReq struct {
	Name      string `json:"name"`
	Stock     int32  `json:"stock"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func CreateActivity(ctx context.Context, c *app.RequestContext) {
	var req CreateActivityHTTPReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": errno.ErrParam.Code, "msg": errno.ErrParam.Msg})
		return
	}
	resp, err := rpc.ActivityClient.CreateActivity(ctx, &activity.CreateActivityReq{
		Name: req.Name, Stock: req.Stock, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": errno.ErrInternal.Code, "msg": errno.ErrInternal.Msg})
		return
	}
	c.JSON(consts.StatusOK, resp)
}
