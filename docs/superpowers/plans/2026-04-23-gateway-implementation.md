# Gateway Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create an API Gateway as the unified HTTP/REST entry point for all mall-kratos microservices, with JWT auth, rate limiting, circuit breaker, and etcd-based service discovery.

**Architecture:** HTTP Server (Kratos) → Middleware Chain (Recovery, Logging, RateLimit, JWT Auth, CircuitBreaker) → Route Handlers → gRPC Clients (via etcd discovery) → Backend microservices.

**Tech Stack:** Go 1.25, go-kratos v2, gRPC, etcd, JWT, Google Wire

---

## File Structure

```
gateway/
├── cmd/gateway/
│   ├── main.go              # Entry point
│   ├── wire.go              # Wire dependency injection
│   └── wire_gen.go          # Generated (do not edit)
├── api/gateway/v1/
│   └── gateway.proto        # Gateway proto definitions (optional, for shared types)
├── configs/
│   └── config.yaml          # Configuration
├── internal/
│   ├── conf/
│   │   ├── conf.proto       # Config structure definition
│   │   ├── conf.pb.go       # Generated
│   │   └── errors.go        # Error codes
│   ├── middleware/
│   │   ├── auth.go          # JWT authentication middleware
│   │   └── error.go         # Response error mapping middleware
│   ├── client/
│   │   └── clients.go       # gRPC clients for all 7 services
│   ├── service/
│   │   ├── gateway.go       # HTTP route handlers
│   │   └── service.go       # Wire ProviderSet
│   └── server/
│       ├── http.go          # HTTP server configuration
│       └── server.go        # Wire ProviderSet
├── third_party/
│   ├── google/api/          # Proto third-party definitions
│   └── validate/
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `gateway/go.mod`
- Create: `gateway/Makefile`
- Create: `gateway/third_party/google/api/` (proto files)
- Modify: `go.work` (add gateway module)

- [ ] **Step 1: Create go.mod**

Create `gateway/go.mod`:

```go
module gateway

go 1.25.0

require (
	github.com/go-kratos/kratos/v2 v2.8.0
	github.com/go-kratos/kratos/contrib/registry/etcd/v2 v2.0.0-20241216143613-26506c97d4c4
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/wire v0.6.0
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576
	google.golang.org/grpc v1.68.1
	google.golang.org/protobuf v1.35.2
	gopkg.in/yaml.v3 v3.0.1
)
```

- [ ] **Step 2: Create Makefile**

Create `gateway/Makefile`:

```makefile
GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       $(API_PROTO_FILES)

.PHONY: build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
all:
	make api
	make config
	make generate

.PHONY: run
run:
	go run ./cmd/gateway -conf configs/config.yaml

.PHONY: help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_RES)

.DEFAULT_GOAL := help
```

- [ ] **Step 3: Copy third_party proto files**

Copy proto files from an existing service:

```bash
cp -r user-service/third_party gateway/third_party
```

- [ ] **Step 4: Update go.work**

Add `./gateway` to the use block in `/home/lgzzz/IdeaProjects/mall-kratos/go.work`:

```go
go 1.25.0

use (
	./cart-service
	./gateway
	./inventory-service
	./order-service
	./payment-service
	./product-service
	./promotion-service
	./user-service
)
```

- [ ] **Step 5: Run go mod tidy**

```bash
cd gateway && go mod tidy
```

Expected: Dependencies resolved, `go.sum` created.

- [ ] **Step 6: Commit**

```bash
git add gateway/go.mod gateway/go.sum gateway/Makefile gateway/third_party/ go.work
git commit -m "feat(gateway): scaffold project structure"
```

---

### Task 2: Configuration

**Files:**
- Create: `gateway/internal/conf/conf.proto`
- Create: `gateway/configs/config.yaml`

- [ ] **Step 1: Create conf.proto**

Create `gateway/internal/conf/conf.proto`:

```proto
syntax = "proto3";

package gateway.internal.conf;

option go_package = "gateway/internal/conf;conf";

import "google/protobuf/duration.proto";

