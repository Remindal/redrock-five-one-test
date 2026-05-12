# 秒杀系统架构设计

## 1. 项目架构图

```
用户请求
    │
    ▼
┌─────────────────────────────────────┐
│  Gateway（Hertz）:8888               │  ← HTTP 入口、限流、CORS、JWT 鉴权
│  - Recovery 中间件                   │
│  - RateLimit 限流中间件              │
│  - JWTAuth 鉴权中间件                │
│  - CORS 跨域中间件                   │
└──────────────┬──────────────────────┘
               │ Kitex RPC
               ▼
┌──────────────┬──────────────┬──────────────┐
│  Activity    │   Seckill    │    Order     │
│  :8081       │   :8082      │   :8083      │
│  活动管理     │  秒杀核心     │  订单处理    │
└──────────────┴──────────────┴──────────────┘
               │              │
               ▼              ▼
        ┌─────────────┐  ┌─────────────┐
        │Redis Sentinel│  │  RabbitMQ   │
        │ 1主2从3哨兵  │  │   :5672     │
        │ 库存/防重/缓存│  │  异步队列    │
        └─────────────┘  └──────┬──────┘
                                │
                                ▼
                         ┌─────────────┐
                         │   MySQL     │
                         │   :3307     │
                         │活动/订单持久化│
                         └─────────────┘
```

**服务职责：**

| 服务 | 端口 | 职责 |
|------|------|------|
| Gateway | 8888 | HTTP 统一入口，限流，CORS，路由转发到下游 RPC |
| Activity | 8081 | 创建活动、查询活动，Redis 预热库存 |
| Seckill | 8082 | 秒杀核心：时间校验 → Lua 扣库存 → 发 MQ |
| Order | 8083 | MQ 消费（5 并发 worker）→ 创建订单 → 查询订单 |

**服务注册发现**：4 个服务启动时向 etcd（:2379）注册，Gateway 通过 etcd 解析目标地址。

**容器网络**：Docker Compose 统一编排 13 个容器，服务间通过容器名通信（`activity`/`seckill`/`order`/`mysql`/`redis-master` 等）。

---

## 2. 项目结构

从扁平多模块重构为**单模块 monorepo**，采用 Go 社区标准布局：

```
cmd/              4 个服务的 main.go 入口
internal/         各服务私有代码（不可被外部导入）
  ├── gateway/    API 网关
  ├── activity/   活动服务
  ├── seckill/    秒杀服务
  └── order/      订单服务
pkg/              公共包（可被全项目导入）
  ├── errno/      统一错误码
  ├── logger/     统一日志初始化
  ├── jwt/        JWT 签发与解析
  └── redis/      Redis 客户端封装（支持 Sentinel/直连自动切换）
configs/          4 个服务的 YAML 配置文件
deployments/      Docker Compose 编排 + MySQL 初始化脚本
scripts/          压测脚本（含一致性自动验证）
web/              前端演示页面
kitex_gen/        Thrift 生成的 RPC 代码
```

**设计意图**：
- `cmd/` 只放 `main.go`，业务逻辑下沉到 `internal/`
- `pkg/` 放通用能力，避免 4 个服务重复写同一段 Redis 初始化代码
- 单模块管理：一个 `go.mod`，Kitex 生成的代码全项目共享

---

## 3. 高并发怎么解决（四层防护）

秒杀系统采用**分层过滤**思想，每一层挡住一部分流量，最终到达数据库的只有能成功下单的请求。

### 第一层：Gateway 限流（接口级）

- **实现**：Redis 固定窗口计数器，每接口每秒 10,000 QPS
- **作用**：恶意刷接口、异常流量在入口就被拒绝，保护下游服务不被冲垮
- **返回**：`code: 4008, msg: 请求过于频繁`

### 第二层：活动时间校验（业务过滤）

- **实现**：Seckill 服务从 Redis HGetAll 读取活动起止时间
- **作用**：活动未开始或已结束的请求直接拒绝，不进入库存逻辑
- **返回**：`code: 4003`

