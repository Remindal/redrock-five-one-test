package dal

import (
	"log"
	"seckill-system/internal/order/conf"
	redispkg "seckill-system/pkg/redis"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

func Init(cfg *conf.Config) {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.Mysql.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("mysql connect error: %v", err)
	}
	log.Println("mysql connected")

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("get sql db error: %v", err)
	}
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	log.Println("mysql pool configured")

	RDB = redispkg.NewClient(&redispkg.Config{
		Addr:          cfg.Redis.Addr,
		SentinelAddrs: cfg.Redis.SentinelAddrs,
		MasterName:    cfg.Redis.MasterName,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.DB,
	})
	log.Println("order redis init done")

	InitRabbitMQ(cfg.RabbitMQ.Addr)
}