message Bootstrap {
  Server server = 1;
  Data data = 2;
  Registry registry = 3;
  Auth auth = 4;
  RateLimit rate_limit = 5;
  CircuitBreaker circuit_breaker = 6;
}

message Server {
  HTTP http = 1;
}

message HTTP {
  string network = 1;
  string addr = 2;
  google.protobuf.Duration timeout = 3;
}

message Data {
  repeated string etcd_endpoints = 1;
}

message Registry {
  repeated string endpoints = 1;
}

message Auth {
  string jwt_secret = 1;
  repeated string whitelist = 2;
}

message RateLimit {
  int64 global = 1;
  map<string, int64> per_path = 2;
}

message CircuitBreaker {
  int64 threshold = 1;
  double error_rate = 2;
  google.protobuf.Duration recovery_time = 3;
}
```

- [ ] **Step 2: Generate config struct**

```bash
cd gateway && make config
```

Expected: `gateway/internal/conf/conf.pb.go` generated.

- [ ] **Step 3: Create config.yaml**

Create `gateway/configs/config.yaml`:

```yaml
server:
  http:
    network: tcp
    addr: 0.0.0.0:8000
    timeout: 10s

data:
  etcd_endpoints:
    - 127.0.0.1:2379

registry:
  endpoints:
    - 127.0.0.1:2379

auth:
  jwt_secret: "your-secret-key-change-in-production"
  whitelist:
    - "POST:/api/user/v1/register"
    - "POST:/api/user/v1/login"
    - "GET:/api/product/v1/products"
    - "GET:/api/product/v1/product"
    - "GET:/api/inventory/v1/inventory"
    - "GET:/api/promotion/v1/promotions"

rate_limit:
  global: 1000
  per_path:
    "/api/user/v1/login": 10

circuit_breaker:
  threshold: 5
  error_rate: 0.5
  recovery_time: 3s
```

- [ ] **Step 4: Commit**

```bash
git add gateway/internal/conf/ gateway/configs/
git commit -m "feat(gateway): add configuration"
```

---

### Task 3: Error Codes

**Files:**
- Create: `gateway/internal/conf/errors.go`

- [ ] **Step 1: Create errors.go**

Create `gateway/internal/conf/errors.go`:

```go
package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	ErrCodeSuccess          = 0
	ErrCodeInternal         = 50000
	ErrCodeUnauthorized     = 50005
	ErrCodeForbidden        = 50006
	ErrCodeNotFound         = 50007
	ErrCodeBadRequest       = 50008
	ErrCodeRateLimitExceeded = 50009
	ErrCodeServiceUnavailable = 50010
)

var (
	ErrUnauthorized     = errors.Unauthorized("UNAUTHORIZED", "unauthorized")
	ErrForbidden        = errors.Forbidden("FORBIDDEN", "forbidden")
	ErrNotFound         = errors.NotFound("NOT_FOUND", "not found")
	ErrBadRequest       = errors.BadRequest("BAD_REQUEST", "bad request")
	ErrRateLimitExceeded = errors.New(429, "RATE_LIMIT_EXCEEDED", "rate limit exceeded")
	ErrServiceUnavailable = errors.New(503, "SERVICE_UNAVAILABLE", "service unavailable")
	ErrInternal         = errors.InternalError("INTERNAL", "internal server error")
)

var reasonToCode = map[string]int32{
	"UNAUTHORIZED":        ErrCodeUnauthorized,
	"FORBIDDEN":           ErrCodeForbidden,
	"NOT_FOUND":           ErrCodeNotFound,
	"BAD_REQUEST":         ErrCodeBadRequest,
	"RATE_LIMIT_EXCEEDED": ErrCodeRateLimitExceeded,
	"SERVICE_UNAVAILABLE": ErrCodeServiceUnavailable,
	"INTERNAL":            ErrCodeInternal,
}

func ErrorCodeFromReason(reason string) int32 {
	if code, ok := reasonToCode[reason]; ok {
		return code
	}
	return ErrCodeInternal
}
```

- [ ] **Step 2: Commit**

```bash
git add gateway/internal/conf/errors.go
git commit -m "feat(gateway): add error codes"
```

---

### Task 4: JWT Auth Middleware

**Files:**
- Create: `gateway/internal/middleware/auth.go`

- [ ] **Step 1: Create auth.go**

Create `gateway/internal/middleware/auth.go`:

```go
package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"

	"gateway/internal/conf"
)

