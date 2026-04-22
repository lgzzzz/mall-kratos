.PHONY: build run test clean protobuf wire

# 构建项目
build:
	@echo "Building..."
	@go build -o bin/server ./cmd/payment

# 运行服务
run:
	@echo "Starting service..."
	@go run ./cmd/payment

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 清理
clean:
	@rm -rf bin/
	@go clean

# 生成 Protobuf 代码
protobuf:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/payment/v1/payment.proto

# 生成 Wire 依赖注入代码
wire:
	@echo "Generating wire..."
	@cd cmd/payment && wire

# 初始化项目
init:
	@echo "Initializing project..."
	@go mod download
	@go mod tidy
