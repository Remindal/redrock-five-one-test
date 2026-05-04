package dal

import (
	"activity/conf"
	"context"
	"log"

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

	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := RDB.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis connect error: %v", err)
	}
	log.Println("redis connected")
}