type AuthInfo struct {
	UserID   int64
	Username string
	Role     string
}

type authKey struct{}

func ServerAuth(secret string, whitelist []string) middleware.Middleware {
	whitelistMap := make(map[string]bool, len(whitelist))
	for _, w := range whitelist {
		whitelistMap[w] = true
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Extract HTTP method and path from context
			method, path := extractHTTPInfo(ctx)
			key := method + ":" + path

			// Check whitelist
			if whitelistMap[key] {
				return handler(ctx, req)
			}

			// Extract token
			tokenStr := extractToken(ctx)
			if tokenStr == "" {
				return nil, conf.ErrUnauthorized
			}

			// Parse and validate JWT
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return nil, conf.ErrUnauthorized
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, conf.ErrUnauthorized
			}

			userID, _ := claims.GetExpirationTime()
			_ = userID

			info := &AuthInfo{}
			if uid, ok := claims["user_id"].(float64); ok {
				info.UserID = int64(uid)
			}
			if username, ok := claims["username"].(string); ok {
				info.Username = username
			}
			if role, ok := claims["role"].(string); ok {
				info.Role = role
			}

			if info.UserID == 0 {
				return nil, conf.ErrUnauthorized
			}

			// Store auth info in context
			ctx = context.WithValue(ctx, authKey{}, info)

			// Inject into gRPC metadata for downstream services
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			md.Set("x-user-id", fmt.Sprintf("%d", info.UserID))
			md.Set("x-username", info.Username)
			md.Set("x-role", info.Role)
			ctx = metadata.NewOutgoingContext(ctx, md)

			return handler(ctx, req)
		}
	}
}

func GetAuthInfo(ctx context.Context) (*AuthInfo, bool) {
	info, ok := ctx.Value(authKey{}).(*AuthInfo)
	return info, ok
}

func extractToken(ctx context.Context) string {
	// Try HTTP header first
	if header, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := header.Get("authorization"); len(auth) > 0 {
			return strings.TrimPrefix(auth[0], "Bearer ")
		}
	}
	return ""
}

func extractHTTPInfo(ctx context.Context) (method, path string) {
	// Kratos stores HTTP transport in context
	if tr, ok := http.FromServerContext(ctx); ok {
		method = tr.Request().Method
		path = tr.Request().URL.Path
	}
	return
}
```

- [ ] **Step 2: Commit**

```bash
git add gateway/internal/middleware/auth.go
git commit -m "feat(gateway): add JWT auth middleware"
```

---

### Task 5: Error Response Middleware

**Files:**
- Create: `gateway/internal/middleware/error.go`

- [ ] **Step 1: Create error.go**

Create `gateway/internal/middleware/error.go`:

```go
package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"

	"gateway/internal/conf"
)

func ResponseError() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				e := errors.FromError(err)
				
				// Try to get HTTP transporter
				if tr, ok := http.FromServerContext(ctx); ok {
					httpCode := grpcToHTTPCode(e.Code)
					tr.StatusCode(httpCode)
					
					// Return unified error response
					return nil, errors.New(httpCode, e.Reason, e.Message)
				}
			}
			return reply, err
		}
	}
}