### 第三层：Redis 原子扣减（核心抗量）

- **实现**：Lua 脚本一次性完成：
  1. `SISMEMBER` 检查用户是否已下单（防重）
  2. `GET` 查询库存
  3. `DECR` 扣减库存
  4. `SADD` 记录用户已下单
- **为什么用 Lua**：Redis 单线程执行 Lua，天然原子性，数万并发同时扣减也不会出现竞态条件
- **作用**：把 MySQL 承受的压力转移到 Redis，Redis 可以轻松扛住几万 QPS

### 第四层：MySQL 唯一索引（最终兜底）

- **实现**：`order` 表有 `uk_activity_user(activity_id, user_id)` 唯一索引
- **作用**：如果极端情况下 Redis 防重失效，MySQL 层面保证同一用户同一活动只能有一条订单

---

## 4. 超卖怎么解决（四重保障）

超卖的本质是**库存扣减不是原子操作**。本系统从四个层面保证不超卖：

### ① Redis Lua 原子脚本（主防线）

```lua
local stock = tonumber(redis.call('get', stockKey))
if stock == nil or stock <= 0 then
    return 0   -- 库存不足
end
redis.call('decr', stockKey)
```

`GET` 和 `DECR` 在同一个 Lua 脚本里执行，Redis 保证这段脚本执行期间不会有其他命令插队。

### ② Redis Set 防重（防止重复扣减）

Lua 脚本先执行 `SISMEMBER`，如果用户已经下单，直接返回 `-1`。避免用户连续点击导致库存被重复扣减。

### ③ MQ 失败回滚（避免少卖）

Seckill 扣减 Redis 库存后，发 MQ 到 `seckill_queue`。如果 MQ 发送失败：

```go
dal.RDB.Incr(s.ctx, stockKey)         // 库存加回去
dal.RDB.SRem(s.ctx, usersKey, userId) // 用户记录删除
```

**保证：MQ 发不出去时，库存回滚，用户可以再抢，系统不少卖。**

### ④ MySQL 唯一索引（最终兜底）

即使前面三层全部失效，MySQL 的 `uk_activity_user` 唯一索引也会阻止重复建单。

---

## 5. 削峰怎么解决（RabbitMQ 异步 + 多 Worker）

秒杀场景下，数千人同时请求，如果直接写 MySQL，数据库瞬间被打爆。

### 异步流程

```
用户请求 → Gateway → Seckill（Redis扣库存+发MQ）→ 立即返回 PROCESSING
                                              ↓
                                        RabbitMQ 队列
                                              ↓
                                        Order Consumer × 5
                                              ↓
                                        INSERT MySQL
```

### 为什么能削峰？

- **同步变异步**：Seckill 只负责"抢资格"，耗时 < 10ms，立即返回。真正耗时的写库操作由 Order 后台慢慢消费。
- **MQ 缓冲**：即使数千请求同时到达，RabbitMQ 队列可以暂存消息，Order 按自己的速度消费。
- **5 个并发 Worker**：Order 服务启动 5 个独立 Channel 并发消费，提升消费吞吐量。
- **手动 ACK**：Order Consumer 处理完订单、写入 MySQL 后才 ACK。如果处理失败，Nack 重新入队，保证消息不丢。

### 幂等消费

Order Consumer 消费消息时：
1. 先查 MySQL 唯一索引，如果订单已存在，直接 ACK（幂等）
2. 插入订单
3. ACK

即使 MQ 消息重复投递，唯一索引 + 先查后插保证不会重复建单。

---

## 6. Redis Sentinel 高可用

生产环境 Redis 单点故障会导致秒杀系统完全瘫痪。本系统采用 **1 主 + 2 从 + 3 哨兵** 架构：

