# Gateway Design Document

**Date:** 2026-04-23
**Status:** Draft

## Summary

为 mall-kratos 微服务系统创建统一的 API Gateway，基于 go-kratos 框架实现 HTTP/REST 到 gRPC 的协议转换，提供统一入口、鉴权、限流、熔断等功能。

## Requirements

- 暴露 HTTP/REST API，内部通过 gRPC 调用微服务
- 网关统一处理 JWT 鉴权
- 静态路由配置
- 支持限流、熔断、重试、负载均衡
- 独立部署，不纳入 docker-compose

## Architecture

### Overall Architecture

```
客户端 (HTTP/REST)
    │
    ▼
┌─────────────────────────────────┐
│         Gateway (:8000)         │
│                                 │
│  ┌───────────────────────────┐  │
│  │    HTTP Server (Kratos)   │  │
│  └─────────────┬─────────────┘  │
│                │                │
│  ┌─────────────▼─────────────┐  │
│  │      Middleware Chain     │  │
│  │  1. Recovery              │  │
│  │  2. Logging               │  │
│  │  3. Rate Limit            │  │
│  │  4. JWT Auth              │  │
│  │  5. Circuit Breaker       │  │
│  └─────────────┬─────────────┘  │
│                │                │
│  ┌─────────────▼─────────────┐  │
│  │      Route Handler        │  │
│  │  /api/user/*     → user   │  │
│  │  /api/product/*  → product│  │
│  │  /api/order/*    → order  │  │
│  │  /api/cart/*     → cart   │  │
│  │  /api/payment/*  → payment│  │
│  │  /api/inventory/*→ inv    │  │
│  │  /api/promotion/*→ promo  │  │
│  └─────────────┬─────────────┘  │
│                │                │
│  ┌─────────────▼─────────────┐  │
│  │    gRPC Clients (etcd)    │  │
│  │  7 service clients        │  │
│  └───────────────────────────┘  │
└─────────────────────────────────┘
    │         │         │
    ▼         ▼         ▼
  user-svc  order-svc  ... (gRPC)
```

### Request Flow

```
HTTP Request → Gateway :8000
  → Middleware: Recovery (panic 恢复)
  → Middleware: Logging (请求日志)
  → Middleware: RateLimit (令牌桶限流)
  → Middleware: JWT Auth (验证 Bearer Token，提取 user_id)
  → Middleware: CircuitBreaker (熔断保护)
  → Route Handler (路由匹配)
  → gRPC Client (通过 etcd 服务发现)
  → 下游微服务
  → 响应转换为 HTTP JSON 返回
```

## Routing Design

### Route Rules

| HTTP Path Prefix | Target Service | gRPC Service |
|-----------------|----------------|-------------|
| `/api/user/v1/*` | user-service (`:9004`) | `user.v1.UserService` |
| `/api/product/v1/*` | product-service (`:9001`) | `product.v1.ProductService` |
| `/api/order/v1/*` | order-service (`:9002`) | `order.v1.OrderService` |
| `/api/cart/v1/*` | cart-service (`:9003`) | `cart.v1.CartService` |
| `/api/payment/v1/*` | payment-service (`:9006`) | `payment.v1.PaymentService` |
| `/api/inventory/v1/*` | inventory-service (`:9007`) | `inventory.v1.InventoryService` |
| `/api/promotion/v1/*` | promotion-service (`:9005`) | `promotion.v1.PromotionService` |

### Implementation

使用 Kratos HTTP Router 按前缀分发到对应 gRPC Client Handler。

## Authentication Design

### JWT Middleware

- 从 `Authorization: Bearer <token>` 提取 Token
- 使用共享 JWT Secret 验证签名
- 将 `user_id`、`username`、`role` 注入到 gRPC metadata 传递给下游

### Auth Whitelist (No Token Required)

| Endpoint | Description |
|----------|-------------|
| `POST /api/user/v1/register` | User registration |
| `POST /api/user/v1/login` | User login |
| `GET /api/product/v1/products` | Product list |
| `GET /api/product/v1/product/:id` | Product detail |
| `GET /api/inventory/v1/inventory/:id` | Inventory query |
| `GET /api/promotion/v1/promotions` | Promotion list |

## Rate Limiting & Circuit Breaker

### Rate Limit

- 使用令牌桶算法
- 全局默认：1000 req/s
- 可按路径配置不同限流阈值

### Circuit Breaker

- 使用 Kratos 内置 `circuitbreaker` middleware
- 策略：连续 5 次失败或错误率 > 50% 时熔断
- 熔断时长：3 秒后进入半开状态

## Error Handling

### Error Code Mapping

| gRPC Status | HTTP Status | Description |
|------------|-------------|-------------|
| OK | 200 | Success |
| InvalidArgument | 400 | Bad request |
| Unauthenticated | 401 | Unauthenticated |
| PermissionDenied | 403 | Forbidden |
| NotFound | 404 | Not found |
| AlreadyExists | 409 | Conflict |
| Internal | 500 | Internal error |

### Unified Error Response

```json
{
  "code": 10001,
  "message": "用户不存在",
  "reason": "USER_NOT_FOUND"
}
```

## Project Structure

```
gateway/
├── cmd/gateway/
│   ├── main.go              # Entry point
│   └── wire.go              # Dependency injection
├── configs/
│   └── config.yaml          # Configuration
├── internal/
│   ├── conf/                # Config struct
│   │   └── conf.proto
│   ├── middleware/
│   │   ├── auth/            # JWT auth middleware
│   │   └── ratelimit/       # Rate limit middleware
│   ├── service/             # HTTP handlers (routing)
│   │   └── gateway.go
│   └── server/              # HTTP server config
│       └── http.go
├── go.mod
├── Makefile
└── Dockerfile
```

## Configuration

### config.yaml

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 10s
  grpc:
    addr: 0.0.0.0:9000

data:
  etcd:
    endpoints:
      - 127.0.0.1:2379

auth:
  jwt_secret: "your-secret-key"
  whitelist:
    - POST:/api/user/v1/register
    - POST:/api/user/v1/login
    - GET:/api/product/v1/products
    - GET:/api/product/v1/product
    - GET:/api/inventory/v1/inventory
    - GET:/api/promotion/v1/promotions

rate_limit:
  global: 1000
  per_path:
    "/api/user/v1/login": 10

circuit_breaker:
  threshold: 5
  error_rate: 0.5
  recovery_time: 3s
```

## Deployment

- 独立部署，不纳入 docker-compose
- 端口：HTTP :8000
- 通过 etcd 服务发现各微服务地址
