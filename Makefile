# Solana-Go 项目 Makefile
# 提供标准的构建、测试和清理命令

VERSION := $(shell git describe --tags --match *-criptobox | sed -e 's/^\(v[0-9]\{1,\}\.[0-9]\{1,\}\)\.[0-9a-x]\{1,\}\(-.*-g[0-9a-f]\{1,\}\)$$/\1.\2/' | sed 's/-criptobox-//')

.PHONY: all build test clean install deps help

# 默认目标
all: deps test build

# 帮助信息
help:
	@echo "Solana-Go 项目构建命令:"
	@echo "  make deps     - 下载依赖"
	@echo "  make test     - 运行测试"
	@echo "  make build    - 构建项目"
	@echo "  make clean    - 清理构建文件"
	@echo "  make install  - 安装到 GOPATH/bin"
	@echo "  make all      - 执行 deps + test + build"

# 下载依赖
#deps:
#	@echo "下载 Go 模块依赖..."
#	go mod download
#	go mod tidy
#	go mod vendor

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 构建项目
build: build-cli build-examples

# 构建 CLI 工具
build-cli:
	# @echo "构建 CLI 工具 (slnc)..."
	# go build -o slnc ./cmd/slnc/
	@echo "构建 CLI 工具 (solana-etl)..."
	go build -o ./build/solana-etl ./cmd/solana-etl/

# 构建示例程序
build-examples:
	@echo "构建示例程序..."
	# go build -o example-getBalance ./rpc/examples/getBalance/
	# go build -o example-getAccountInfo ./rpc/examples/getAccountInfo/

# 安装到 GOPATH/bin
install:
	# @echo "安装 CLI 工具到 GOPATH/bin..."
	# go install ./cmd/slnc/
	@echo "构建 CLI 工具 (solana-etl)..."
	# go build -o ./build/solana-etl ./cmd/solana-etl/
	# sudo mv ./build/solana-etl $(shell go env GOBIN)/
	echo ${VERSION}
	go install --mod=vendor ./cmd/solana-etl/

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f slnc example-*

# 交叉编译
build-linux:
	@echo "构建 Linux 版本..."
	GOOS=linux GOARCH=amd64 go build -o slnc-linux ./cmd/slnc/

build-windows:
	@echo "构建 Windows 版本..."
	GOOS=windows GOARCH=amd64 go build -o slnc-windows.exe ./cmd/slnc/

build-darwin:
	@echo "构建 macOS 版本..."
	GOOS=darwin GOARCH=amd64 go build -o slnc-darwin ./cmd/slnc/

# 构建所有平台
build-all: build-linux build-windows build-darwin

# 代码格式化
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
vet:
	@echo "代码检查..."
	go vet ./...

# 生成文档
doc:
	@echo "生成文档..."
	go doc -all ./...

# 运行基准测试
bench:
	@echo "运行基准测试..."
	go test -bench=. ./...
