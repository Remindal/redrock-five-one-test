package handler

import (
	"context"
	"seckill-system/pkg/errno"
	"seckill-system/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Login(ctx context.Context, c *app.RequestContext) {
	var req struct {
		UserId string `json:"user_id"`
	}
	if err := c.BindJSON(&req); err != nil || req.UserId == "" {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"code": errno.ErrParam.Code, "msg": errno.ErrParam.Msg,
		})
		return
	}

	token, err := jwt.GenerateToken(req.UserId)
	if err != nil {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"code": errno.ErrInternal.Code, "msg": errno.ErrInternal.Msg,
		})
		return
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": errno.OK.Code, "msg": errno.OK.Msg,
		"token": token,
	})
}
