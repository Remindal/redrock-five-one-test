# 秒杀系统（Seckill System）

基于 Kitex + Hertz + RabbitMQ + etcd + MySQL + Redis + JWT 的微服务秒杀系统。

## 技术栈

| 层级 | 技术 | 作用 |
|------|------|------|
| API 网关 | Hertz (CloudWeGo) | HTTP 入口、限流、路由转发 |
| RPC 框架 | Kitex (CloudWeGo) | 微服务间通信 |
| 服务注册 | etcd | 服务注册与发现 |
| 缓存 | Redis | 抗高并发、原子扣减库存、防重 |
| 消息队列 | RabbitMQ | 异步削峰、订单异步创建 |
| 数据库 | MySQL | 活动信息、订单持久化 |
| 鉴权 | JWT (HS256) | Gateway 统一鉴权，user_id 自动注入 |
| 开发语言 | Go 1.25 | 后端服务 |
| 部署 | Docker Compose | 基础设施一键启动 |

## 项目结构

```
├── idl/                    Thrift IDL 定义
├── deploy/                 部署配置
│   └── mysql/init.sql      数据库初始化脚本
├── docker-compose.yaml     基础设施编排
├── Makefile                一键命令
├── tools/
│   └── bench.go            压测脚本
├── gateway/                API 网关（HTTP 8888）
│   ├── biz/handler/        HTTP 处理器（含登录）
│   ├── biz/middleware/     限流、Recovery、JWT 鉴权
│   ├── biz/dal/            Redis 连接
│   ├── pkg/jwt/            JWT 签发与解析
│   ├── conf/               配置文件
│   └── rpc/                RPC 客户端初始化
├── activity/               活动服务（RPC 8081）
│   ├── biz/service/        活动创建、查询
│   └── biz/dal/            MySQL + Redis
├── seckill/                秒杀服务（RPC 8082）
│   ├── biz/service/        秒杀核心逻辑
│   └── biz/dal/            Redis Lua + RabbitMQ
└── order/                  订单服务（RPC 8083）
    ├── biz/service/        订单创建、查询
    └── biz/mq/             RabbitMQ Consumer
```

## 环境要求

- Go >= 1.24
- Docker Desktop
- Windows 10/11 + WSL2（推荐WSL2，压测更准确）

## 快速启动

### 1. 启动基础设施

```bash
make infra
```

启动 MySQL(3307)、Redis(6379)、etcd(2379)、RabbitMQ(5672/15672)。

> 首次启动时，MySQL 会自动执行 `deploy/mysql/init.sql` 建表。

### 2. 编译服务

```bash
make build
```

### 3. 启动服务（先 RPC 后 Gateway）

**方式 A：手动启动（Windows PowerShell，4 个窗口）**

```powershell
# 窗口1
./bin/activity.exe

# 窗口2
./bin/seckill.exe

# 窗口3
./bin/order.exe

# 窗口4
./bin/gateway.exe
```

**方式 B：后台启动（WSL/Linux）**

```bash
make run
```

### 4. 停止

```bash
make stop      # 停止服务
make down      # 停止 Docker
make clean     # 清理全部产物
```

---

## AI vibe 前端演示

项目根目录提供 `index.html`，双击即可在浏览器中打开。

**功能：**
- 登录获取 JWT Token（localStorage 持久化）
- 实时检测 Gateway 在线状态
- 创建活动 → 秒杀下单 → 查询订单，一步操作
- 交互反馈：成功（绿色）、库存不足/重复（橙色）、错误（红色）
- 创建活动后自动填充 `activity_id`
- 所有请求自动携带 `Authorization: Bearer` Header
- 无压测功能

**使用方式：**
1. 启动 4 个微服务（见上文）
2. 双击打开 `index.html`
3. 输入 user_id 登录获取 Token
4. 点击按钮完成全流程测试

> 需要 Gateway 开启 CORS（已默认开启，允许 `Authorization` Header）

## 接口文档

### 创建活动

```http
POST /api/activity/create
Content-Type: application/json
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
{
  "code": 200,
  "msg": "创建成功",
  "activity_id": 1
}
```

---

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
{
  "user_id": "user_001"
}
```

**响应示例：**

```json
{
  "code": 200,
  "msg": "登录成功",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

> 登录接口**不需要** Authorization Header。其他业务接口需在 Header 中携带 `Authorization: Bearer {token}`。

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
| user_id | string | 否 | 用户 ID（JWT 会自动覆盖，前端可不传） |

**请求示例：**

```json
{
  "activity_id": 1
}
```

**响应示例：**

```json
{
  "code": 200,
  "msg": "抢购成功",
  "status": "PROCESSING"
}
```

---

### 查询订单

```http
GET /api/order/query?activity_id=1
Authorization: Bearer {token}
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int64 | 条件必填 | 活动 ID（与 user_id 组合查询） |
| user_id | string | 否 | 用户 ID（JWT 自动注入，前端可不传） |
| order_id | int64 | 否 | 订单 ID（优先按此查） |

**响应示例：**

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": 1,
    "activity_id": 1,
    "user_id": "user_001",
    "status": "SUCCESS",
    "created_at": "2026-05-07T11:19:23Z"
  }
}
```

## 错误码表

| 错误码 | 含义 | 触发场景 |
|--------|------|---------|
| 200 | 成功 | 请求处理成功 |
| 4001 | 参数错误 | 库存<=0、时间格式错误、缺少必要参数 |
| 4002 | 记录不存在 | 活动不存在、订单不存在 |
| 4003 | 活动未开始/已结束 | 当前时间不在活动有效期内 |
| 4005 | 库存不足 | Redis 库存已扣减至 0 |
| 4006 | 请勿重复抢购 | 同一用户对同一活动重复秒杀 |
| 4007 | 系统繁忙 | MQ 发送失败，库存已回滚 |
| 4008 | 请求过于频繁 | Gateway 限流触发 |
| 4009 | 未登录/Token 无效 | 缺少 Authorization Header、格式错误、Token 过期 |
| 5000 | 系统错误 | 内部异常、数据库/Redis 故障 |

## 压测脚本

项目提供 `tools/bench.go` 压测工具，支持自定义并发数和持续时间。

```bash
cd tools
go run bench.go
```

默认配置：并发 1000，持续 10 秒，压测 `127.0.0.1:8888/api/seckill/do`。
修改 `bench.go` 中的 `concurrency`、`duration`、`activityId` 即可调整参数。

## 压测结果

| 指标 | 数据 |
|------|------|
| 并发数 | 3000 ~ 5000 |
| 持续时长 | 10s |
| 总 QPS | ~6200 |
| 成功处理 | ~4500 QPS |
| 库存 100 / 并发 1000 | 订单数=100，无重复，库存>=0 |
| 系统极限 | 约 3000~4000 稳定 QPS，超过后失败率上升 |

> 压测环境：Windows 11 + WSL2，单机部署 4 个微服务 + Docker 基础设施。
>
> WSL 环境下压测时，如果 `127.0.0.1` 无法访问 Windows 服务，请将 `tools/bench.go` 中的 URL 改为 Windows 主机的实际局域网 IP（通过 `ipconfig` 查看）。

## 架构设计

详见 [SUMMARY.md](./SUMMARY.md)

## 注意事项

1. **启动顺序**：必须先启动 RPC 服务（Activity/Seckill/Order），最后启动 Gateway
2. **首次建表**：如果 MySQL 容器不是首次创建，不会自动执行 init.sql，需手动进容器执行
3. **Redis 数据**：容器重启后 Redis 数据会清空，但系统会从 MySQL 自动回填
