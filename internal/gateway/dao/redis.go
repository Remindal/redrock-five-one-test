package dal

import (
	"context"
	"log"
	"seckill-system/internal/gateway/conf"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg *conf.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     200,
		MinIdleConns: 10,
		DialTimeout:  500 * time.Millisecond,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		PoolTimeout:  500 * time.Millisecond,
	})
	if err := RDB.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis connect error: %v", err)
	}
	log.Println("gateway redis connected")
}
