package dal

import (
	"log"
	"seckill-system/internal/gateway/conf"
	redispkg "seckill-system/pkg/redis"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg *conf.Config) {
	RDB = redispkg.NewClient(&redispkg.Config{
		Addr:          cfg.Redis.Addr,
		SentinelAddrs: cfg.Redis.SentinelAddrs,
		MasterName:    cfg.Redis.MasterName,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.DB,
	})
	log.Println("gateway redis init done")
}
