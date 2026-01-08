.PHONY: all build run test clean docker-build docker-run

# 变量
APP_NAME := docs-hub
GO := go
DOCKER := docker

# 默认目标
all: build

# 构建二进制文件
build:
	@echo "Building $(APP_NAME)..."
	$(GO) build -o bin/$(APP_NAME) ./cmd/main.go
	@echo "Build complete: bin/$(APP_NAME)"

# 本地运行（使用本地配置）
run-local:
	@echo "Running $(APP_NAME) with local config..."
	CONFIG_PATH=configs/local.yaml $(GO) run ./cmd/main.go

# 开发环境运行
run-dev:
	@echo "Running $(APP_NAME) with dev config..."
	CONFIG_PATH=configs/dev.yaml $(GO) run ./cmd/main.go

# 运行测试
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# 清理构建产物
clean:
	@echo "Cleaning..."
	rm -rf bin/
	@echo "Clean complete"

# Docker 构建
docker-build:
	@echo "Building Docker image..."
	$(DOCKER) build -t $(APP_NAME):latest .

# Docker 运行
docker-run:
	@echo "Running Docker container..."
	$(DOCKER) run -p 9000:9000 $(APP_NAME):latest

# 安装依赖
deps:
	@echo "Installing dependencies..."
	$(GO) mod download

# 代码格式化
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# 代码检查
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  all         - Build"
	@echo "  build       - Build binary"
	@echo "  run-local   - Run with local config"
	@echo "  run-dev     - Run with dev config"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run Docker container"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
