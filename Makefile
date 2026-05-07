.PHONY: infra build run stop down clean bench test

infra:
	docker-compose up -d

build:
	@mkdir -p bin
	cd activity && go build -o ../bin/activity.exe .
	cd seckill && go build -o ../bin/seckill.exe .
	cd order && go build -o ../bin/order.exe .
	cd gateway && go build -o ../bin/gateway.exe .
	cd tools && go build -o ../bin/bench.exe .
	@echo "编译完成"

run:
	@mkdir -p logs
	nohup ./bin/activity.exe > logs/activity.log 2>&1 &
	nohup ./bin/seckill.exe > logs/seckill.log 2>&1 &
	nohup ./bin/order.exe > logs/order.log 2>&1 &
	nohup ./bin/gateway.exe > logs/gateway.log 2>&1 &
	@echo "服务已启动，日志在 logs/ 目录"

stop:
	-pkill -f "bin/activity.exe" || true
	-pkill -f "bin/seckill.exe" || true
	-pkill -f "bin/order.exe" || true
	-pkill -f "bin/gateway.exe" || true
	@echo "服务已停止"

down:
	docker-compose down

clean:
	-rm -rf bin/ logs/
	docker-compose down -v

bench:
	cd tools && go run bench.go

test:
	@echo "登录..."
	@TOKEN=$$(curl -s -X POST http://127.0.0.1:8888/api/auth/login -H "Content-Type: application/json" -d '{"user_id":"test_user"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	echo "创建活动..."; \
	curl -s -X POST http://127.0.0.1:8888/api/activity/create -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" -d '{"name":"Test","stock":100,"start_time":"2024-01-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z"}'; \
	echo ""; \
	echo "秒杀..."; \
	curl -s -X POST http://127.0.0.1:8888/api/seckill/do -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" -d '{"activity_id":1}'; \
	echo ""; \
	echo "查询订单..."; \
	curl -s "http://127.0.0.1:8888/api/order/query?activity_id=1" -H "Authorization: Bearer $$TOKEN"; \
	echo ""

help:
	@echo "可用命令:"
	@echo "  make infra  - 启动 Docker"
	@echo "  make build  - 编译服务"
	@echo "  make run    - 启动服务"
	@echo "  make stop   - 停止服务"
	@echo "  make down   - 停止 Docker"
	@echo "  make clean  - 清理"
	@echo "  make bench  - 压测"
	@echo "  make test   - 接口测试"
