# 秒杀系统（Seckill System）

基于 Kitex + Hertz + RabbitMQ + etcd + MySQL + Redis Sentinel + JWT 的微服务秒杀系统。

## 技术栈

| 层级 | 技术 | 作用 |
|------|------|------|
| API 网关 | Hertz (CloudWeGo) | HTTP 入口、限流、CORS、路由转发 |
| RPC 框架 | Kitex (CloudWeGo) | 微服务间通信 |
| 服务注册 | etcd | 服务注册与发现 |
| 缓存 | Redis Sentinel | 高可用缓存、原子扣减库存、防重 |
| 消息队列 | RabbitMQ | 异步削峰、订单异步创建 |
| 数据库 | MySQL | 活动信息、订单持久化 |
| 鉴权 | JWT (HS256) | Gateway 统一鉴权，user_id 自动注入 |
| 开发语言 | Go 1.25 | 后端服务 |
| 部署 | Docker Compose | 13 容器一键启动 |

## 项目结构

```
├── cmd/                    服务入口（activity / gateway / order / seckill）
├── internal/               各服务内部代码
│   ├── gateway/            API 网关（HTTP 8888）
│   │   ├── handler/        HTTP 处理器
│   │   ├── middleware/     限流、Recovery、JWT 鉴权、CORS
│   │   ├── dao/            Redis 连接
│   │   ├── rpc/            RPC 客户端初始化
│   │   └── conf/           配置读取
│   ├── activity/           活动服务（RPC 8081）
│   ├── seckill/            秒杀服务（RPC 8082）
│   └── order/              订单服务（RPC 8083）
├── pkg/                    公共包
│   ├── errno/              统一错误码
│   ├── logger/             统一日志
│   ├── jwt/                JWT 签发与解析
│   └── redis/              Redis 客户端封装（支持 Sentinel）
├── configs/                各服务配置文件
├── deployments/            部署配置
│   ├── docker-compose.yaml Docker Compose 编排
│   └── mysql/init.sql      数据库初始化脚本
├── scripts/                压测脚本
│   └── bench.go
├── web/                    前端演示页面
│   └── index.html
├── kitex_gen/              Thrift 生成的 RPC 代码
└── api/idl/                Thrift IDL 定义
```

## 快速启动

### Docker Compose 一键部署（推荐）

```bash
make up
```

或手动：

```bash
cd deployments
docker-compose up -d --build
```

启动 13 个容器：
- **基础设施**：MySQL(3307)、Redis Sentinel(6379/26379)、etcd(2379)、RabbitMQ(5672/15672)
- **业务服务**：Gateway(8888)、Activity(8081)、Seckill(8082)、Order(8083)

> 首次启动时，MySQL 会自动执行 `deployments/mysql/init.sql` 建表。

停止：
```bash
make down
```

查看日志：
```bash
make logs
```

### 本地开发（可选）

```bash
# 1. 先启动基础设施
cd deployments && docker-compose up -d mysql redis-master redis-sentinel-1 etcd rabbitmq

# 2. 编译并运行服务（4 个窗口）
go run ./cmd/activity
go run ./cmd/seckill
go run ./cmd/order
go run ./cmd/gateway
```

---

## 前端演示

项目提供 `web/index.html`，双击即可在浏览器中打开。

**功能：**
- 登录获取 JWT Token（localStorage 持久化）
- 实时检测 Gateway 在线状态
- 创建活动 → 秒杀下单 → 查询订单，一步操作
- 交互反馈：成功（绿色）、库存不足/重复（橙色）、错误（红色）
- 所有请求自动携带 `Authorization: Bearer` Header

**使用方式：**
1. 启动服务（`make up`）
2. 双击打开 `web/index.html`
3. 输入 user_id 登录获取 Token
4. 点击按钮完成全流程测试

---

## 接口文档

### 登录

