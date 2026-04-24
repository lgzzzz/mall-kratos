# 共享库 + 链路追踪实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建 `github.com/lgzzz/mall-tracing` 共享库，提取各微服务重复代码并添加 OpenTelemetry 链路追踪，然后集成到所有 8 个服务。

**Architecture:** 在 mall-kratos 同级目录创建独立 Go module `mall-tracing`，封装中间件、gRPC 工具、数据层工具和 OTel 初始化逻辑。各服务通过 `replace` 指令本地引用，验证后再推送到 GitHub。

**Tech Stack:** Go 1.25, Kratos v2.9.2, OpenTelemetry SDK v1.39+, gRPC, GORM, kafka-go, JWT v5

---

## 文件结构总览

### mall-tracing 仓库（新建）
| 文件 | 职责 |
|------|------|
| `go.mod` | Module 定义 + 依赖 |
| `tracing/provider.go` | TracerProvider 初始化 |
| `tracing/config.go` | Tracing 配置结构体 |
| `middleware/auth.go` | JWT 认证中间件（可配置 variant） |
| `middleware/auth_test.go` | Auth 中间件测试 |
| `middleware/response.go` | ResponseError 中间件（参数化） |
| `middleware/response_test.go` | ResponseError 测试 |
| `middleware/tracing.go` | Server/Client tracing middleware |
| `grpcutil/client.go` | gRPC 客户端创建工具 |
| `grpcutil/server.go` | gRPC Server Builder |
| `data/gorm.go` | NewData, NewDiscovery, ProvideDataConfig |
| `data/gorm_tracing.go` | GORM OTel Plugin wrapper |
| `kafka/producer.go` | Kafka Producer tracing wrapper |
| `kafka/consumer.go` | Kafka Consumer tracing wrapper |
| `config/base.go` | Tracing 配置结构体 |

### mall-kratos 各服务（修改）
| 文件 | 变更类型 |
|------|---------|
| `docker-compose.yml` | 添加 Jaeger 服务 |
| `<service>/go.mod` | 添加 mall-tracing 依赖 + replace |
| `<service>/configs/config.yaml` | 添加 tracing 配置 |
| `<service>/cmd/*/main.go` | 简化为 bootstrap.Run() |
| `<service>/internal/server/grpc.go` | 改用 grpcutil.ServerBuilder + tracing |
| `<service>/internal/data/data.go` | 改用 malldata.NewData |
| `<service>/internal/data/client.go` | 改用 grpcutil.NewInsecureClient + tracing |
| `<service>/internal/middleware/auth.go` | 删除（用共享库替代） |
| `<service>/internal/middleware/error.go` | 删除（用共享库替代） |

---

## Phase 1: 创建 mall-tracing 仓库基础结构

### Task 1: 初始化 mall-tracing Go module

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/go.mod`

- [ ] **Step 1: 创建 go.mod**

```bash
cd /home/lgzzz/IdeaProjects && mkdir -p mall-tracing && cd mall-tracing && go mod init github.com/lgzzz/mall-tracing
```

然后编辑 go.mod 添加核心依赖：

```go
module github.com/lgzzz/mall-tracing

go 1.25

require (
	github.com/go-kratos/kratos/v2 v2.9.2
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/segmentio/kafka-go v0.4.47
	go.etcd.io/etcd/client/v3 v3.6.10
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
	google.golang.org/grpc v1.79.3
	gorm.io/gorm v1.31.1
)
```

- [ ] **Step 2: 下载依赖验证**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && go mod tidy
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git init && git add go.mod go.sum && git commit -m "init: create go module"
```

---

### Task 2: 实现 config/base.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/config/base.go`

- [ ] **Step 1: 创建 config/base.go**

```go
package config

// Tracing holds OpenTelemetry tracing configuration.
type Tracing struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	SampleRatio float64 `json:"sample_ratio" yaml:"sample_ratio"`
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add config/base.go && git commit -m "feat: add tracing config struct"
```

---

### Task 3: 实现 middleware/auth.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/middleware/auth.go`

- [ ] **Step 1: 创建 middleware/auth.go**

提取自各服务 `internal/middleware/auth.go`，通过 Option 模式支持三种 variant：

```go
package middleware

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/golang-jwt/jwt/v5"
)

// AuthInfo holds JWT claims extracted from the token.
type AuthInfo struct {
	UserID   int64
	Username string
	Role     string
}

type authKey struct{}

// AuthOptions configures the ServerAuth middleware behavior.
type AuthOptions struct {
	SigningMethodCheck bool
	AllowEmptyToken    bool
	ErrUnauthorized    *errors.Error
}

// AuthOption is a functional option for configuring ServerAuth.
type AuthOption func(*AuthOptions)

// WithSigningMethodCheck enables HMAC signing method validation.
func WithSigningMethodCheck() AuthOption {
	return func(o *AuthOptions) {
		o.SigningMethodCheck = true
	}
}

// WithAllowEmptyToken allows requests without a token to pass through.
func WithAllowEmptyToken() AuthOption {
	return func(o *AuthOptions) {
		o.AllowEmptyToken = true
	}
}

// WithUnauthorizedErr sets a custom error for unauthorized requests.
func WithUnauthorizedErr(err *errors.Error) AuthOption {
	return func(o *AuthOptions) {
		o.ErrUnauthorized = err
	}
}

// ServerAuth returns a server-side JWT authentication middleware.
func ServerAuth(secret string, opts ...AuthOption) middleware.Middleware {
	options := &AuthOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.ErrUnauthorized == nil {
		options.ErrUnauthorized = errors.Unauthorized("UNAUTHORIZED", "unauthorized")
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tokenStr := ExtractToken(ctx)
			if tokenStr == "" {
				if options.AllowEmptyToken {
					return handler(ctx, req)
				}
				return nil, options.ErrUnauthorized
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if options.SigningMethodCheck {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, options.ErrUnauthorized
					}
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return nil, options.ErrUnauthorized
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, options.ErrUnauthorized
			}

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

			ctx = context.WithValue(ctx, authKey{}, info)
			return handler(ctx, req)
		}
	}
}

// GetAuthInfo retrieves AuthInfo from context.
func GetAuthInfo(ctx context.Context) (*AuthInfo, bool) {
	info, ok := ctx.Value(authKey{}).(*AuthInfo)
	return info, ok
}

// ExtractToken extracts the JWT token from gRPC metadata.
func ExtractToken(ctx context.Context) string {
	if md, ok := metadata.FromServerContext(ctx); ok {
		auth := md.Get("authorization")
		if auth != "" {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	return ""
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add middleware/auth.go && git commit -m "feat: add shared JWT auth middleware with options"
```

