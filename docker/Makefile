# Solana ETL Docker Makefile

.PHONY: build test run clean help

# 默认目标
.DEFAULT_GOAL := help

# 变量
IMAGE_NAME := solana-etl
TAG := test
DOCKERFILE := Dockerfile
CONTEXT := ../
COMPOSE_FILE := docker-compose.yml

# 帮助信息
help: ## 显示帮助信息
	@echo "Solana ETL Docker build tool"
	@echo ""
	@echo "available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 构建 Docker 镜像
build: ## 构建 Docker 镜像
	@echo "build docker image: $(IMAGE_NAME):$(TAG)"
	docker build -f $(DOCKERFILE) -t $(IMAGE_NAME):$(TAG) $(CONTEXT)
	@echo "build image done"

# 测试镜像
test: ## 测试 Docker 镜像
	@echo "测试 Docker 镜像..."
	@echo "1. test help command"
	docker run --rm $(IMAGE_NAME):$(TAG) ./solana-etl --help
	@echo ""
	@echo "2. test status command"
	docker run --rm $(IMAGE_NAME):$(TAG) ./solana-etl status

# 运行基本监控
run: ## 运行 USDC 监控示例
	@echo "run USDC monitor..."
	docker run --rm -p 20000:20000 $(IMAGE_NAME):$(TAG) \
		./solana-etl start \
		--mint-account="EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" \
		--verbose \
		--exporter=:20000

# 运行多币种监控
run-multi: ## 运行多币种监控 (USDC + USDT)
	@echo "run multi-currency monitor..."
	docker run --rm -p 20000:20000 $(IMAGE_NAME):$(TAG) \
		./solana-etl start \
		--mint-account="EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB" \
		--verbose \
		--exporter=:20000

# 使用 Docker Compose 启动
compose-up: ## 使用 Docker Compose 启动服务
	docker-compose -f $(COMPOSE_FILE) up solana-etl

# 使用 Docker Compose 启动监控服务
compose-monitoring: ## 使用 Docker Compose 启动 ETL + 监控服务
	docker-compose -f $(COMPOSE_FILE) --profile monitoring up -d

# 停止 Docker Compose 服务
compose-down: ## 停止 Docker Compose 服务
	docker-compose -f $(COMPOSE_FILE) down

# 查看 Docker Compose 日志
compose-logs: ## 查看 Docker Compose 日志
	docker-compose -f $(COMPOSE_FILE) logs -f solana-etl

# 清理 Docker 镜像
clean: ## 清理 Docker 镜像
	@echo "clean image..."
	docker rmi $(IMAGE_NAME):$(TAG) || true
	@echo "clean image done"

# 清理所有相关容器和镜像
clean-all: ## 清理所有相关容器和镜像
	@echo "clean all..."
	docker-compose -f $(COMPOSE_FILE) down --rmi all --volumes --remove-orphans || true
	docker rmi $(IMAGE_NAME):$(TAG) || true
	@echo "clean all done"

# 显示镜像信息
info: ## 显示镜像信息
	@echo "Docker image information:"
	docker images $(IMAGE_NAME):$(TAG) || echo "镜像不存在"
	@echo ""
	@echo "Docker Compose 服务状态:"
	docker-compose -f $(COMPOSE_FILE) ps || echo "服务未运行"

# 进入容器
shell: ## 进入容器 shell
	docker run --rm -it $(IMAGE_NAME):$(TAG) /bin/sh

# 检查健康状态
health: ## 检查容器健康状态
	@echo "health check..."
	docker run --rm -p 20000:20000 -d --name solana-etl-health $(IMAGE_NAME):$(TAG) \
		./solana-etl start \
		--mint-account="EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" \
		--exporter=:20000
	@sleep 10
	@echo "检查 metrics 端点..."
	@curl -s http://localhost:20000/metrics | head -10 || echo "Metrics 端点不可访问"
	@docker stop solana-etl-health || true
	@echo "health check done"
