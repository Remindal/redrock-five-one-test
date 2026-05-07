package dal

import (
	"context"
	"log"
	"seckill/conf"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	RDB *redis.Client
	DB  *gorm.DB
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
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	log.Println("mysql pool configured")

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
	log.Println("redis connected")

	InitRabbitMQ(cfg.RabbitMQ.Addr)
}