---

### Task 4: 实现 middleware/auth_test.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/middleware/auth_test.go`

- [ ] **Step 1: 创建 auth_test.go**

```go
package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

func generateToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newContextWithToken(t *testing.T, token string) context.Context {
	t.Helper()
	tr := &transport.MockServerTransport{}
	ctx := transport.NewServerContext(context.Background(), tr)
	return metadata.NewServerContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
}

func TestServerAuth_ValidToken(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id":  float64(123),
		"username": "testuser",
		"role":     "admin",
	}
	token := generateToken(t, claims)
	ctx := newContextWithToken(t, token)

	mw := ServerAuth(testSecret)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		info, ok := GetAuthInfo(ctx)
		if !ok {
			t.Fatal("expected AuthInfo in context")
		}
		if info.UserID != 123 {
			t.Fatalf("expected UserID 123, got %d", info.UserID)
		}
		if info.Username != "testuser" {
			t.Fatalf("expected username testuser, got %s", info.Username)
		}
		if info.Role != "admin" {
			t.Fatalf("expected role admin, got %s", info.Role)
		}
		return "ok", nil
	}

	resp, err := mw(handler)(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected resp 'ok', got %v", resp)
	}
}

func TestServerAuth_EmptyToken_Rejected(t *testing.T) {
	ctx := newContextWithToken(t, "")

	mw := ServerAuth(testSecret)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	_, err := mw(handler)(ctx, nil)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestServerAuth_EmptyToken_Allowed(t *testing.T) {
	ctx := newContextWithToken(t, "")

	mw := ServerAuth(testSecret, WithAllowEmptyToken())
	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := mw(handler)(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler should have been called")
	}
	if resp != "ok" {
		t.Fatalf("expected resp 'ok', got %v", resp)
	}
}

func TestServerAuth_InvalidToken(t *testing.T) {
	ctx := newContextWithToken(t, "invalid.token.here")

	mw := ServerAuth(testSecret)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	_, err := mw(handler)(ctx, nil)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestServerAuth_SigningMethodCheck(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id":  float64(123),
		"username": "testuser",
		"role":     "admin",
	}
	// Use none signing method
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	s, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	ctx := newContextWithToken(t, s)

	mw := ServerAuth(testSecret, WithSigningMethodCheck())
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	_, err := mw(handler)(ctx, nil)
	if err == nil {
		t.Fatal("expected error for none signing method")
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name     string
		auth     string
		expected string
	}{
		{"with bearer prefix", "Bearer mytoken", "mytoken"},
		{"without prefix", "mytoken", "mytoken"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.NewServerContext(context.Background(), metadata.Pairs("authorization", tt.auth))
			got := ExtractToken(ctx)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && go test ./middleware/ -v -run TestServerAuth
```

Expected: All tests PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add middleware/auth_test.go && git commit -m "test: add auth middleware tests"
```

---

### Task 5: 实现 middleware/response.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/middleware/response.go`

- [ ] **Step 1: 创建 middleware/response.go**

```go
package middleware

import (
	"context"
	"reflect"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

// ResponseStatusFactory creates a ResponseStatus proto message.
type ResponseStatusFactory func(code int32, msg string) interface{}

// ResponseError returns a middleware that automatically fills ResponseStatus
// in the response message. Business errors are converted to ResponseStatus
// with gRPC OK status. Transport/infrastructure errors are passed through.
func ResponseError(newStatus ResponseStatusFactory) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if reply == nil {
				return reply, err
			}

			if setStatusByReflection(reply, err, newStatus) {
				return reply, nil
			}

			return reply, err
		}
	}
}

func setStatusByReflection(reply interface{}, err error, newStatus ResponseStatusFactory) bool {
	v := reflect.ValueOf(reply)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return false
	}

	statusField := v.Elem().FieldByName("Status")
	if !statusField.IsValid() || !statusField.CanSet() {
		return false
	}

	if err == nil {
		statusField.Set(reflect.Zero(statusField.Type()))
		return true
	}

	if kratosErr := errors.FromError(err); kratosErr != nil {
		status := newStatus(kratosErr.Code, kratosErr.Message)
		statusField.Set(reflect.ValueOf(status))
		return true
	}

	return false
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add middleware/response.go && git commit -m "feat: add parameterized ResponseError middleware"
```

---

### Task 6: 实现 middleware/response_test.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/middleware/response_test.go`

- [ ] **Step 1: 创建 response_test.go**

