package handler

import (
	"context"
	"gateway/rpc"

	"activity/kitex_gen/activity"

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
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 4001, "msg": "参数错误"})
		return
	}

	resp, err := rpc.ActivityClient.CreateActivity(ctx, &activity.CreateActivityReq{
		Name:      req.Name,
		Stock:     req.Stock,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 5000, "msg": "系统错误"})
		return
	}

	c.JSON(consts.StatusOK, resp)
}
