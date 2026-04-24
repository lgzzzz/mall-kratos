# 共享库 + 链路追踪设计文档

**日期**: 2026-04-24
**主题**: mall-tracing 共享库 + OpenTelemetry 链路追踪

---

## 1. 概述

将 mall-kratos 各微服务中的重复代码提取到独立仓库 `github.com/lgzzz/mall-tracing`，并在此基础上添加 OpenTelemetry 链路追踪功能。

### 目标
- 消除 7 个微服务 + 1 gateway 中的重复代码（约 1500+ 行）
- 为所有服务添加端到端链路追踪（gRPC、数据库、Kafka）
- 统一配置模式，简化新服务接入

### 范围
| 组件 | 状态 |
|------|------|
| 共享中间件（auth、response error） | 提取 |
| 共享 gRPC 客户端工具 | 提取 |
| 共享数据层工具 | 提取 |
| 共享配置结构体 | 提取 |
| OpenTelemetry Tracing | 新增 |
| Jaeger 本地部署 | 新增 |
| Kafka Tracing | 新增 |

---

## 2. mall-tracing 仓库结构

```
mall-tracing/
├── go.mod                          # module: github.com/lgzzz/mall-tracing
├── README.md
│
├── tracing/                        # OpenTelemetry 链路追踪
│   ├── provider.go                 # TracerProvider 初始化
│   ├── config.go                   # Tracing 配置结构体
│   └── resource.go                 # OTel Resource 配置
│
├── middleware/                     # Kratos 中间件
│   ├── auth.go                     # ServerAuth（可配置 variant）
│   ├── response.go                 # ResponseError（参数化）
│   ├── tracing.go                  # Server/Client tracing middleware
│   └── auth_test.go
│
├── grpcutil/                       # gRPC 工具
│   ├── client.go                   # NewInsecureClient
│   └── server.go                   # gRPC Server Builder
│
├── data/                           # 数据层工具
│   ├── gorm.go                     # NewData, ProvideDataConfig
│   └── gorm_tracing.go             # GORM OTel Plugin wrapper
│
├── kafka/                          # Kafka 工具
│   ├── producer.go                 # Producer wrapper with tracing
│   └── consumer.go                 # Consumer wrapper with tracing
│
├── config/                         # 基础配置结构体
│   └── base.go                     # Server, Grpc, Data, Database, Registry 等
│
└── bootstrap/                      # 应用启动引导
    └── app.go                      # Config loading + etcd + wire + app.Run
```

---

## 3. 各模块详细设计

### 3.1 tracing/ - OpenTelemetry 初始化

**provider.go**:
```go
type Config struct {
    ServiceName  string
    Version      string
    OTLPEndpoint string        // e.g., "localhost:4317"
    SampleRatio  float64       // 0.0 - 1.0
    Insecure     bool          // 开发环境用
}

func Init(cfg Config) (*sdktrace.TracerProvider, error)
```

- 使用 OTLP gRPC exporter 发送到 Jaeger
- 配置 `ParentBased(Sampler)` 采样器
- 自动注入 `service.name`, `service.version` resource attributes
- 返回 TracerProvider 供调用方注册到全局

### 3.2 middleware/ - 共享中间件

#### auth.go - JWT 认证中间件

提取各服务 `internal/middleware/auth.go` 中的公共部分，通过 Option 模式支持 variant：

```go
type AuthOptions struct {
    SigningMethodCheck bool
    AllowEmptyToken    bool
    ErrUnauthorized    *errors.Error  // 各服务自定义错误
}

type AuthOption func(*AuthOptions)

func WithSigningMethodCheck() AuthOption
func WithAllowEmptyToken() AuthOption
func WithUnauthorizedErr(err *errors.Error) AuthOption

func ServerAuth(secret string, opts ...AuthOption) middleware.Middleware
func GetAuthInfo(ctx context.Context) (*AuthInfo, bool)
func ExtractToken(ctx context.Context) string
```

**各服务接入示例**:
```go
// user-service, payment-service (Variant A)
middleware.ServerAuth(secret, middleware.WithSigningMethodCheck())

// order-service, cart-service, promotion-service (Variant B)
middleware.ServerAuth(secret)

// product-service, inventory-service (Variant C)
middleware.ServerAuth(secret, middleware.WithAllowEmptyToken())
```