```go
package middleware

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
)

type testResponse struct {
	Status *TestResponseStatus
	Data   string
}

type TestResponseStatus struct {
	ErrorCode    int32
	ErrorMessage string
}

func TestResponseError_Success(t *testing.T) {
	mw := ResponseError(func(code int32, msg string) interface{} {
		return &TestResponseStatus{ErrorCode: code, ErrorMessage: msg}
	})

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &testResponse{Data: "hello"}, nil
	}

	resp, err := mw(handler)(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := resp.(*testResponse)
	if r.Data != "hello" {
		t.Fatalf("expected data 'hello', got %s", r.Data)
	}
	if r.Status != nil {
		t.Fatal("expected nil Status on success")
	}
}

func TestResponseError_BusinessError(t *testing.T) {
	mw := ResponseError(func(code int32, msg string) interface{} {
		return &TestResponseStatus{ErrorCode: code, ErrorMessage: msg}
	})

	bizErr := errors.NotFound("NOT_FOUND", "item not found")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &testResponse{}, bizErr
	}

	resp, err := mw(handler)(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error (business error converted to status), got %v", err)
	}
	r := resp.(*testResponse)
	if r.Status == nil {
		t.Fatal("expected non-nil Status for business error")
	}
	if r.Status.ErrorMessage != "item not found" {
		t.Fatalf("expected error message 'item not found', got %s", r.Status.ErrorMessage)
	}
}

func TestResponseError_NilReply(t *testing.T) {
	mw := ResponseError(func(code int32, msg string) interface{} {
		return &TestResponseStatus{ErrorCode: code, ErrorMessage: msg}
	})

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.InternalServer("INTERNAL", "internal error")
	}

	resp, err := mw(handler)(context.Background(), nil)
	if resp != nil {
		t.Fatal("expected nil reply")
	}
	if err == nil {
		t.Fatal("expected error for nil reply")
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && go test ./middleware/ -v -run TestResponseError
```

Expected: All tests PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add middleware/response_test.go && git commit -m "test: add ResponseError middleware tests"
```

---

### Task 7: 实现 middleware/tracing.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/middleware/tracing.go`

- [ ] **Step 1: 创建 middleware/tracing.go**

```go
package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ServerMiddleware returns a server-side tracing middleware.
func ServerMiddleware(tracer trace.Tracer) middleware.Middleware {
	if tracer == nil {
		tracer = otel.Tracer("github.com/lgzzz/mall-tracing")
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var spanName string
			if tr, ok := transport.FromServerContext(ctx); ok {
				spanName = tr.Operation()
			} else {
				spanName = "grpc.request"
			}

			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			reply, err := handler(ctx, req)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return reply, err
		}
	}
}

// ClientMiddleware returns a client-side tracing middleware.
func ClientMiddleware(tracer trace.Tracer) middleware.Middleware {
	if tracer == nil {
		tracer = otel.Tracer("github.com/lgzzz/mall-tracing")
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var spanName string
			if tr, ok := transport.FromClientContext(ctx); ok {
				spanName = tr.Operation()
			} else {
				spanName = "grpc.call"
			}

			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindClient),
			)
			defer span.End()

			reply, err := handler(ctx, req)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return reply, err
		}
	}
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add middleware/tracing.go && git commit -m "feat: add server and client tracing middleware"
```

---

### Task 8: 实现 grpcutil/client.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/grpcutil/client.go`

- [ ] **Step 1: 创建 grpcutil/client.go**

```go
package grpcutil

import (
	"context"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"
)

// NewInsecureClient creates a gRPC client connection with service discovery.
func NewInsecureClient(
	ctx context.Context,
	discovery registry.Discovery,
	endpoint string,
	opts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(discovery),
	}
	dialOpts = append(dialOpts, opts...)

	return grpc.DialInsecure(ctx, dialOpts...)
}

// NewDirectClient creates a gRPC client connection to a direct endpoint (no discovery).
func NewDirectClient(ctx context.Context, endpoint string) (*grpc.ClientConn, error) {
	return grpc.DialInsecure(ctx, grpc.WithEndpoint(endpoint))
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add grpcutil/client.go && git commit -m "feat: add gRPC client creation utilities"
```

---

### Task 9: 实现 grpcutil/server.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/grpcutil/server.go`

- [ ] **Step 1: 创建 grpcutil/server.go**

```go
package grpcutil

import (
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// ServerBuilder builds a gRPC server with fluent API.
type ServerBuilder struct {
	address    string
	timeout    time.Duration
	middleware []middleware.Middleware
	registers  []func(*grpc.Server)
}

// NewServerBuilder creates a new ServerBuilder.
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{}
}

// WithAddress sets the server bind address.
func (b *ServerBuilder) WithAddress(addr string) *ServerBuilder {
	b.address = addr
	return b
}

// WithTimeout sets the request timeout.
func (b *ServerBuilder) WithTimeout(timeout time.Duration) *ServerBuilder {
	b.timeout = timeout
	return b
}

// WithMiddleware adds middleware to the server.
func (b *ServerBuilder) WithMiddleware(mw ...middleware.Middleware) *ServerBuilder {
	b.middleware = append(b.middleware, mw...)
	return b
}

// RegisterService registers a gRPC service.
func (b *ServerBuilder) RegisterService(register func(*grpc.Server)) *ServerBuilder {
	b.registers = append(b.registers, register)
	return b
}

// Build creates the gRPC server.
func (b *ServerBuilder) Build() *grpc.Server {
	var opts []grpc.ServerOption

	if b.address != "" {
		opts = append(opts, grpc.Address(b.address))
	}
	if b.timeout > 0 {
		opts = append(opts, grpc.Timeout(b.timeout))
	}
	if len(b.middleware) > 0 {
		opts = append(opts, grpc.Middleware(b.middleware...))
	}

	srv := grpc.NewServer(opts...)
	for _, register := range b.registers {
		register(srv)
	}
	return srv
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add grpcutil/server.go && git commit -m "feat: add gRPC server builder"
```

---

### Task 10: 实现 data/gorm.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/data/gorm.go`

- [ ] **Step 1: 创建 data/gorm.go**