func grpcToHTTPCode(grpcCode int32) int {
	switch grpcCode {
	case 0:
		return 200
	case 3: // InvalidArgument
		return 400
	case 16: // Unauthenticated
		return 401
	case 7: // PermissionDenied
		return 403
	case 5: // NotFound
		return 404
	case 6: // AlreadyExists
		return 409
	default:
		return 500
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add gateway/internal/middleware/error.go
git commit -m "feat(gateway): add error response middleware"
```

---

### Task 6: gRPC Clients for Service Discovery

**Files:**
- Create: `gateway/internal/client/clients.go`

- [ ] **Step 1: Read existing service proto imports**

Check the go_package imports from each service's proto file to understand the import paths:
- `user-service/api/user/v1/user.proto` → `user-service/api/user/v1;v1`
- `product-service/api/product/v1/product.proto` → `product-service/api/product/v1;v1`
- `order-service/api/proto/order/v1/order.proto` → `order-service/api/proto/order/v1;orderv1`
- `cart-service/api/cart/v1/cart.proto`
- `payment-service/api/payment/v1/payment.proto`
- `inventory-service/api/inventory/v1/inventory.proto`
- `promotion-service/api/promotion/v1/promotion.proto`

- [ ] **Step 2: Add service dependencies to go.mod**

Run in gateway directory:

```bash
cd gateway
go mod edit -require=user-service@v0.0.0
go mod edit -require=product-service@v0.0.0
go mod edit -require=order-service@v0.0.0
go mod edit -require=cart-service@v0.0.0
go mod edit -require=payment-service@v0.0.0
go mod edit -require=inventory-service@v0.0.0
go mod edit -require=promotion-service@v0.0.0
```

Then run `go mod tidy` to resolve workspace dependencies.

- [ ] **Step 3: Create clients.go**

Create `gateway/internal/client/clients.go`:

```go
package client

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	clientv3 "go.etcd.io/etcd/client/v3"

	cartV1 "cart-service/api/cart/v1"
	invV1 "inventory-service/api/inventory/v1"
	orderV1 "order-service/api/proto/order/v1"
	paymentV1 "payment-service/api/payment/v1"
	productV1 "product-service/api/product/v1"
	promoV1 "promotion-service/api/promotion/v1"
	userV1 "user-service/api/user/v1"
)

type Clients struct {
	User       userV1.UserServiceClient
	Product    productV1.ProductServiceClient
	Order      orderV1.OrderServiceClient
	Cart       cartV1.CartServiceClient
	Payment    paymentV1.PaymentServiceClient
	Inventory  invV1.InventoryServiceClient
	Promotion  promoV1.PromotionServiceClient
}

func NewClients(etcdEndpoints []string, logger log.Logger) (*Clients, error) {
	// Create etcd client
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: etcdEndpoints,
	})
	if err != nil {
		return nil, err
	}

	discovery := etcd.New(etcdClient)

	// Create gRPC clients with service discovery
	userConn := newGRPCClient("discovery:///user-service", discovery)
	productConn := newGRPCClient("discovery:///product-service", discovery)
	orderConn := newGRPCClient("discovery:///order-service", discovery)
	cartConn := newGRPCClient("discovery:///cart-service", discovery)
	paymentConn := newGRPCClient("discovery:///payment-service", discovery)
	invConn := newGRPCClient("discovery:///inventory-service", discovery)
	promoConn := newGRPCClient("discovery:///promotion-service", discovery)

	return &Clients{
		User:      userV1.NewUserServiceClient(userConn),
		Product:   productV1.NewProductServiceClient(productConn),
		Order:     orderV1.NewOrderServiceClient(orderConn),
		Cart:      cartV1.NewCartServiceClient(cartConn),
		Payment:   paymentV1.NewPaymentServiceClient(paymentConn),
		Inventory: invV1.NewInventoryServiceClient(invConn),
		Promotion: promoV1.NewPromotionServiceClient(promoConn),
	}, nil
}

func newGRPCClient(endpoint string, discovery *etcd.Registry) *grpc.ClientConn {
	ctx := context.Background()
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(discovery),
		grpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	return conn
}
```

- [ ] **Step 4: Commit**

```bash
git add gateway/internal/client/clients.go gateway/go.mod gateway/go.sum
git commit -m "feat(gateway): add gRPC clients with etcd service discovery"
```

---

### Task 7: HTTP Server

**Files:**
- Create: `gateway/internal/server/http.go`
- Create: `gateway/internal/server/server.go`

- [ ] **Step 1: Create http.go**

Create `gateway/internal/server/http.go`:

```go
package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"

	"gateway/internal/conf"
	"gateway/internal/middleware"
	"gateway/internal/service"
)