#### response.go - 响应错误中间件

提取 `setStatusByReflection` 逻辑，参数化 ResponseStatus 创建：

```go
type ResponseStatusFactory func(code int32, msg string) interface{}

func ResponseError(newStatus ResponseStatusFactory) middleware.Middleware
```

**各服务接入示例**:
```go
// product-service
middleware.ResponseError(func(code int32, msg string) interface{} {
    return &pb.ResponseStatus{ErrorCode: code, ErrorMessage: msg}
})
```

#### tracing.go - 链路追踪中间件

```go
func ServerMiddleware(tracer trace.Tracer) middleware.Middleware
func ClientMiddleware(tracer trace.Tracer) middleware.Middleware
```

- Server: 从 incoming context 提取或创建 span，记录 method、peer、status
- Client: 创建 span 并注入到 outgoing context，确保 trace 传播

### 3.3 grpcutil/ - gRPC 工具

**client.go**:
```go
func NewInsecureClient(
    ctx context.Context,
    discovery registry.Discovery,
    endpoint string,
    opts ...grpc.DialOption,
) (*grpc.ClientConn, error)
```

**server.go** - Server Builder:
```go
type ServerBuilder struct { ... }

func NewServerBuilder() *ServerBuilder
func (b *ServerBuilder) WithAddress(addr string) *ServerBuilder
func (b *ServerBuilder) WithTimeout(timeout time.Duration) *ServerBuilder
func (b *ServerBuilder) WithMiddleware(mw ...middleware.Middleware) *ServerBuilder
func (b *ServerBuilder) RegisterService(register func(*grpc.Server)) *ServerBuilder
func (b *ServerBuilder) Build() *grpc.Server
```

### 3.4 data/ - 数据层工具

**gorm.go**:
```go
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error)
func ProvideDataConfig(c *conf.Config) *conf.Data
func NewDiscovery(endpoints []string) (registry.Discovery, error)
```

**gorm_tracing.go**:
```go
func NewGORMTracingPlugin(tracer trace.Tracer) gorm.Plugin
```

包装 `gorm.io/plugin/opentelemetry`，追踪 SQL 查询，span 名称为表名 + 操作。

### 3.5 kafka/ - Kafka 追踪

```go
type TracedWriter struct {
    inner  *kafka.Writer
    tracer trace.Tracer
}

func NewTracedProducer(inner *kafka.Writer, tracer trace.Tracer) *TracedWriter
func (w *TracedWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error

type TracedReader struct {
    inner  *kafka.Reader
    tracer trace.Tracer
}

func NewTracedConsumer(inner *kafka.Reader, tracer trace.Tracer) *TracedReader
func (r *TracedReader) FetchMessage(ctx context.Context) (kafka.Message, error)
```

- Producer: 创建 span，将 trace context 注入到 Kafka message headers
- Consumer: 从 message headers 提取 trace context，创建 child span

### 3.6 config/ - 基础配置结构体

**注意**：各服务已有自己的 `conf.*` 类型（部分为 proto 生成，部分手写），不强制替换。此模块提供：

1. **`Tracing` 配置结构体** - 供各服务嵌入到自身 Config 中：
```go
type Tracing struct {
    Enabled     bool
    Endpoint    string
    SampleRatio float64
}
```

2. **配置解析辅助函数** - 从 YAML 中提取 tracing 配置：
```go
func ExtractTracingConfig(raw map[string]interface{}) (*Tracing, error)
```

3. **可选的基础类型** - 新服务可直接使用，现有服务可保持不变：
```go
type Server struct { Grpc Grpc }
type Grpc struct { Network string; Addr string; Timeout time.Duration }
type Data struct { Database Database }
type Database struct { Driver string; Source string }
type Registry struct { Endpoints []string }
type Auth struct { JwtSecret string; TokenExpireHours int }
```

### 3.7 bootstrap/ - 应用启动引导

**注意**：各服务的 `conf.Bootstrap` 和 `conf.Config` 类型不同（部分为 proto 生成，部分为手写），因此使用泛型：