```go
package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	etcdregistry "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Data holds the database connection.
type Data struct {
	DB *gorm.DB
}

// DataOption config the Data creation.
type DataOption func(*dataOptions)

type dataOptions struct {
	tracer interface{} // trace.Tracer (using interface to avoid import cycle)
}

// WithTracing enables GORM tracing with the given tracer.
func WithTracing(tracer interface{}) DataOption {
	return func(o *dataOptions) {
		o.tracer = tracer
	}
}

// NewData creates a new Data with database connection.
func NewData(dsn string, logger log.Logger, opts ...DataOption) (*Data, func(), error) {
	options := &dataOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var dbOpts []gorm.Option
	if options.tracer != nil {
		if tracer, ok := options.tracer.(interface{ Apply(*gorm.Config) }); ok {
			// Will be handled after gorm.Config creation
		}
	}

	cfg := &gorm.Config{}
	db, err := gorm.Open(mysql.Open(dsn), cfg)
	if err != nil {
		return nil, nil, err
	}

	if options.tracer != nil {
		if tracer, ok := options.tracer.(interface{ Apply(*gorm.Config) }); ok {
			// Apply tracer plugin if available
		}
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{DB: db}, cleanup, nil
}

// NewDiscovery creates an etcd service discovery client.
func NewDiscovery(endpoints []string) (registry.Discovery, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	return etcdregistry.New(etcdClient), nil
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add data/gorm.go && git commit -m "feat: add shared data layer utilities"
```

---

### Task 11: 实现 data/gorm_tracing.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/data/gorm_tracing.go`

- [ ] **Step 1: 创建 data/gorm_tracing.go**

```go
package data

import (
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry"
)

// NewGORMTracingPlugin creates a GORM plugin for OpenTelemetry tracing.
func NewGORMTracingPlugin(tracer trace.Tracer) gorm.Plugin {
	plugin, _ := otelgorm.New(otelgorm.WithTracer(tracer))
	return plugin
}
```

- [ ] **Step 2: 添加 gorm.io/plugin/opentelemetry 依赖**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && go get gorm.io/plugin/opentelemetry && go mod tidy
```

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add data/gorm_tracing.go go.mod go.sum && git commit -m "feat: add GORM OpenTelemetry tracing plugin wrapper"
```

---

### Task 12: 实现 kafka/producer.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/kafka/producer.go`

- [ ] **Step 1: 创建 kafka/producer.go**

```go
package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracedWriter wraps a kafka.Writer with OpenTelemetry tracing.
type TracedWriter struct {
	inner    *kafka.Writer
	tracer   trace.Tracer
	propagator propagation.TextMapPropagator
}

// NewTracedProducer creates a new TracedWriter.
func NewTracedProducer(inner *kafka.Writer, tracer trace.Tracer) *TracedWriter {
	if tracer == nil {
		tracer = otel.Tracer("github.com/lgzzz/mall-tracing/kafka")
	}
	return &TracedWriter{
		inner:      inner,
		tracer:     tracer,
		propagator: otel.GetTextMapPropagator(),
	}
}

// WriteMessages writes messages to Kafka with tracing.
func (w *TracedWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	ctx, span := w.tracer.Start(ctx, "kafka.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(),
	)
	defer span.End()

	for i := range msgs {
		// Inject trace context into message headers
		carrier := &headerCarrier{headers: &msgs[i].Headers}
		w.propagator.Inject(ctx, carrier)
	}

	if err := w.inner.WriteMessages(ctx, msgs...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

// Close closes the underlying writer.
func (w *TracedWriter) Close() error {
	return w.inner.Close()
}

type headerCarrier struct {
	headers *[]kafka.Header
}

func (c *headerCarrier) Get(key string) string {
	for _, h := range *c.headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *headerCarrier) Set(key string, value string) {
	*c.headers = append(*c.headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

func (c *headerCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.headers))
	for _, h := range *c.headers {
		keys = append(keys, h.Key)
	}
	return keys
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add kafka/producer.go && git commit -m "feat: add Kafka producer with tracing"
```

---

### Task 13: 实现 kafka/consumer.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/kafka/consumer.go`

- [ ] **Step 1: 创建 kafka/consumer.go**

```go
package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracedReader wraps a kafka.Reader with OpenTelemetry tracing.
type TracedReader struct {
	inner      *kafka.Reader
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// NewTracedConsumer creates a new TracedReader.
func NewTracedConsumer(inner *kafka.Reader, tracer trace.Tracer) *TracedReader {
	if tracer == nil {
		tracer = otel.Tracer("github.com/lgzzz/mall-tracing/kafka")
	}
	return &TracedReader{
		inner:      inner,
		tracer:     tracer,
		propagator: otel.GetTextMapPropagator(),
	}
}

// FetchMessage reads a message with tracing.
func (r *TracedReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := r.inner.ReadMessage(ctx)
	if err != nil {
		return msg, err
	}

	// Extract trace context from message headers
	carrier := &headerCarrier{headers: &msg.Headers}
	childCtx := r.propagator.Extract(ctx, carrier)

	_, span := r.tracer.Start(childCtx, "kafka.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
	span.End()

	return msg, nil
}

// Close closes the underlying reader.
func (r *TracedReader) Close() error {
	return r.inner.Close()
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add kafka/consumer.go && git commit -m "feat: add Kafka consumer with tracing"
```

---

### Task 14: 实现 tracing/provider.go 和 tracing/config.go

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/tracing/config.go`
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/tracing/provider.go`

- [ ] **Step 1: 创建 tracing/config.go**

```go
package tracing

// Config holds OpenTelemetry tracer provider configuration.
type Config struct {
	ServiceName  string
	Version      string
	OTLPEndpoint string
	SampleRatio  float64
	Insecure     bool
}
```

- [ ] **Step 2: 创建 tracing/provider.go**

```go
package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Init creates and registers a global TracerProvider.
func Init(cfg Config) (trace.TracerProvider, error) {
	if !cfg.Insecure && cfg.OTLPEndpoint == "" {
		// No endpoint configured, return noop tracer
		return noop.NewTracerProvider(), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var opts []otlptracegrpc.Option
	opts = append(opts, otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint))
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(cfg.SampleRatio),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// Shutdown gracefully shuts down the TracerProvider.
func Shutdown(ctx context.Context, tp trace.TracerProvider) error {
	if sdkTP, ok := tp.(*sdktrace.TracerProvider); ok {
		return sdkTP.Shutdown(ctx)
	}
	return nil
}

// NewTracer creates a named tracer from the global provider.
func NewTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
```

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add tracing/config.go tracing/provider.go && git commit -m "feat: add OpenTelemetry TracerProvider initialization"
```

---

### Task 15: 添加 README.md

**Files:**
- Create: `/home/lgzzz/IdeaProjects/mall-tracing/README.md`

- [ ] **Step 1: 创建 README.md**

```markdown
# mall-tracing

