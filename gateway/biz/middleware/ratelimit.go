package middleware

import (
	"context"
	"fmt"
	"gateway/biz/dal"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func RateLimit(maxQPS int64, windowSeconds int64) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.Path())
		now := time.Now().Unix()
		// 固定窗口：按时间戳取整
		windowStart := now - (now % windowSeconds)
		redisKey := fmt.Sprintf("ratelimit:%s:%d", path, windowStart)

		count, err := dal.RDB.Incr(ctx, redisKey).Result()
		if err != nil {
			c.AbortWithStatusJSON(consts.StatusOK, map[string]interface{}{
				"code": 5000,
				"msg":  "系统错误",
			})
			return
		}

		// 第一次计数，设置过期时间
		if count == 1 {
			dal.RDB.Expire(ctx, redisKey, time.Duration(windowSeconds)*time.Second)
		}

		if count > maxQPS {
			c.AbortWithStatusJSON(consts.StatusOK, map[string]interface{}{
				"code": 4008,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next(ctx)
	}
}
