package handler

import (
	"context"
	"gateway/rpc"
	"strconv"

	"order/kitex_gen/order"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func QueryOrder(ctx context.Context, c *app.RequestContext) {
	activityIdStr := c.Query("activity_id")
	userId := c.Query("user_id")
	orderIdStr := c.Query("order_id")

	var req order.QueryOrderReq

	if orderIdStr != "" {
		if id, err := strconv.ParseInt(orderIdStr, 10, 64); err == nil {
			req.OrderId = &id
		}
	}
	if activityIdStr != "" && userId != "" {
		if id, err := strconv.ParseInt(activityIdStr, 10, 64); err == nil {
			req.ActivityId = &id
			req.UserId = &userId
		}
	}

	resp, err := rpc.OrderClient.QueryOrder(ctx, &req)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 5000, "msg": "系统错误"})
		return
	}

	c.JSON(consts.StatusOK, resp)
}