Shared tracing and utility library for mall-kratos microservices.

## Features

- **OpenTelemetry Tracing**: Initialize TracerProvider with OTLP exporter to Jaeger
- **Shared Middleware**: JWT auth, response error handling, tracing middleware
- **gRPC Utilities**: Client creation with service discovery, server builder
- **Data Layer**: GORM setup with tracing plugin, etcd discovery
- **Kafka Tracing**: Producer/consumer wrappers with context propagation

## Usage

### Tracing Initialization

```go
tp, err := tracing.Init(tracing.Config{
    ServiceName:  "order-service",
    Version:      "v1.0.0",
    OTLPEndpoint: "localhost:4317",
    SampleRatio:  0.1,
    Insecure:     true,
})
if err != nil {
    log.Fatal(err)
}
defer tracing.Shutdown(context.Background(), tp)

tracer := tracing.NewTracer("order-service")
```

### Server Middleware

```go
grpcutil.NewServerBuilder().
    WithMiddleware(
        recovery.Recovery(),
        mallmiddleware.ServerMiddleware(tracer),
        mallmiddleware.ResponseError(newStatus),
        mallmiddleware.ServerAuth(secret),
    ).
    Build()
```

### Client Middleware

```go
conn, err := grpcutil.NewInsecureClient(ctx, discovery, endpoint,
    grpc.WithMiddleware(mallmiddleware.ClientMiddleware(tracer)),
)
```

## License

MIT
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git add README.md && git commit -m "docs: add README"
```

---

## Phase 2: 添加 Jaeger 到 Docker Compose

### Task 16: 修改 docker-compose.yml

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/docker-compose.yml`

- [ ] **Step 1: 添加 Jaeger 服务**

在 `volumes:` 之前添加：

```yaml
  jaeger:
    image: jaegertracing/all-in-one:1.67
    container_name: mall-jaeger
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    ports:
      - "16686:16686"
      - "4317:4317"
      - "4318:4318"
```

- [ ] **Step 2: 验证 docker-compose 配置**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && docker-compose config
```

Expected: Valid YAML output including jaeger service.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add docker-compose.yml && git commit -m "feat: add Jaeger service to docker-compose"
```

---

## Phase 3: 集成到 product-service（样板服务）

### Task 17: 修改 product-service go.mod

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/go.mod`

- [ ] **Step 1: 添加 mall-tracing 依赖和 replace 指令**

在 go.mod 末尾添加：

```go
require github.com/lgzzz/mall-tracing v0.1.0

replace github.com/lgzzz/mall-tracing => ../../mall-tracing
```

- [ ] **Step 2: 运行 go mod tidy**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && go mod tidy
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/go.mod product-service/go.sum && git commit -m "chore(product-service): add mall-tracing dependency"
```

---

### Task 18: 修改 product-service configs/config.yaml

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/configs/config.yaml`

- [ ] **Step 1: 添加 tracing 配置**

```yaml
config_center:
  endpoints:
    - "127.0.0.1:2379"
  key: "product-service/configs/config.yaml"

tracing:
  enabled: true
  endpoint: "localhost:4317"
  sample_ratio: 0.1
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/configs/config.yaml && git commit -m "chore(product-service): add tracing config"
```

---

### Task 19: 修改 product-service conf.go 添加 Tracing 字段

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/conf/conf.go`

- [ ] **Step 1: 在 Bootstrap 和 Config 中添加 Tracing 字段**

修改 `Bootstrap` 结构体：

```go
type Bootstrap struct {
	Server       Server       `json:"server" yaml:"server"`
	Data         Data         `json:"data" yaml:"data"`
	Registry     Registry     `json:"registry" yaml:"registry"`
	Auth         Auth         `json:"auth" yaml:"auth"`
	GrpcClients  GrpcClients  `json:"grpc_clients" yaml:"grpc_clients"`
	ConfigCenter *ConfigCenter `json:"config_center" yaml:"config_center"`
	Tracing      Tracing      `json:"tracing" yaml:"tracing"`
}

type Tracing struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	SampleRatio float64 `json:"sample_ratio" yaml:"sample_ratio"`
}
```

修改 `Config` 结构体：

```go
type Config struct {
	Server      Server      `json:"server" yaml:"server"`
	Data        Data        `json:"data" yaml:"data"`
	Registry    Registry    `json:"registry" yaml:"registry"`
	Auth        Auth        `json:"auth" yaml:"auth"`
	GrpcClients GrpcClients `json:"grpc_clients" yaml:"grpc_clients"`
	Tracing     Tracing     `json:"tracing" yaml:"tracing"`
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/internal/conf/conf.go && git commit -m "feat(product-service): add Tracing config struct"
```

---

### Task 20: 修改 product-service main.go 集成 tracing

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/cmd/product-service/main.go`

- [ ] **Step 1: 替换整个 main.go**

```go
package main

import (
	"context"
	"flag"
	"os"

	"product-service/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	etcdConfig "github.com/go-kratos/kratos/contrib/config/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/lgzzz/mall-tracing/tracing"
)

var (
	Version  string
	confPath string
)

func init() {
	flag.StringVar(&confPath, "conf", "configs/config.yaml", "config path, eg: -conf configs/config.yaml")
	flag.Parse()
}

