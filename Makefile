# Makefile for 抖音RTMP推流地址获取工具

# 项目信息
PROJECT_NAME := douyin-rtmp
VERSION := v1.0.0
AUTHOR := Go版本

# Go相关设置
GO_VERSION := $(shell go version | awk '{print $$3}' | sed 's/go//')
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 构建设置
BUILD_DIR := build
DIST_DIR := dist
BINARY_NAME := $(PROJECT_NAME)
LDFLAGS := -s -w -X main.AppVersion=$(VERSION)

# 支持的平台
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	darwin/amd64 \
	darwin/arm64

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m

.PHONY: all clean deps test build build-all package install uninstall help

# 默认目标
all: clean deps test build

# 显示项目信息
info:
	@echo "$(BLUE)=== $(PROJECT_NAME) 构建系统 ===$(NC)"
	@echo "$(BLUE)版本: $(VERSION)$(NC)"
	@echo "$(BLUE)作者: $(AUTHOR)$(NC)"
	@echo "$(BLUE)Go版本: $(GO_VERSION)$(NC)"
	@echo "$(BLUE)目标平台: $(GOOS)/$(GOARCH)$(NC)"
	@echo ""

# 清理构建文件
clean:
	@echo "$(YELLOW)清理构建文件...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME).exe
	@echo "$(GREEN)清理完成$(NC)"

# 下载依赖
deps:
	@echo "$(YELLOW)下载Go依赖...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)依赖下载完成$(NC)"

# 运行测试
test:
	@echo "$(YELLOW)运行单元测试...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)测试完成$(NC)"

# 运行基准测试
benchmark:
	@echo "$(YELLOW)运行基准测试...$(NC)"
	@go test -bench=. -benchmem ./...

# 代码静态检查
lint:
	@echo "$(YELLOW)运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)警告: golangci-lint 未安装，跳过代码检查$(NC)"; \
	fi

# 格式化代码
fmt:
	@echo "$(YELLOW)格式化代码...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)代码格式化完成$(NC)"

# 构建当前平台
build: deps
	@echo "$(YELLOW)构建 $(GOOS)/$(GOARCH)...$(NC)"
	@mkdir -p $(BUILD_DIR)/$(GOOS)_$(GOARCH)
	@if [ "$(GOOS)" = "windows" ]; then \
		go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(GOOS)_$(GOARCH)/$(BINARY_NAME).exe .; \
	else \
		go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(GOOS)_$(GOARCH)/$(BINARY_NAME) .; \
	fi
	@if [ -d "assets" ]; then \
		cp -r assets $(BUILD_DIR)/$(GOOS)_$(GOARCH)/; \
	fi
	@echo "$(GREEN)构建完成$(NC)"

# 构建所有平台
build-all: deps
	@echo "$(YELLOW)构建所有平台版本...$(NC)"
	@$(foreach platform,$(PLATFORMS), \
		$(eval os := $(word 1,$(subst /, ,$(platform)))) \
		$(eval arch := $(word 2,$(subst /, ,$(platform)))) \
		echo "$(BLUE)构建 $(os)/$(arch)...$(NC)"; \
		mkdir -p $(BUILD_DIR)/$(os)_$(arch); \
		if [ "$(os)" = "windows" ]; then \
			GOOS=$(os) GOARCH=$(arch) go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(os)_$(arch)/$(BINARY_NAME).exe .; \
		else \
			GOOS=$(os) GOARCH=$(arch) go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(os)_$(arch)/$(BINARY_NAME) .; \
		fi; \
		if [ -d "assets" ]; then \
			cp -r assets $(BUILD_DIR)/$(os)_$(arch)/; \
		fi; \
	)
	@echo "$(GREEN)所有平台构建完成$(NC)"

# 创建发布包
package: build-all
	@echo "$(YELLOW)创建发布包...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for dir in $(BUILD_DIR)/*/; do \
		if [ -d "$$dir" ]; then \
			platform=$$(basename "$$dir"); \
			echo "$(BLUE)打包 $$platform...$(NC)"; \
			cd $(BUILD_DIR) && \
			if [[ "$$platform" == *"windows"* ]]; then \
				zip -r ../$(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)_$$platform.zip $$platform/; \
			else \
				tar -czf ../$(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)_$$platform.tar.gz $$platform/; \
			fi && \
			cd ..; \
		fi \
	done
	@echo "$(GREEN)发布包创建完成，位于 $(DIST_DIR)/ 目录$(NC)"

# 安装到系统
install: build
	@echo "$(YELLOW)安装到系统...$(NC)"
	@if [ ! -f "$(BUILD_DIR)/$(GOOS)_$(GOARCH)/$(BINARY_NAME)" ]; then \
		echo "$(RED)错误: 未找到构建文件$(NC)"; \
		exit 1; \
	fi
	@sudo cp $(BUILD_DIR)/$(GOOS)_$(GOARCH)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)安装完成: /usr/local/bin/$(BINARY_NAME)$(NC)"

# 从系统卸载
uninstall:
	@echo "$(YELLOW)从系统卸载...$(NC)"
	@if [ -f "/usr/local/bin/$(BINARY_NAME)" ]; then \
		sudo rm /usr/local/bin/$(BINARY_NAME); \
		echo "$(GREEN)卸载完成$(NC)"; \
	else \
		echo "$(YELLOW)未找到已安装的程序$(NC)"; \
	fi

# 运行程序
run: build
	@echo "$(YELLOW)运行程序...$(NC)"
	@./$(BUILD_DIR)/$(GOOS)_$(GOARCH)/$(BINARY_NAME)

# 开发模式运行
dev:
	@echo "$(YELLOW)开发模式运行...$(NC)"
	@go run . -debug

# 创建Docker镜像
docker-build:
	@echo "$(YELLOW)构建Docker镜像...$(NC)"
	@docker build -t $(PROJECT_NAME):$(VERSION) .
	@docker tag $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest
	@echo "$(GREEN)Docker镜像构建完成$(NC)"

# 生成依赖图
deps-graph:
	@echo "$(YELLOW)生成依赖图...$(NC)"
	@if command -v godepgraph >/dev/null 2>&1; then \
		godepgraph -s . | dot -Tpng -o deps.png; \
		echo "$(GREEN)依赖图已生成: deps.png$(NC)"; \
	else \
		echo "$(YELLOW)警告: godepgraph 未安装$(NC)"; \
	fi

# 检查更新
check-updates:
	@echo "$(YELLOW)检查依赖更新...$(NC)"
	@go list -u -m all

# 显示帮助
help:
	@echo "$(BLUE)可用目标:$(NC)"
	@echo "  info          显示项目信息"
	@echo "  clean         清理构建文件"
	@echo "  deps          下载Go依赖"
	@echo "  test          运行单元测试"
	@echo "  benchmark     运行基准测试"
	@echo "  lint          代码静态检查"
	@echo "  fmt           格式化代码"
	@echo "  build         构建当前平台"
	@echo "  build-all     构建所有平台"
	@echo "  package       创建发布包"
	@echo "  install       安装到系统"
	@echo "  uninstall     从系统卸载"
	@echo "  run           运行程序"
	@echo "  dev           开发模式运行"
	@echo "  docker-build  构建Docker镜像"
	@echo "  deps-graph    生成依赖图"
	@echo "  check-updates 检查依赖更新"
	@echo "  help          显示此帮助"
	@echo ""
	@echo "$(BLUE)示例:$(NC)"
	@echo "  make build        # 构建当前平台"
	@echo "  make build-all    # 构建所有平台"
	@echo "  make package      # 创建发布包"
	@echo "  make install      # 安装到系统"