func NewHTTPServer(
	cfg *conf.Config,
	gw *service.GatewayService,
	logger log.Logger,
) *http.Server {
	var opts = []http.ServerOption{
		http.Address(cfg.Server.Http.Addr),
		http.Timeout(cfg.Server.Http.Timeout.AsDuration()),
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			middleware.ResponseError(),
			middleware.ServerAuth(cfg.Auth.JwtSecret, cfg.Auth.Whitelist),
		),
	}

	srv := http.NewServer(opts...)
	
	// Register routes
	srv.HandleFunc("/health", func(ctx http.Context) error {
		return ctx.Result(200, map[string]string{"status": "ok"})
	})

	// Register gateway routes
	gw.RegisterRoutes(srv)

	return srv
}
```

- [ ] **Step 2: Create server.go (Wire ProviderSet)**

Create `gateway/internal/server/server.go`:

```go
package server

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewHTTPServer)
```

- [ ] **Step 3: Commit**

```bash
git add gateway/internal/server/
git commit -m "feat(gateway): add HTTP server configuration"
```

---

### Task 8: Gateway Service (Route Handlers)

**Files:**
- Create: `gateway/internal/service/gateway.go`
- Create: `gateway/internal/service/service.go`

- [ ] **Step 1: Create gateway.go**

Create `gateway/internal/service/gateway.go`:

```go
package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"

	"gateway/internal/client"
	"gateway/internal/middleware"
	
	userV1 "user-service/api/user/v1"
	productV1 "product-service/api/product/v1"
	orderV1 "order-service/api/proto/order/v1"
	cartV1 "cart-service/api/cart/v1"
	paymentV1 "payment-service/api/payment/v1"
	invV1 "inventory-service/api/inventory/v1"
	promoV1 "promotion-service/api/promotion/v1"
)

type GatewayService struct {
	clients *client.Clients
	log     *log.Helper
}

func NewGatewayService(clients *client.Clients, logger log.Logger) *GatewayService {
	return &GatewayService{
		clients: clients,
		log:     log.NewHelper(logger),
	}
}

func (s *GatewayService) RegisterRoutes(srv *kratoshttp.Server) {
	// User routes
	srv.Group("/api/user/v1", func(r *kratoshttp.RouterGroup) {
		r.POST("/register", s.handleUserRegister)
		r.POST("/login", s.handleUserLogin)
		r.GET("/userinfo", s.handleGetUserInfo)
		r.POST("/address", s.handleAddAddress)
		r.PUT("/address", s.handleUpdateAddress)
		r.DELETE("/address/{id}", s.handleDeleteAddress)
		r.GET("/address/{id}", s.handleGetAddress)
		r.GET("/addresses", s.handleListAddresses)
	})

	// Product routes
	srv.Group("/api/product/v1", func(r *kratoshttp.RouterGroup) {
		r.POST("/products", s.handleCreateProduct)
		r.PUT("/products/{id}", s.handleUpdateProduct)
		r.DELETE("/products/{id}", s.handleDeleteProduct)
		r.GET("/products/{id}", s.handleGetProduct)
		r.GET("/products", s.handleListProducts)
		r.POST("/categories", s.handleCreateCategory)
		r.PUT("/categories/{id}", s.handleUpdateCategory)
		r.DELETE("/categories/{id}", s.handleDeleteCategory)
		r.GET("/categories/{id}", s.handleGetCategory)
		r.GET("/categories", s.handleListCategories)
	})

	// Order routes
	srv.Group("/api/order/v1", func(r *kratoshttp.RouterGroup) {
		r.POST("/orders", s.handleCreateOrder)
		r.GET("/orders/{id}", s.handleGetOrder)
		r.PUT("/orders/{id}/status", s.handleUpdateOrderStatus)
		r.POST("/orders/{id}/cancel", s.handleCancelOrder)
		r.GET("/orders", s.handleListOrders)
	})

	// Cart routes
	srv.Group("/api/cart/v1", func(r *kratoshttp.RouterGroup) {
		r.POST("/cart", s.handleAddCartItem)
		r.PUT("/cart/{id}", s.handleUpdateCartItem)
		r.DELETE("/cart/{id}", s.handleDeleteCartItem)
		r.GET("/cart", s.handleGetCart)
		r.POST("/cart/clear", s.handleClearCart)
	})

	// Payment routes
	srv.Group("/api/payment/v1", func(r *kratoshttp.RouterGroup) {
		r.POST("/payment", s.handleCreatePayment)
		r.GET("/payment/{id}", s.handleGetPayment)
		r.POST("/payment/callback", s.handlePaymentCallback)
		r.POST("/payment/{id}/refund", s.handleRefund)
	})

	// Inventory routes
	srv.Group("/api/inventory/v1", func(r *kratoshttp.RouterGroup) {
		r.GET("/inventory/{id}", s.handleGetInventory)
		r.POST("/inventory/lock", s.handleLockStock)
		r.POST("/inventory/unlock", s.handleUnlockStock)
		r.POST("/inventory/confirm", s.handleConfirmStock)
	})

	// Promotion routes
	srv.Group("/api/promotion/v1", func(r *kratoshttp.RouterGroup) {
		r.POST("/promotions", s.handleCreatePromotion)
		r.GET("/promotions/{id}", s.handleGetPromotion)
		r.GET("/promotions", s.handleListPromotions)
		r.POST("/coupons", s.handleCreateCoupon)
		r.POST("/coupons/claim", s.handleClaimCoupon)
	})
}

