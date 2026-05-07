package middleware

import (
	"context"
	"gateway/pkg/jwt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 跳过 CORS 预检请求
		if string(c.Method()) == "OPTIONS" {
			c.Next(ctx)
			return
		}
		auth := string(c.GetHeader("Authorization"))
		if auth == "" {
			c.AbortWithStatusJSON(consts.StatusOK, map[string]interface{}{
				"code": 4009,
				"msg":  "缺少 Authorization Header",
			})
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(consts.StatusOK, map[string]interface{}{
				"code": 4009,
				"msg":  "Authorization 格式错误，应为 Bearer {token}",
			})
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(consts.StatusOK, map[string]interface{}{
				"code": 4009,
				"msg":  "Token 无效或已过期",
			})
			return
		}

		c.Set("user_id", claims.UserId)
		c.Next(ctx)
	}
}
