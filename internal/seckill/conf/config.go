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

	Redis struct {
		Addr          string   `yaml:"addr"`
		SentinelAddrs []string `yaml:"sentinel_addrs"`
		MasterName    string   `yaml:"master_name"`
		Password      string   `yaml:"password"`
		DB            int      `yaml:"db"`
	} `yaml:"redis"`

	RabbitMQ struct {
		Addr string `yaml:"addr"`
	} `yaml:"rabbitmq"`

	Etcd struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"etcd"`
}

func LoadConfig() *Config {
	data, err := os.ReadFile("configs/seckill.yaml")
	if err != nil {
		log.Fatalf("read config error: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("unmarshal config error: %v", err)
	}
	return &cfg
}
