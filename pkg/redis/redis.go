package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr          string   `yaml:"addr"`
	SentinelAddrs []string `yaml:"sentinel_addrs"`
	MasterName    string   `yaml:"master_name"`
	Password      string   `yaml:"password"`
	DB            int      `yaml:"db"`
}

func NewClient(cfg *Config) *redis.Client {
	if len(cfg.SentinelAddrs) > 0 {
		client := redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.MasterName,
			SentinelAddrs: cfg.SentinelAddrs,
			Password:      cfg.Password,
			DB:            cfg.DB,
			PoolSize:      500,
			MinIdleConns:  10,
			DialTimeout:   500 * time.Millisecond,
			ReadTimeout:   500 * time.Millisecond,
			WriteTimeout:  500 * time.Millisecond,
			PoolTimeout:   500 * time.Millisecond,
		})
		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Fatalf("redis sentinel connect error: %v", err)
		}
		log.Println("redis sentinel connected")
		return client
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     500,
		MinIdleConns: 10,
		DialTimeout:  500 * time.Millisecond,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		PoolTimeout:  500 * time.Millisecond,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis connect error: %v", err)
	}
	log.Println("redis connected")
	return client
}