```
        ┌─────────────┐
        │  Sentinel 1 │
        │   :26379    │
        └──────┬──────┘
               │ 监控 + 故障转移
        ┌──────┴──────┐
        │             │
   ┌────┴────┐   ┌────┴────┐
   │  Master │   │ Sentinel│
   │  :6379  │   │   2/3   │
   └────┬────┘   └─────────┘
        │
   ┌────┴────┐
   │ Replica │
   │  1/2    │
   └─────────┘
```

### 客户端自动切换

`pkg/redis` 公共包封装了 `redis.NewFailoverClient`：

- 客户端只连 3 个 Sentinel 节点（`redis-sentinel-1/2/3:26379`）
- Sentinel 负责监控主节点健康，主节点宕机时自动选举从节点晋升
- 客户端自动重连新的主节点，业务代码零感知

### 向后兼容

配置保留 `addr` 直连字段，如果 `sentinel_addrs` 为空，自动回退到 `redis.NewClient` 直连模式。本地开发无需启动 Sentinel 也能跑。

---

## 7. 服务间怎么通信

### 技术选型：Kitex + Thrift + etcd

| 技术 | 作用 | 为什么选它 |
|------|------|-----------|
| **Kitex** | RPC 框架 | 字节跳动开源，Go 语言原生，性能高，内置多路复用 |
| **Thrift IDL** | 接口定义 | 跨语言、强类型，自动生成 Go 代码，减少手写重复代码 |
| **etcd** | 服务注册发现 | Cloud Native 标准，Kitex 原生支持，无需手写续约逻辑 |

### 通信流程

1. **Activity/Seckill/Order** 启动时，通过 `etcd.NewEtcdRegistry` 注册自己的地址
2. **Gateway** 启动时，通过 `etcd.NewEtcdResolver` 订阅服务列表
3. Gateway 收到 HTTP 请求后，通过 Kitex Client 调用对应的 RPC 服务
4. RPC 服务地址通过 `ADVERTISE_HOST` 环境变量注入容器名（`activity`/`seckill`/`order`），解决容器内服务发现

---

## 8. 部署架构

### Docker Compose 一键部署

```bash
cd deployments
docker-compose up -d --build
```

启动 **13 个容器**：
- **基础设施**：MySQL、Redis Master + 2 Replica + 3 Sentinel、etcd、RabbitMQ
- **业务服务**：Gateway、Activity、Seckill、Order

### Dockerfile 多阶段构建

```dockerfile
FROM golang:1.25-alpine AS builder
ARG SERVICE
RUN go build -o /app/server ./cmd/${SERVICE}

FROM alpine:latest
COPY --from=builder /app/server .
COPY configs/ ./configs/
RUN sed -i 's/127.0.0.1/mysql/g' configs/*.yaml
CMD ["./server"]
```

- 单 Dockerfile 模板，通过 `ARG SERVICE` 编译不同服务
- 多阶段构建，最终镜像仅 ~20MB
- `sed` 自动替换配置文件中的 `127.0.0.1` 为 Docker Compose 服务名

---

## 9. 压测结果

### 测试环境
- OS：Windows 11 + WSL2
- 部署：Docker Compose 13 容器
- 工具：`scripts/bench.go`（5000 并发，持续 10s）

### 核心数据

| 指标 | 结果 |
|------|------|
| 并发数 | 3000 ~ 5000 |
| 系统总 QPS | ~12,000 |
| 成功处理 | 与并发用户数一致（无丢失） |
| 是否超卖 | 订单数 = 成功数，无超卖 |
| 重复用户 | 无重复 |
| Redis 库存 | >= 0（无负库存） |
| MQ 消费 | 30 秒内全部消费完成 |

### 结论

- **系统稳定承载区间**：3000~4000 并发，失败/超时主要来自客户端队列堆积
- **性能天花板**：约 12,000 总 QPS，首次请求成功率 100%
- **瓶颈分析**：5000 并发时，大量重复请求被防重拦截，Order MQ 消费需要 5 并发 worker + 30s 等待才能全部处理完
- **优化方向**：RabbitMQ 集群化、MySQL 读写分离、Redis Cluster 扩展