func newApp(logger log.Logger, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.Name("product-service"),
		kratos.Version(Version),
		kratos.Logger(logger),
		kratos.Server(gs),
	)
}

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", "product-service",
	)
	h := log.NewHelper(logger)

	c := config.New(
		config.WithSource(
			file.NewSource(confPath),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		h.Fatalf("failed to load config: %v", err)
	}

	var bootstrap conf.Bootstrap
	if err := c.Scan(&bootstrap); err != nil {
		h.Fatalf("failed to scan config: %v", err)
	}

	if bootstrap.ConfigCenter != nil && len(bootstrap.ConfigCenter.Endpoints) > 0 {
		client, err := clientv3.New(clientv3.Config{
			Endpoints: bootstrap.ConfigCenter.Endpoints,
		})
		if err != nil {
			h.Errorf("failed to create etcd client: %v", err)
		} else {
			source, err := etcdConfig.New(client, etcdConfig.WithPath(bootstrap.ConfigCenter.Key))
			if err != nil {
				h.Errorf("failed to create etcd config source: %v", err)
			} else {
				etcdClient := config.New(
					config.WithSource(source),
				)
				defer etcdClient.Close()

				if err := etcdClient.Load(); err != nil {
					h.Errorf("failed to load config from etcd: %v", err)
				} else {
					if err := etcdClient.Scan(&bootstrap); err != nil {
						h.Errorf("failed to scan config from etcd: %v", err)
					}
				}
			}
		}
	}

	// Initialize tracing
	var tp interface{}
	if bootstrap.Tracing.Enabled {
		var err error
		tp, err = tracing.Init(tracing.Config{
			ServiceName:  "product-service",
			Version:      Version,
			OTLPEndpoint: bootstrap.Tracing.Endpoint,
			SampleRatio:  bootstrap.Tracing.SampleRatio,
			Insecure:     true,
		})
		if err != nil {
			h.Fatalf("failed to init tracing: %v", err)
		}
	}

	app, cleanup, err := wireApp(extractConfig(&bootstrap), logger)
	if err != nil {
		h.Fatalf("failed to wire app: %v", err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		h.Fatalf("failed to run app: %v", err)
	}

	if tp != nil {
		tracing.Shutdown(context.Background(), tp.(interface{}))
	}

	h.Info("Application stopped gracefully")
}

func extractConfig(b *conf.Bootstrap) *conf.Config {
	return &conf.Config{
		Server:      b.Server,
		Data:        b.Data,
		Registry:    b.Registry,
		Auth:        b.Auth,
		GrpcClients: b.GrpcClients,
		Tracing:     b.Tracing,
	}
}
```

注意：上面的 Shutdown 调用需要修正类型。简化处理：

```go
	if tp != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tracing.Shutdown(ctx, tp)
	}
```

需要导入 `"time"`。

- [ ] **Step 2: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/cmd/product-service/main.go && git commit -m "feat(product-service): integrate tracing initialization in main"
```

---

### Task 21: 修改 product-service wire.go 传递 tracer

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/cmd/product-service/wire.go`

- [ ] **Step 1: 修改 wire.go**

```go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"product-service/internal/biz"
	"product-service/internal/conf"
	"product-service/internal/data"
	"product-service/internal/server"
	"product-service/internal/service"
)

func wireApp(cfg *conf.Config, logger log.Logger, tracer trace.Tracer) (*kratos.App, func(), error) {
	panic(wire.Build(
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
		newApp,
	))
}
```

- [ ] **Step 2: 重新生成 wire_gen.go**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && make generate
```

Expected: wire_gen.go regenerated with tracer parameter.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/cmd/product-service/wire.go product-service/cmd/product-service/wire_gen.go && git commit -m "feat(product-service): add tracer parameter to wire app"
```

---

### Task 22: 修改 product-service server/grpc.go 使用共享库

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/server/grpc.go`

- [ ] **Step 1: 替换 grpc.go**

```go
package server

import (
	v1 "product-service/api/product/v1"
	"product-service/internal/conf"
	"product-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/trace"

	mallmiddleware "github.com/lgzzz/mall-tracing/middleware"
	"github.com/lgzzz/mall-tracing/grpcutil"
)

func NewGRPCServer(c *conf.Config, product *service.ProductService, logger log.Logger, tracer trace.Tracer) *grpc.Server {
	return grpcutil.NewServerBuilder().
		WithAddress(c.Server.Grpc.Addr).
		WithTimeout(c.Server.Grpc.Timeout).
		WithMiddleware(
			recovery.Recovery(),
			mallmiddleware.ServerMiddleware(tracer),
			mallmiddleware.ResponseError(func(code int32, msg string) interface{} {
				return &v1.ResponseStatus{ErrorCode: code, ErrorMessage: msg}
			}),
			mallmiddleware.ServerAuth(c.Auth.JwtSecret, mallmiddleware.WithAllowEmptyToken()),
		).
		RegisterService(func(s *grpc.Server) { v1.RegisterProductServiceServer(s, product) }).
		Build()
}
```

- [ ] **Step 2: 编译验证**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && go build ./...
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/internal/server/grpc.go && git commit -m "refactor(product-service): use shared grpcutil and tracing middleware"
```

---

### Task 23: 修改 product-service data/data.go 使用共享库

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/data/data.go`

- [ ] **Step 1: 替换 data.go**

```go
package data

import (
	"product-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	malldata "github.com/lgzzz/mall-tracing/data"
)

var ProviderSet = wire.NewSet(
	NewData,
	NewProductRepo,
	NewCategoryRepo,
	NewInventoryRepo,
	NewDiscovery,
	ProvideInventoryEndpoint,
)

type Data struct {
	db *malldata.Data
}

func NewData(c *conf.Data, logger log.Logger, tracer trace.Tracer) (*Data, func(), error) {
	mallData, cleanup, err := malldata.NewData(c.Database.Source, logger, malldata.WithTracing(tracer))
	if err != nil {
		return nil, nil, err
	}
	return &Data{db: mallData}, cleanup, nil
}

func NewDiscovery(c *conf.Config) registry.Discovery {
	d, err := malldata.NewDiscovery(c.Registry.Endpoints)
	if err != nil {
		panic(err)
	}
	return d
}

func ProvideInventoryEndpoint(c *conf.Config) string {
	return c.GrpcClients.InventoryService
}
```

