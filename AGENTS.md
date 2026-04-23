# AGENTS.md — Mall Kratos 开发指南

## 项目概览

基于 [go-kratos](https://go-kratos.dev/) v2 的分布式商城微服务系统，Go 1.25，DDD 四层架构。

## 开发命令

### 启动基础设施
```bash
cp .env.example .env
docker-compose up -d          # MySQL(3306) + etcd(2379) + Kafka(9092)
```
MySQL 启动时自动执行 `migrations/` 下的建表脚本。

### 启动单个服务
```bash
cd <service> && make run      # 每个服务需独立终端
```

### 修改 Proto 后重新生成
```bash
cd <service>
make api        # 生成 api/ 下的 gRPC 代码
make config     # 生成 internal/ 下的配置结构体
make generate   # 运行 go generate（Wire 依赖注入）
make all        # 以上全部
```

### 运行测试
```bash
cd <service> && make test
```

### 代码格式化
```bash
gofmt -w .
```

## 服务架构

每个服务是独立的 Go module，通过根目录 `go.work` 组成 workspace。

| 服务 | Module 名 | 端口 | 数据库 |
|------|-----------|------|--------|
| product-service | `product-service` | :9001 | mall_product |
| order-service | `order-service` | :9002 | mall_order |
| cart-service | `cart-service` | :9003 | mall_cart |
| user-service | `user-service` | :9004 | mall_user |
| promotion-service | `promotion-service` | :9005 | mall_promotion |
| payment-service | `payment-service` | :9006 | mall_payment |
| inventory-service | `inventory-service` | :9007 | mall_inventory |

### 目录结构（每个服务相同）
```
service-name/
├── api/                    # Protobuf 定义 + 生成的 gRPC 代码（不要手动编辑）
├── cmd/<service>/          # 入口 + wire.go（依赖注入声明）
├── configs/config.yaml     # 配置（本地 YAML + 可选 etcd 远程配置）
├── internal/
│   ├── conf/               # 配置结构体 + 错误码
│   ├── model/              # GORM 数据模型
│   ├── data/               # Repository 实现 + 远程服务客户端
│   ├── biz/                # 业务逻辑（UseCase）
│   ├── service/            # gRPC Handler
│   └── server/             # gRPC 服务器配置
└── third_party/            # Protobuf 第三方定义
```

## 重要约定

### Wire 依赖注入
- `cmd/<service>/wire.go` 声明依赖，带 `//go:build wireinject` 构建标签
- `wire_gen.go` 由 `make generate` 自动生成，**不要手动编辑**
- 修改 `internal/biz/biz.go` 或 `internal/data/data.go` 的 `ProviderSet` 后需重新生成
- **注意**：user-service 未使用 Wire，采用手动 DI

### 服务间调用
- 通过 etcd 服务发现 + gRPC 调用
- 调用方在 `internal/data/` 中创建远程客户端
- 跨服务编排主要在 order-service 中（下单流程）

### 配置加载
- 先加载本地 `configs/config.yaml`
- 若配置了 `config_center`，再从 etcd 加载远程配置覆盖
- JWT Secret 需所有服务保持一致

### 鉴权
- 所有 gRPC 接口需要 JWT，通过 metadata `Authorization: Bearer <token>` 传递
- 部分只读接口（如 GetProduct、GetInventory）允许无 token 访问
- Token 由 user-service.Login 返回，claims 包含 user_id、username、role

### 数据库迁移
- 迁移脚本在根目录 `migrations/`，按序号执行
- Docker Compose 启动 MySQL 时自动执行 `docker-entrypoint-initdb.d` 中的脚本
- 新增迁移：在 `migrations/` 中添加 `NNN_description.sql`

### Kafka 事件
- order-service 使用 kafka-go 消费事件
- Kafka 配置在 `configs/config.yaml` 中

## 核心业务流程

### 下单（order-service.CreateOrder）
```
校验收货地址(user) → 校验商品/获取价格(product) → 锁定库存(inventory) → 清空购物车(cart) → 创建订单
```

### 支付回调（payment-service.PaymentCallback）
```
更新支付状态 → 通知 order-service 更新订单 → 确认扣减库存(inventory.ConfirmStock)
```

### 取消订单（order-service.CancelOrder）
```
已支付则退款(payment.Refund) → 解锁库存(inventory.UnlockStock)
```

## 常见陷阱

- `api/` 和 `cmd/*/wire_gen.go` 是生成代码，修改 proto 或 ProviderSet 后必须重新生成
- 每个服务是独立 module，import 路径使用 module 名（如 `order-service/internal/biz`）
- 使用 `go.work` workspace 时，`go mod tidy` 在各服务目录内执行，不在根目录
- 配置文件 `config.yaml` 默认路径为 `configs/config.yaml`，可通过 `-conf` 标志覆盖
