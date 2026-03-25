# ========== 全局变量 ==========
BIN_DIR := bin
MK_DIR := .mk

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version 2>/dev/null | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# ========== 引入共享规则 ==========
-include $(MK_DIR)/lint.mk
-include $(MK_DIR)/format.mk
-include $(MK_DIR)/coverage.mk

# ========== 构建目标 ==========
## build: 构建 watch 二进制文件（默认）
build: build-watch

## build-%: 构建指定二进制文件（如 build-worker）
build-%:
	@echo "Building $*..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$* ./cmd/$*

# ========== 测试目标 ==========
## test: 运行测试
test:
	@echo "Running tests..."
	$(GOTEST) ./...

## test-verbose: 运行测试（详细输出）
test-verbose:
	@echo "Running tests (verbose)..."
	$(GOTEST) -v ./...

# ========== 清理目标 ==========
## clean: 清理构建产物和临时文件
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)/
	@rm -f coverage.out coverage.html

# ========== 运行目标 ==========
## run: 构建并运行 watch
run: build-watch
	@./$(BIN_DIR)/watch

# ========== 帮助目标 ==========
## help: 显示帮助信息
help:
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -hE '^## [a-zA-Z_%-]+: ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ": "}; {printf "  \033[36m%-20s\033[0m %s\n", substr($$1, 4), $$2}'

# ========== .PHONY 声明（集中管理所有伪目标）==========
.PHONY: version build build-% test test-verbose clean run help
.PHONY: lint lint-fix fmt vet coverage coverage-html
.PHONY: docker-up docker-down docker-logs docker-ps docker-reset

# ==================== Docker ====================

## docker-up: 启动所有 Docker 服务
docker-up:
	@echo "Starting Docker services..."
	@cp -n .env.example .env 2>/dev/null || true
	docker-compose up -d
	@echo "Services started. Grafana: http://localhost:$$(grep GRAFANA_PORT .env 2>/dev/null | cut -d'=' -f2 || echo 3000)"

## docker-down: 停止所有 Docker 服务
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "Services stopped."

## docker-logs: 查看 Docker 服务日志
docker-logs:
	docker-compose logs -f

## docker-ps: 查看 Docker 服务状态
docker-ps:
	docker-compose ps

## docker-reset: 重置所有 Docker 数据并重启
docker-reset:
	@echo "Resetting Docker data..."
	docker-compose down -v
	@rm -rf docker/data/*
	@touch docker/data/.gitkeep
	@echo "Starting services..."
	docker-compose up -d
	@echo "Reset complete."
