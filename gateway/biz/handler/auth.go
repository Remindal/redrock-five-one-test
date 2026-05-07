package handler

import (
	"context"
	"gateway/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Login(ctx context.Context, c *app.RequestContext) {
	var req struct {
		UserId string `json:"user_id"`
	}
	if err := c.BindJSON(&req); err != nil || req.UserId == "" {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"code": 4001,
			"msg":  "参数错误，需要 user_id",
		})
		return
	}

	token, err := jwt.GenerateToken(req.UserId)
	if err != nil {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"code": 5000,
			"msg":  "签发 Token 失败",
		})
		return
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code":  200,
		"msg":   "登录成功",
		"token": token,
	})
}
