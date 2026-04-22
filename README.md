# Mall Kratos — 微服务商城系统

基于 [go-kratos](https://go-kratos.dev/) v2 微服务框架构建的分布式商城系统，采用领域驱动设计（DDD）四层架构。

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway（待扩展）                     │
└────────────────────────┬────────────────────────────────────┘
                         │ gRPC
    ┌───────────┬────────┴────────┬───────────┬──────────┐
    ▼           ▼                 ▼           ▼          ▼
user-service  cart-service   order-service  payment  promotion
    │           │                 │           │          │
    │           ▼                 │           │          │
    │       product-service ◄─────┘           │          │
    │           │                 │           │          │
    └───────────┴─────────────────┴───────────┴──────────┘
                         │
                    inventory-service
```

## 技术栈

| 组件 | 选型 |
|------|------|
| 框架 | [go-kratos](https://go-kratos.dev/) v2.9.2 |
| 语言 | Go 1.25 |
| ORM | GORM + MySQL 8.0 |
| 依赖注入 | [google/wire](https://github.com/google/wire) |
| 服务发现 | etcd |
| 消息队列 | Kafka |
| 鉴权 | JWT (golang-jwt v5) |
| 日志 | zap (Kratos 内置) |
| 配置 | Kratos config (YAML + etcd 远程配置) |

## 服务列表

| 服务 | 端口 | 数据库 | 职责 |
|------|------|--------|------|
| **user-service** | `:9004` | `mall_user` | 用户注册/登录、收货地址管理 |
| **product-service** | `:9001` | `mall_product` | 商品管理、分类管理、创建商品时初始化库存 |
| **cart-service** | `:9003` | `mall_cart` | 购物车管理（调用 product-service 校验商品） |
| **order-service** | `:9002` | `mall_order` | 订单创建/查询/取消（跨服务编排：地址→商品→库存→购物车） |
| **payment-service** | `:9006` | `mall_payment` | 支付单管理、支付回调、退款 |
| **inventory-service** | `:9007` | `mall_inventory` | 库存管理：入库/出库/锁定/解锁/确认扣减 |
| **promotion-service** | `:9005` | `mall_promotion` | 优惠券管理、发放、核销、折扣计算 |

## 快速开始

### 前置要求

- Go 1.25+
- Docker & Docker Compose
- protoc + protoc-gen-go + protoc-gen-go-grpc
- wire (`go install github.com/google/wire/cmd/wire@latest`)

### 1. 启动基础设施

```bash
cp .env.example .env
docker-compose up -d
```

这将启动：
- **MySQL** (端口 3306) — 自动执行 `migrations/` 下的建表脚本
- **etcd** (端口 2379) — 服务注册与发现
- **Kafka** (端口 9092) — 事件驱动消息队列

### 2. 初始化数据库

Docker Compose 已挂载 `migrations/` 目录，MySQL 启动时会自动执行。如需手动执行：

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot < migrations/001_create_cart_table.sql
mysql -h 127.0.0.1 -P 3306 -u root -proot < migrations/002_create_product_tables.sql
# ... 依次执行 003-007
```

### 3. 启动各微服务

每个服务一个终端：

```bash
# 终端 1: 用户服务
cd user-service && make run

# 终端 2: 商品服务
cd product-service && make run

# 终端 3: 库存服务
cd inventory-service && make run

# 终端 4: 购物车服务
cd cart-service && make run

# 终端 5: 订单服务
cd order-service && make run

# 终端 6: 支付服务
cd payment-service && make run

# 终端 7: 促销服务
cd promotion-service && make run
```

### 4. 修改配置

每个服务的配置文件位于 `configs/config.yaml`，主要修改项：

```yaml
data:
  database:
    source: "root:root@tcp(127.0.0.1:3306)/mall_xxx?..."
auth:
  jwt_secret: "your-secret-key"  # 所有服务保持一致
```

## 项目结构

每个服务遵循统一的分层架构：

```
service-name/
├── api/                    # Protobuf 定义 + 生成的 gRPC 代码
├── cmd/                    # 入口 + Wire 依赖注入
│   └── <service>/
│       ├── main.go         # 配置加载 + Wire 调用
│       ├── wire.go         # Wire 注入声明
│       └── wire_gen.go     # Wire 自动生成（不手动编辑）
├── configs/
│   └── config.yaml         # 配置文件
├── internal/
│   ├── conf/               # 配置结构体 + 错误码
│   ├── model/              # GORM 数据模型
│   ├── data/               # 数据访问层（Repository 实现 + 远程客户端）
│   ├── biz/                # 业务逻辑层（UseCase）
│   ├── service/            # gRPC Handler 层
│   ├── server/             # gRPC 服务器配置
│   └── middleware/         # 自定义中间件（JWT 鉴权等）
├── migrations/             # SQL 迁移脚本（在根目录统一管理）
├── go.mod
├── go.sum
└── Makefile
```

## 核心业务流程

### 下单流程

```
Client → order-service.CreateOrder
  ├─→ user-service.GetAddress        (校验收货地址)
  ├─→ product-service.GetProduct     (校验商品、获取价格)
  ├─→ inventory-service.LockStock    (锁定库存)
  ├─→ cart-service.ClearCart         (清空购物车)
  └─→ order-service 创建订单
```

### 支付流程

```
Client → payment-service.CreatePayment
  └─→ 生成支付单（状态: Pending）

第三方回调 → payment-service.PaymentCallback
  ├─→ 更新支付状态 (Pending → Success)
  ├─→ 通知 order-service 更新订单状态
  └─→ 调用 inventory-service.ConfirmStock 确认扣减库存
```

### 取消订单

```
Client → order-service.CancelOrder
  ├─→ 如果已支付 → payment-service.Refund (退款)
  └─→ inventory-service.UnlockStock (解锁库存)
```

## gRPC API 概览

### user-service

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Register` | `RegisterRequest` | `UserReply` | 用户注册（bcrypt 加密） |
| `Login` | `LoginRequest` | `LoginReply` | 登录（返回 JWT Token） |
| `GetUserInfo` | `GetUserInfoRequest` | `UserReply` | 获取用户信息 |
| `AddAddress` | `AddAddressRequest` | `AddressReply` | 添加收货地址 |
| `ListAddresses` | `ListAddressesRequest` | `ListAddressesReply` | 获取用户地址列表 |

### product-service

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `CreateProduct` | `CreateProductRequest` | `ProductReply` | 创建商品（同时初始化库存） |
| `GetProduct` | `GetProductRequest` | `ProductReply` | 查询商品详情 |
| `ListProducts` | `ListProductsRequest` | `ListProductsReply` | 分页查询商品列表 |
| `CreateCategory` | `CreateCategoryRequest` | `CategoryReply` | 创建分类 |
| `ListCategories` | `ListCategoriesRequest` | `ListCategoriesReply` | 查询分类列表 |

### cart-service

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `AddCart` | `AddCartRequest` | `CartReply` | 加入购物车（自动校验商品） |
| `UpdateCart` | `UpdateCartRequest` | `CartReply` | 更新购物车项 |
| `DeleteCart` | `DeleteCartRequest` | `Empty` | 删除购物车项 |
| `ListCart` | `ListCartRequest` | `ListCartReply` | 查询用户购物车 |
| `ClearCart` | `ClearCartRequest` | `Empty` | 清空购物车 |

### order-service

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `CreateOrder` | `CreateOrderRequest` | `CreateOrderResponse` | 创建订单（完整下单流程） |
| `GetOrder` | `GetOrderRequest` | `GetOrderResponse` | 查询订单详情 |
| `ListOrders` | `ListOrdersRequest` | `ListOrdersResponse` | 查询订单列表 |
| `CancelOrder` | `CancelOrderRequest` | `CancelOrderResponse` | 取消订单 |
| `UpdateOrderStatus` | `UpdateOrderStatusRequest` | `UpdateOrderStatusResponse` | 更新订单状态 |

### payment-service

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `CreatePayment` | `CreatePaymentRequest` | `PaymentReply` | 创建支付单 |
| `GetPayment` | `GetPaymentRequest` | `PaymentReply` | 查询支付单 |
| `PaymentCallback` | `PaymentCallbackRequest` | `Empty` | 第三方支付回调 |
| `Refund` | `RefundRequest` | `PaymentReply` | 退款 |

### inventory-service

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `CreateInventory` | `CreateInventoryRequest` | `CreateInventoryReply` | 创建库存记录 |
| `GetInventory` | `GetInventoryRequest` | `GetInventoryReply` | 查询库存 |
| `LockStock` | `LockStockRequest` | `LockStockReply` | 锁定库存（下单时调用） |
| `UnlockStock` | `UnlockStockRequest` | `UnlockStockReply` | 解锁库存（取消订单时调用） |
| `ConfirmStock` | `ConfirmStockRequest` | `ConfirmStockReply` | 确认扣减（支付成功后调用） |
| `StockIn` / `StockOut` | `StockIn/OutRequest` | `StockIn/OutReply` | 入库/出库操作 |

### promotion-service

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `CreateCoupon` | `CreateCouponRequest` | `CouponReply` | 创建优惠券 |
| `GrantCoupon` | `GrantCouponRequest` | `Empty` | 发放优惠券给用户 |
| `UseCoupon` | `UseCouponRequest` | `Empty` | 核销优惠券 |
| `CalculateDiscount` | `CalculateDiscountRequest` | `CalculateDiscountReply` | 计算优惠金额 |

## 鉴权

所有 gRPC 接口均需要 JWT 鉴权。客户端在请求 metadata 中携带 `Authorization: Bearer <token>`。

Token 由 `user-service.Login` 方法返回，包含以下 claims：
- `user_id` — 用户 ID
- `username` — 用户名
- `role` — 角色（user / admin）

部分服务的只读接口（如 `product-service.GetProduct`、`inventory-service.GetInventory`）允许无 token 访问。

## 错误码规范

各服务使用 Kratos `errors` 包定义业务错误码：

| 错误码 | 说明 |
|--------|------|
| 40100 | UNAUTHORIZED — 未授权 |
| 40300 | FORBIDDEN — 权限不足 |
| 40401 | CART_ITEM_NOT_FOUND |
| 40402 | PRODUCT_NOT_FOUND |
| 40403 | ORDER_NOT_FOUND |
| 40404 | PAYMENT_NOT_FOUND |
| 40405 | INVENTORY_NOT_FOUND |
| 40406 | COUPON_NOT_FOUND |
| 40001 | QUANTITY_EXCEEDS_LIMIT |
| 40002 | COUPON_EXHAUSTED |
| 40003 | INSUFFICIENT_STOCK |
| 50000 | 内部错误 |

## 开发指南

### 修改 Proto 文件

```bash
cd <service> && make protobuf
```

### 修改依赖注入

编辑 `internal/biz/biz.go` 或 `internal/data/data.go` 的 `ProviderSet`，然后重新生成 Wire：

```bash
cd <service>/cmd/<service> && wire
```

### 运行测试

```bash
cd <service> && make test
```

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- 提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org/)

## License

MIT
