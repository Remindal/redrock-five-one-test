package conf

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host  string `yaml:"host"`
	Port  string `yaml:"port"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	Etcd struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"etcd"`
	Rpc struct {
		Activity string `yaml:"activity"`
		Seckill  string `yaml:"seckill"`
		Order    string `yaml:"order"`
	} `yaml:"rpc"`
}

func LoadConfig() *Config {
	data, err := os.ReadFile("conf/gateway.yaml")
	if err != nil {
		log.Fatalf("read config error: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("unmarshal config error: %v", err)
	}
	return &cfg
}