```go
type Options[B any, C any] struct {
    Name          string
    Version       string
    ConfigPath    string
    ExtractConfig func(*B) *C
    WireApp       func(*C, log.Logger) (*kratos.App, func(), error)
    LoggerFields  []interface{}
}

func Run[B any, C any](opts Options[B, C]) error
```

封装：
1. 加载本地 `configs/config.yaml` 到 `*B`（服务的 Bootstrap 类型）
2. 若配置了 `ConfigCenter`，从 etcd 加载远程配置覆盖
3. 调用 `ExtractConfig` 将 `*B` 转换为 `*C`（服务的 Config 类型）
4. 初始化 TracerProvider（若配置中启用）
5. 调用 `WireApp` 构建应用
6. 启动 `app.Run()`

**各服务接入示例**:
```go
func main() {
    err := bootstrap.Run(bootstrap.Options[conf.Bootstrap, conf.Config]{
        Name:          "product-service",
        Version:       "v1.0.0",
        ExtractConfig: extractConfig,
        WireApp:       wireApp,
    })
    if err != nil {
        panic(err)
    }
}
```

---

## 4. 各服务集成变更

### 4.1 go.mod 变更

每个服务添加：
```
require github.com/lgzzz/mall-tracing v0.1.0
```

### 4.2 配置变更

每个服务的 `configs/config.yaml` 添加：
```yaml
tracing:
  enabled: true
  endpoint: "localhost:4317"
  sample_ratio: 0.1
```

### 4.3 main.go 变更

**变更前**（以 product-service 为例，约 80 行）:
```go
func main() {
    // flag parse
    // logger setup
    // config loading (local + etcd)
    // wire app
    // app.Run
}
```

**变更后**（约 15 行）:
```go
func main() {
    err := bootstrap.Run(bootstrap.Options[conf.Bootstrap, conf.Config]{
        Name:          "product-service",
        Version:       "v1.0.0",
        ExtractConfig: extractConfig,
        WireApp:       wireApp,
    })
    if err != nil {
        panic(err)
    }
}
```

### 4.4 server/grpc.go 变更

**变更前**:
```go
func NewGRPCServer(c *conf.Config, svc *service.ProductService, logger log.Logger) *grpc.Server {
    var opts = []grpc.ServerOption{
        grpc.Middleware(
            recovery.Recovery(),
            middleware.ResponseError(),
            middleware.ServerAuth(c.Auth.JwtSecret),
        ),
    }
    // ...
}
```

**变更后**:
```go
func NewGRPCServer(c *conf.Config, svc *service.ProductService, logger log.Logger, tracer trace.Tracer) *grpc.Server {
    return grpcutil.NewServerBuilder().
        WithAddress(c.Server.Grpc.Addr).
        WithTimeout(c.Server.Grpc.Timeout).
        WithMiddleware(
            recovery.Recovery(),
            malltracing.ServerMiddleware(tracer),
            mallmiddleware.ResponseError(newResponseStatus),
            mallmiddleware.ServerAuth(c.Auth.JwtSecret),
        ).
        RegisterService(func(s *grpc.Server) { v1.RegisterProductServiceServer(s, svc) }).
        Build()
}
```

### 4.5 data/client.go 变更

**变更前**:
```go
func NewInventoryRepo(r registry.Discovery, endpoint string, logger log.Logger) biz.InventoryRepo {
    conn, err := grpc.DialInsecure(context.Background(), grpc.WithEndpoint(endpoint), grpc.WithDiscovery(r))
    // ...
}
```

**变更后**:
```go
func NewInventoryRepo(r registry.Discovery, endpoint string, logger log.Logger, tracer trace.Tracer) biz.InventoryRepo {
    conn, err := grpcutil.NewInsecureClient(context.Background(), r, endpoint,
        grpc.WithMiddleware(malltracing.ClientMiddleware(tracer)),
    )
    // ...
}
```

### 4.6 data/data.go 变更

**变更前**（约 20 行）:
```go
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
    cleanup := func() { log.NewHelper(logger).Info("closing the data resources") }
    db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
    if err != nil { return nil, nil, err }
    return &Data{db: db}, cleanup, nil
}
```

