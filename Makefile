.PHONY: up down logs build bench

up:
	cd deployments && docker-compose up -d --build

down:
	cd deployments && docker-compose down

logs:
	cd deployments && docker-compose logs -f

build:
	@mkdir -p bin
	go build -o bin/activity ./cmd/activity
	go build -o bin/seckill ./cmd/seckill
	go build -o bin/order ./cmd/order
	go build -o bin/gateway ./cmd/gateway
	@echo "编译完成"

bench:
	cd scripts && go run bench.go