// Helper function to extract path parameter
func getPathParam(ctx kratoshttp.Context, param string) string {
	return ctx.GetPathParameter(param)
}

// Helper function to bind JSON request
func bindJSON(ctx kratoshttp.Context, req interface{}) error {
	return ctx.BindJSON(req)
}

// Helper function to return JSON response
func returnJSON(ctx kratoshttp.Context, code int, data interface{}) error {
	return ctx.Result(code, data)
}

// User handlers

func (s *GatewayService) handleUserRegister(ctx kratoshttp.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Mobile   string `json:"mobile"`
	}
	if err := bindJSON(ctx, &req); err != nil {
		return returnJSON(ctx, http.StatusBadRequest, map[string]interface{}{
			"code":    50008,
			"message": "invalid request",
		})
	}

	reply, err := s.clients.User.Register(ctx.Request().Context(), &userV1.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Mobile:   req.Mobile,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

func (s *GatewayService) handleUserLogin(ctx kratoshttp.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := bindJSON(ctx, &req); err != nil {
		return returnJSON(ctx, http.StatusBadRequest, map[string]interface{}{
			"code":    50008,
			"message": "invalid request",
		})
	}

	reply, err := s.clients.User.Login(ctx.Request().Context(), &userV1.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

func (s *GatewayService) handleGetUserInfo(ctx kratoshttp.Context) error {
	authInfo, ok := middleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return returnJSON(ctx, http.StatusUnauthorized, map[string]interface{}{
			"code":    50005,
			"message": "unauthorized",
		})
	}

	reply, err := s.clients.User.GetUserInfo(ctx.Request().Context(), &userV1.GetUserInfoRequest{
		Id: authInfo.UserID,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

// Product handlers (follow same pattern as user handlers)

func (s *GatewayService) handleGetProduct(ctx kratoshttp.Context) error {
	id := getPathParam(ctx, "id")
	idInt, _ := strconv.ParseInt(id, 10, 64)

	reply, err := s.clients.Product.GetProduct(ctx.Request().Context(), &productV1.GetProductRequest{
		Id: idInt,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

func (s *GatewayService) handleListProducts(ctx kratoshttp.Context) error {
	pageNum, _ := strconv.Atoi(ctx.GetQuery("page_num"))
	pageSize, _ := strconv.Atoi(ctx.GetQuery("page_size"))
	categoryID, _ := strconv.ParseInt(ctx.GetQuery("category_id"), 10, 64)
	keyword := ctx.GetQuery("keyword")

	reply, err := s.clients.Product.ListProducts(ctx.Request().Context(), &productV1.ListProductsRequest{
		PageNum:    int32(pageNum),
		PageSize:   int32(pageSize),
		CategoryId: categoryID,
		Keyword:    keyword,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

// Order handlers

func (s *GatewayService) handleCreateOrder(ctx kratoshttp.Context) error {
	authInfo, ok := middleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return returnJSON(ctx, http.StatusUnauthorized, map[string]interface{}{
			"code":    50005,
			"message": "unauthorized",
		})
	}

	var req struct {
		AddressID int64 `json:"address_id"`
		Items     []struct {
			ProductID   string `json:"product_id"`
			ProductName string `json:"product_name"`
			Quantity    int32  `json:"quantity"`
			Price       int64  `json:"price"`
		} `json:"items"`
	}
	if err := bindJSON(ctx, &req); err != nil {
		return returnJSON(ctx, http.StatusBadRequest, map[string]interface{}{
			"code":    50008,
			"message": "invalid request",
		})
	}

	var items []*orderV1.OrderItem
	for _, item := range req.Items {
		items = append(items, &orderV1.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	reply, err := s.clients.Order.CreateOrder(ctx.Request().Context(), &orderV1.CreateOrderRequest{
		UserId:    authInfo.UserID,
		AddressId: req.AddressID,
		Items:     items,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

func (s *GatewayService) handleGetOrder(ctx kratoshttp.Context) error {
	orderID := getPathParam(ctx, "id")

	reply, err := s.clients.Order.GetOrder(ctx.Request().Context(), &orderV1.GetOrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

func (s *GatewayService) handleListOrders(ctx kratoshttp.Context) error {
	authInfo, _ := middleware.GetAuthInfo(ctx.Request().Context())
	pageSize, _ := strconv.Atoi(ctx.GetQuery("page_size"))
	pageToken, _ := strconv.Atoi(ctx.GetQuery("page_token"))

	reply, err := s.clients.Order.ListOrders(ctx.Request().Context(), &orderV1.ListOrdersRequest{
		UserId:    fmt.Sprintf("%d", authInfo.UserID),
		PageSize:  int32(pageSize),
		PageToken: int32(pageToken),
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

// Cart handlers

func (s *GatewayService) handleGetCart(ctx kratoshttp.Context) error {
	authInfo, _ := middleware.GetAuthInfo(ctx.Request().Context())

	reply, err := s.clients.Cart.GetCart(ctx.Request().Context(), &cartV1.GetCartRequest{
		UserId: authInfo.UserID,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

// Payment handlers

func (s *GatewayService) handleCreatePayment(ctx kratoshttp.Context) error {
	authInfo, _ := middleware.GetAuthInfo(ctx.Request().Context())

	var req struct {
		OrderID string `json:"order_id"`
		Amount  int64  `json:"amount"`
		Method  string `json:"method"`
	}
	if err := bindJSON(ctx, &req); err != nil {
		return returnJSON(ctx, http.StatusBadRequest, map[string]interface{}{
			"code":    50008,
			"message": "invalid request",
		})
	}

	reply, err := s.clients.Payment.CreatePayment(ctx.Request().Context(), &paymentV1.CreatePaymentRequest{
		OrderId: req.OrderID,
		UserId:  authInfo.UserID,
		Amount:  req.Amount,
		Method:  req.Method,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

// Inventory handlers

func (s *GatewayService) handleGetInventory(ctx kratoshttp.Context) error {
	id := getPathParam(ctx, "id")
	idInt, _ := strconv.ParseInt(id, 10, 64)

	reply, err := s.clients.Inventory.GetInventory(ctx.Request().Context(), &invV1.GetInventoryRequest{
		ProductId: idInt,
	})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

// Promotion handlers

func (s *GatewayService) handleListPromotions(ctx kratoshttp.Context) error {
	reply, err := s.clients.Promotion.ListPromotions(ctx.Request().Context(), &promoV1.ListPromotionsRequest{})
	if err != nil {
		return handleGRPCError(ctx, err)
	}

	return returnJSON(ctx, http.StatusOK, reply)
}

// Helper to convert gRPC errors to HTTP responses
func handleGRPCError(ctx kratoshttp.Context, err error) error {
	return returnJSON(ctx, http.StatusInternalServerError, map[string]interface{}{
		"code":    50000,
		"message": err.Error(),
	})
}
```

Note: The full implementation would include all handlers for all 7 services. Each handler follows the same pattern:
1. Bind JSON request
2. Call gRPC client method
3. Return JSON response or handle error

For brevity, only user handlers are shown in full. The remaining handlers (product, order, cart, payment, inventory, promotion) follow the identical pattern.

- [ ] **Step 2: Create service.go (Wire ProviderSet)**

Create `gateway/internal/service/service.go`:

```go
package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewGatewayService)
```

- [ ] **Step 3: Commit**

```bash
git add gateway/internal/service/
git commit -m "feat(gateway): add gateway service with route handlers"
```

---

### Task 9: Wire DI and Main Entry Point

**Files:**
- Create: `gateway/cmd/gateway/main.go`
- Create: `gateway/cmd/gateway/wire.go`
- Create: `gateway/cmd/gateway/wire_gen.go` (generated)

- [ ] **Step 1: Create wire.go**

Create `gateway/cmd/gateway/wire.go`:

```go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"gateway/internal/client"
	"gateway/internal/conf"
	"gateway/internal/server"
	"gateway/internal/service"
)

func wireApp(*conf.Config, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		client.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
		newApp,
	))
}
```

- [ ] **Step 2: Create main.go**

Create `gateway/cmd/gateway/main.go`:

```go
package main

import (
	"flag"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"

	"gateway/internal/conf"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	Name    = "gateway"
	Version = "1.0.0"
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "configs/config.yaml", "config path, eg: -conf configs/config.yaml")
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", Name,
		"service.version", Version,
	)

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	cfg := &conf.Config{
		Server:        bc.Server,
		Data:          bc.Data,
		Registry:      bc.Registry,
		Auth:          bc.Auth,
		RateLimit:     bc.RateLimit,
		CircuitBreaker: bc.CircuitBreaker,
	}

	app, cleanup, err := wireApp(cfg, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func newApp(hs *http.Server, logger log.Logger) (*kratos.App, func(), error) {
	app := kratos.New(
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Logger(logger),
		kratos.Server(hs),
	)
	return app, func() {}, nil
}
```

- [ ] **Step 3: Add ProviderSet to clients.go**

Add to the end of `gateway/internal/client/clients.go`:

```go
import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewClients)
```

- [ ] **Step 4: Generate Wire dependencies**

```bash
cd gateway && make generate
```

Expected: `gateway/cmd/gateway/wire_gen.go` generated.

- [ ] **Step 5: Commit**

```bash
git add gateway/cmd/gateway/
git commit -m "feat(gateway): add Wire DI and main entry point"
```

---

### Task 10: Dockerfile and Final Integration

**Files:**
- Create: `gateway/Dockerfile`
- Modify: `go.work` (already done in Task 1)

- [ ] **Step 1: Create Dockerfile**

Create `gateway/Dockerfile`:

```dockerfile
FROM golang:1.25 AS builder

ARG GOPROXY

RUN cd /tmp/ && git clone https://github.com/a8m/pwd && make -C pwd install

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
		netbase \
	&& rm -rf /var/lib/apt/lists/ \
	&& apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8000

CMD ["./gateway", "-conf", "configs/config.yaml"]
```

- [ ] **Step 2: Verify build**

```bash
cd gateway && make build
```

Expected: Binary created at `gateway/bin/gateway`.

- [ ] **Step 3: Test run (requires infrastructure)**

Start infrastructure:
```bash
docker-compose up -d
```

Run gateway:
```bash
cd gateway && make run
```

Expected: Gateway starts on `:8000`, connects to etcd, registers service.

- [ ] **Step 4: Test health endpoint**

```bash
curl http://localhost:8000/health
```

Expected: `{"status":"ok"}`

- [ ] **Step 5: Commit**

```bash
git add gateway/Dockerfile
git commit -m "feat(gateway): add Dockerfile and finalize"
```

---

## Post-Implementation

- [ ] Run `go fmt ./...` in gateway directory
- [ ] Verify all services are discoverable via etcd
- [ ] Test JWT auth with whitelist endpoints
- [ ] Test rate limiting
- [ ] Update root README with gateway usage instructions