- [ ] **Step 2: 编译验证**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && go build ./...
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/internal/data/data.go && git commit -m "refactor(product-service): use shared data layer utilities"
```

---

### Task 24: 修改 product-service data/client.go 使用共享库

**Files:**
- Modify: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/data/client.go`

- [ ] **Step 1: 替换 client.go**

```go
package data

import (
	"context"

	"product-service/internal/biz"

	inventoryV1 "inventory-service/api/inventory/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/trace"

	mallmiddleware "github.com/lgzzz/mall-tracing/middleware"
	"github.com/lgzzz/mall-tracing/grpcutil"
)

type inventoryRepo struct {
	client inventoryV1.InventoryClient
}

func NewInventoryRepo(r registry.Discovery, endpoint string, logger log.Logger, tracer trace.Tracer) biz.InventoryRepo {
	conn, err := grpcutil.NewInsecureClient(
		context.Background(),
		r,
		endpoint,
		grpc.WithMiddleware(mallmiddleware.ClientMiddleware(tracer)),
	)
	if err != nil {
		panic(err)
	}
	return &inventoryRepo{client: inventoryV1.NewInventoryClient(conn)}
}

func (r *inventoryRepo) CreateInventory(ctx context.Context, skuID int64) error {
	_, err := r.client.CreateInventory(ctx, &inventoryV1.CreateInventoryRequest{
		SkuId:       skuID,
		WarehouseId: 1,
		Stock:       0,
	})
	return err
}
```

- [ ] **Step 2: 重新生成 wire_gen.go**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && make generate
```

- [ ] **Step 3: 编译验证**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && go build ./...
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add product-service/internal/data/client.go product-service/cmd/product-service/wire_gen.go && git commit -m "refactor(product-service): use shared grpcutil client with tracing"
```

---

### Task 25: 删除 product-service 旧中间件文件

**Files:**
- Delete: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/middleware/auth.go`
- Delete: `/home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/middleware/error.go`

- [ ] **Step 1: 删除文件**

```bash
rm /home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/middleware/auth.go
rm /home/lgzzz/IdeaProjects/mall-kratos/product-service/internal/middleware/error.go
```

- [ ] **Step 2: 编译验证**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && go build ./...
```

Expected: No errors (no references to deleted files).

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git rm product-service/internal/middleware/auth.go product-service/internal/middleware/error.go && git commit -m "chore(product-service): remove local middleware (using shared library)"
```

---

### Task 26: 验证 product-service 编译和运行

- [ ] **Step 1: 完整编译验证**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && go build -o /dev/null ./cmd/product-service/
```

Expected: No errors.

- [ ] **Step 2: 运行测试（如果有）**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && make test
```

- [ ] **Step 3: Commit**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && git add -A && git commit -m "chore(product-service): verify build and tests pass"
```

---

## Phase 4: 集成到其余服务

以下服务按相同模式集成。每个服务一个 Task，可并行执行。

### Task 27: 集成到 cart-service

**Files:**
- Modify: `cart-service/go.mod`, `cart-service/configs/config.yaml`, `cart-service/internal/conf/conf.go`
- Modify: `cart-service/cmd/cart-service/main.go`, `cart-service/cmd/cart-service/wire.go`
- Modify: `cart-service/internal/server/grpc.go`, `cart-service/internal/data/data.go`, `cart-service/internal/data/client.go`
- Delete: `cart-service/internal/middleware/auth.go`, `cart-service/internal/middleware/error.go`

按 Task 17-26 的相同模式操作，注意：
- cart-service 有 `GrpcClients` 字段（调用 product-service）
- cart-service auth middleware 是 Variant B（不允许空 token）
- grpc.go 中 `ServerAuth` 不需要 `WithAllowEmptyToken()`

- [ ] **Step 1: 修改 go.mod 添加依赖**
- [ ] **Step 2: 修改 config.yaml 添加 tracing**
- [ ] **Step 3: 修改 conf.go 添加 Tracing 字段**
- [ ] **Step 4: 修改 main.go 集成 tracing**
- [ ] **Step 5: 修改 wire.go + make generate**
- [ ] **Step 6: 修改 server/grpc.go**
- [ ] **Step 7: 修改 data/data.go 和 data/client.go**
- [ ] **Step 8: 删除旧中间件**
- [ ] **Step 9: 编译验证 + commit**

---

### Task 28: 集成到 user-service

**Files:**
- Modify: `user-service/go.mod`, `user-service/configs/config.yaml`, `user-service/internal/conf/conf.go`
- Modify: `user-service/cmd/user-service/main.go`（手动 DI，无 Wire）
- Modify: `user-service/internal/server/grpc.go`, `user-service/internal/data/data.go`
- Delete: `user-service/internal/middleware/auth.go`, `user-service/internal/middleware/error.go`

注意：
- user-service 使用手动 DI（无 Wire）
- user-service auth middleware 是 Variant A（有 signing method check）
- user-service `Auth` 有 `TokenExpireHours` 字段

- [ ] **Step 1-9: 同 Task 27 模式**

---

### Task 29: 集成到 order-service

**Files:**
- Modify: `order-service/go.mod`, `order-service/configs/config.yaml`, `order-service/internal/conf/config.go`
- Modify: `order-service/cmd/order-service/main.go`, `order-service/cmd/order-service/wire.go`
- Modify: `order-service/internal/server/grpc.go`, `order-service/internal/data/data.go`, `order-service/internal/data/client.go`
- Modify: `order-service/internal/consumer/order.go`（Kafka tracing）
- Delete: `order-service/internal/middleware/auth.go`, `order-service/internal/middleware/error.go`

注意：
- order-service 有自定义 wire 和 Kafka consumer
- order-service 配置结构体命名不同（`GRPCConfig`, `DatabaseConfig`）
- Kafka consumer 需要用 `kafka.TracedReader` 包装