```http
POST /api/auth/login
Content-Type: application/json
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | string | 是 | 用户 ID |

**请求示例：**
```json
{"user_id": "user_001"}
```

**响应示例：**
```json
{"code": 200, "msg": "成功", "token": "eyJhbGciOiJIUzI1NiIs..."}
```

> 登录接口**不需要** Authorization Header。其他业务接口需在 Header 中携带 `Authorization: Bearer {token}`。

---

### 创建活动

```http
POST /api/activity/create
Content-Type: application/json
Authorization: Bearer {token}
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 活动名称 |
| stock | int32 | 是 | 库存数量（>0） |
| start_time | string | 是 | 开始时间，RFC3339 格式 |
| end_time | string | 是 | 结束时间，RFC3339 格式 |

**请求示例：**
```json
{
  "name": "五一秒杀",
  "stock": 100,
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2026-12-31T23:59:59Z"
}
```

**响应示例：**
```json
{"code": 200, "msg": "创建成功", "activity_id": 1}
```

---

### 秒杀下单

```http
POST /api/seckill/do
Content-Type: application/json
Authorization: Bearer {token}
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int64 | 是 | 活动 ID |

**请求示例：**
```json
{"activity_id": 1}
```

**响应示例：**
```json
{"code": 200, "msg": "抢购成功", "status": "PROCESSING"}
```

---

### 查询订单

```http
GET /api/order/query?activity_id=1
Authorization: Bearer {token}
```

**响应示例：**
```json
{"code": 200, "msg": "ok", "data": {"id": 1, "activity_id": 1, "user_id": "user_001", "status": "SUCCESS"}}
```

---

## 错误码表

| 错误码 | 含义 | 触发场景 |
|--------|------|---------|
| 200 | 成功 | 请求处理成功 |
| 4001 | 参数错误 | 库存<=0、时间格式错误 |
| 4002 | 记录不存在 | 活动不存在、订单不存在 |
| 4003 | 活动未开始/已结束 | 当前时间不在活动有效期内 |
| 4005 | 库存不足 | Redis 库存已扣减至 0 |
| 4006 | 请勿重复抢购 | 同一用户对同一活动重复秒杀 |
| 4007 | 系统繁忙 | MQ 发送失败，库存已回滚 |
| 4008 | 请求过于频繁 | Gateway 限流触发 |
| 4009 | 未登录/Token 无效 | JWT 鉴权失败 |
| 5000 | 系统错误 | 内部异常 |

---

## 压测

```bash
cd scripts
go run bench.go
```

默认配置：并发 5000，持续 10 秒，活动 ID 8。

压测结束后自动执行一致性检查：
- Redis 库存核对
- 数据库订单数核对
- 重复用户检查

---

## 关键点

1. **Redis Sentinel 高可用**：1 主 2 从 3 哨兵，自动故障转移，客户端通过 Sentinel 自动发现主节点
2. **Lua 原子脚本**：`seckill:stock` + `seckill:users` 原子扣减，防超卖、防重复
3. **JWT 统一鉴权**：Gateway 层签发解析，RPC 层自动注入 user_id，登录接口免鉴权
4. **限流保护**：Redis 固定窗口 10,000 QPS，登录接口豁免
5. **MQ 异步削峰**：秒杀成功后发 MQ，Order 服务 5 个并发 worker 消费，降低数据库压力
6. **CORS 跨域**：支持前端直连，自动携带 Authorization Header
7. **统一错误码 & 日志**：`pkg/errno` + `pkg/logger` 全服务接入
8. **Docker 一键部署**：`make up` 启动 13 容器，考官零配置运行

---

## 注意事项

1. **启动顺序**：`docker-compose up -d --build` 会自动处理依赖顺序
2. **首次建表**：如果 MySQL 容器不是首次创建，不会自动执行 init.sql，需 `docker-compose down -v` 后重新启动
3. **Redis 数据**：容器重启后 Redis 数据会清空，但系统会从 MySQL 自动回填