**变更后**（约 5 行）:
```go
func NewData(c *conf.Data, logger log.Logger, tracer trace.Tracer) (*Data, func(), error) {
    return malldata.NewData(c, logger, malldata.WithTracing(tracer))
}
```

### 4.7 服务变更影响矩阵

| 文件 | 变更前行数 | 变更后行数 | 节省 |
|------|-----------|-----------|------|
| `cmd/*/main.go` | ~80-120 | ~15 | ~65-105 |
| `internal/server/grpc.go` | ~30-37 | ~15 | ~15-22 |
| `internal/data/data.go` | ~20 | ~5 | ~15 |
| `internal/data/client.go` | ~20-40 | ~10 | ~10-30 |
| `internal/middleware/auth.go` | ~68-86 | 删除 | ~68-86 |
| `internal/middleware/error.go` | ~57-63 | 删除 | ~57-63 |
| **单服务总计** | **~275-366** | **~45** | **~230-321** |
| **8 个服务总计** | **~2200-2928** | **~360** | **~1840-2568** |

---

## 5. Docker Compose 变更

添加 Jaeger 服务到根目录 `docker-compose.yml`：

```yaml
services:
  jaeger:
    image: jaegertracing/all-in-one:1.67
    ports:
      - "16686:16686"   # Web UI
      - "4317:4317"     # OTLP gRPC
      - "4318:4318"     # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - mall-network
```

---

## 6. 链路传播示意

```
Browser → Gateway(HTTP) → order-service(gRPC) → product-service(gRPC)
  │           │                  │                     │
  │       span: HTTP         span: CreateOrder      span: GetProduct
  │           │                  │                     │
  │           │                  ├→ inventory-service  ├→ DB query
  │           │                  │   span: LockStock   span: SELECT
  │           │                  │   └→ DB query
  │           │                  │       span: LOCK
  │           │                  └→ Kafka
  │           │                      span: Publish OrderCreated
  │           │                  └→ cart-service
  │           │                      span: ClearCart
```

所有 span 共享同一个 `trace_id`，通过 gRPC metadata 和 Kafka headers 传播。

---

## 7. 实施顺序

### Phase 1: 创建 mall-tracing 仓库
1. 初始化 Go module
2. 实现 config/base.go
3. 实现 middleware/auth.go + response.go
4. 实现 grpcutil/client.go + server.go
5. 实现 data/gorm.go
6. 编写单元测试

### Phase 2: 添加 Tracing 功能
7. 实现 tracing/provider.go
8. 实现 middleware/tracing.go
9. 实现 data/gorm_tracing.go
10. 实现 kafka/producer.go + consumer.go
11. 编写 tracing 相关测试

### Phase 3: 发布到 GitHub
12. 创建 GitHub 仓库 `github.com/lgzzz/mall-tracing`
13. 推送代码，打 tag `v0.1.0`

### Phase 4: 集成到各服务（可并行）
14. 添加 `docker-compose.yml` Jaeger 服务
15. 逐个服务集成 mall-tracing
16. 验证链路追踪端到端工作

---

## 8. 测试策略

### mall-tracing 仓库
- **单元测试**: auth middleware 各 variant、response error factory、gRPC client creation
- **集成测试**: 使用 `testify` + mock gRPC server 验证 tracing middleware span 创建
- **Tracing 测试**: 使用 `go.opentelemetry.io/otel/sdk/trace/tracetest` 验证 span 导出

### 各服务集成后
- **手动验证**: 启动 Jaeger，发起跨服务请求，在 UI 中查看完整 trace
- **回归测试**: 确保现有功能测试仍通过

---

## 9. 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| mall-tracing API 变更影响所有服务 | 高 | 遵循 semver，breaking change 升主版本 |
| 采样率配置不当导致性能影响 | 中 | 默认 10%，生产环境可调低 |
| Jaeger 单点故障 | 低 | 本地开发用，生产用 OTLP collector 缓冲 |
| 共享库依赖版本冲突 | 中 | mall-tracing 使用与各服务兼容的依赖版本 |
