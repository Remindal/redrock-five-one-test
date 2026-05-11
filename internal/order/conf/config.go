package conf

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`

	Mysql struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`

	MQ struct {
		Type           string `yaml:"type"`
		RedisStreamKey string `yaml:"redis_stream_key"`
		RedisAddr      string `yaml:"redis_addr"`
	} `yaml:"mq"`

	Etcd struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"etcd"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	RabbitMQ struct {
		Addr string `yaml:"addr"`
	} `yaml:"rabbitmq"`
}

func LoadConfig() *Config {
	data, err := os.ReadFile("configs/order.yaml")
	if err != nil {
		log.Fatalf("read config error: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("unmarshal config error: %v", err)
	}
	return &cfg
}