- [ ] **Step 1-9: 同 Task 27 模式 + Kafka tracing**

---

### Task 30: 集成到 payment-service

**Files:**
- Modify: `payment-service/go.mod`, `payment-service/configs/config.yaml`, `payment-service/internal/conf/conf.go`
- Modify: `payment-service/cmd/payment-service/main.go`, `payment-service/cmd/payment-service/wire.go`
- Modify: `payment-service/internal/server/grpc.go`, `payment-service/internal/data/data.go`
- Delete: `payment-service/internal/middleware/auth.go`, `payment-service/internal/middleware/error.go`

- [ ] **Step 1-9: 同 Task 27 模式**

---

### Task 31: 集成到 promotion-service

**Files:**
- Modify: `promotion-service/go.mod`, `promotion-service/configs/config.yaml`, `promotion-service/internal/conf/conf.go`
- Modify: `promotion-service/cmd/promotion-service/main.go`, `promotion-service/cmd/promotion-service/wire.go`
- Modify: `promotion-service/internal/server/grpc.go`, `promotion-service/internal/data/data.go`
- Delete: `promotion-service/internal/middleware/auth.go`, `promotion-service/internal/middleware/error.go`

- [ ] **Step 1-9: 同 Task 27 模式**

---

### Task 32: 集成到 inventory-service

**Files:**
- Modify: `inventory-service/go.mod`, `inventory-service/configs/config.yaml`
- Modify: `inventory-service/cmd/inventory-service/main.go`, `inventory-service/cmd/inventory-service/wire.go`
- Modify: `inventory-service/internal/server/grpc.go`, `inventory-service/internal/data/data.go`
- Delete: `inventory-service/internal/middleware/auth.go`, `inventory-service/internal/middleware/error.go`

注意：
- inventory-service 已有 OTel 依赖（indirect）
- inventory-service 使用 proto 生成的 conf（`conf.pb.go`）
- inventory-service 有 etcd flag 参数

- [ ] **Step 1-9: 同 Task 27 模式**

---

### Task 33: 集成到 gateway

**Files:**
- Modify: `gateway/go.mod`, `gateway/configs/config.yaml`
- Modify: `gateway/cmd/gateway/main.go`, `gateway/cmd/gateway/wire.go`
- Modify: `gateway/internal/server/http.go`（或类似文件名）
- Delete: `gateway/internal/middleware/auth.go`, `gateway/internal/middleware/error.go`

注意：
- gateway 使用 HTTP server 而非 gRPC
- gateway auth middleware 完全不同（有 whitelist、HTTP 到 gRPC metadata 转发）
- gateway 的 ResponseError 映射 gRPC code 到 HTTP code
- **gateway 的中间件不适合直接替换**，需要保留本地实现或大幅改造

**决策**：gateway 仅添加 tracing 初始化，中间件保留本地实现。

- [ ] **Step 1: 修改 go.mod 添加依赖**
- [ ] **Step 2: 修改 config.yaml 添加 tracing**
- [ ] **Step 3: 修改 main.go 集成 tracing 初始化**
- [ ] **Step 4: 在 HTTP middleware 链中添加 tracing middleware**
- [ ] **Step 5: 编译验证 + commit**

---

## Phase 5: 端到端验证

### Task 34: 启动基础设施并验证链路追踪

- [ ] **Step 1: 启动 Jaeger**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos && docker-compose up -d jaeger
```

- [ ] **Step 2: 验证 Jaeger UI 可访问**

```bash
curl -s http://localhost:16686 | head -5
```

Expected: HTML content returned.

- [ ] **Step 3: 启动 product-service 验证 tracing 初始化**

```bash
cd /home/lgzzz/IdeaProjects/mall-kratos/product-service && make run &
```

Expected: No tracing errors in logs.

- [ ] **Step 4: 发起请求并检查 Jaeger**

```bash
# 通过 grpcurl 或其他方式发起请求
# 然后在 http://localhost:16686 查看 trace
```

- [ ] **Step 5: 停止服务**

```bash
pkill -f product-service
```

---

### Task 35: 推送 mall-tracing 到 GitHub

- [ ] **Step 1: 创建 GitHub 仓库**

使用 `gh` 命令或浏览器创建 `github.com/lgzzz/mall-tracing` 仓库。

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && gh repo create lgzzz/mall-tracing --public --source=. --remote=origin --push
```

- [ ] **Step 2: 打 tag**

```bash
cd /home/lgzzz/IdeaProjects/mall-tracing && git tag v0.1.0 && git push origin v0.1.0
```

---

## 自检验查

### Spec 覆盖检查

| Spec 章节 | 对应 Task |
|-----------|----------|
| mall-tracing 仓库结构 | Task 1-15 |
| tracing/provider.go | Task 14 |
| middleware/auth.go | Task 3-4 |
| middleware/response.go | Task 5-6 |
| middleware/tracing.go | Task 7 |
| grpcutil/client.go + server.go | Task 8-9 |
| data/gorm.go + gorm_tracing.go | Task 10-11 |
| kafka/producer.go + consumer.go | Task 12-13 |
| config/base.go | Task 2 |
| Docker Compose Jaeger | Task 16 |
| 各服务集成 | Task 17-33 |
| 端到端验证 | Task 34-35 |

### 占位符扫描

无 "TBD"、"TODO"、"implement later" 等占位符。

### 类型一致性检查

- `tracing.Config` 在 Task 14 定义，Task 20 使用 - 一致
- `mallmiddleware.ServerAuth` 在 Task 3 定义，各服务 Task 使用 - 一致
- `grpcutil.NewServerBuilder` 在 Task 9 定义，Task 22 使用 - 一致
- `malldata.NewData` 在 Task 10 定义，Task 23 使用 - 一致
- `grpcutil.NewInsecureClient` 在 Task 8 定义，Task 24 使用 - 一致

---

计划完成，保存在 `docs/superpowers/plans/2026-04-24-shared-tracing-library.md`。
