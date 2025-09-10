# Solana-Go 项目构建指南

## 项目概述

solana-go 是一个用 Go 语言编写的 Solana 区块链 SDK 库，提供了与 Solana JSON RPC 和 WebSocket 接口交互的完整功能。

## 系统要求

- **Go 版本**: Go 1.19 或更高版本
- **操作系统**: Linux, macOS, Windows
- **内存**: 至少 2GB RAM
- **磁盘空间**: 至少 1GB 可用空间

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/gagliardetto/solana-go.git
cd solana-go
```

### 2. 使用 Makefile 构建（推荐）

```bash
# 查看所有可用命令
make help

# 下载依赖并构建
make all

# 或者分步执行
make deps    # 下载依赖
make test    # 运行测试
make build   # 构建项目
```

### 3. 使用构建脚本

```bash
# 使用提供的构建脚本
chmod +x build.sh
./build.sh
```

### 4. 手动构建

```bash
# 下载依赖
go mod download
go mod tidy

# 运行测试
go test -v ./...

# 构建 CLI 工具
go build -o slnc ./cmd/slnc/

# 构建示例程序
go build -o example ./rpc/examples/getBalance/
```

## 构建选项

### 构建目标

1. **CLI 工具 (slnc)**: 命令行工具，提供 Solana 区块链交互功能
2. **示例程序**: 各种 RPC 调用的示例程序
3. **库文件**: 作为 Go 模块供其他项目使用

### 交叉编译

```bash
# Linux
make build-linux

# Windows
make build-windows

# macOS
make build-darwin

# 所有平台
make build-all
```

### 安装到系统

```bash
# 安装到 GOPATH/bin
make install

# 或者直接使用 go install
go install ./cmd/slnc/
```

## 项目结构

```
solana-go/
├── cmd/slnc/              # CLI 工具源码
├── rpc/                   # RPC 客户端实现
│   ├── examples/          # 示例程序
│   └── ws/               # WebSocket 实现
├── programs/              # 各种程序客户端
│   ├── system/           # 系统程序
│   ├── token/            # SPL Token 程序
│   ├── serum/            # Serum DEX 程序
│   └── ...
├── text/                  # 文本格式化工具
├── Makefile              # 构建配置
├── build.sh              # 构建脚本
└── go.mod                # Go 模块配置
```

## 开发环境设置

### 1. 安装 Go

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# macOS (使用 Homebrew)
brew install go

# Windows
# 从 https://golang.org/dl/ 下载安装包
```

### 2. 设置环境变量

```bash
# 设置 GOPATH
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# 设置 Go 代理（可选，用于加速下载）
export GOPROXY=https://goproxy.cn,direct
```

### 3. 验证安装

```bash
go version
go env
```

## 测试

### 运行所有测试

```bash
make test
# 或者
go test -v ./...
```

### 运行特定包的测试

```bash
go test -v ./rpc
go test -v ./programs/token
```

### 运行基准测试

```bash
make bench
# 或者
go test -bench=. ./...
```

## 代码质量

### 代码格式化

```bash
make fmt
# 或者
go fmt ./...
```

### 代码检查

```bash
make vet
# 或者
go vet ./...
```

### 生成文档

```bash
make doc
# 或者
go doc -all ./...
```

## 故障排除

### 常见问题

1. **Go 版本过低**
   ```
   go: requires go >= 1.19
   ```
   解决方案：升级到 Go 1.19 或更高版本

2. **网络问题导致依赖下载失败**
   ```bash
   # 设置 Go 代理
   export GOPROXY=https://goproxy.cn,direct
   go mod download
   ```

3. **权限问题**
   ```bash
   # 确保有执行权限
   chmod +x build.sh
   chmod +x slnc
   ```

4. **内存不足**
   - 增加系统内存
   - 使用 `-ldflags="-s -w"` 减少二进制文件大小

### 调试构建

```bash
# 详细构建信息
go build -v -x ./cmd/slnc/

# 检查依赖
go mod graph

# 清理模块缓存
go clean -modcache
```

## 发布构建

### 使用 GoReleaser

项目配置了 GoReleaser 用于发布构建：

```bash
# 安装 GoReleaser
go install github.com/goreleaser/goreleaser@latest

# 本地测试发布
goreleaser release --snapshot

# 正式发布
goreleaser release
```

### 手动发布构建

```bash
# 构建所有平台
make build-all

# 创建发布包
tar -czf solana-go-linux.tar.gz slnc-linux
zip solana-go-windows.zip slnc-windows.exe
tar -czf solana-go-darwin.tar.gz slnc-darwin
```

## 贡献

### 开发流程

1. Fork 项目
2. 创建功能分支
3. 编写代码和测试
4. 运行测试确保通过
5. 提交 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 编写单元测试
- 添加适当的注释

## 许可证

本项目使用 Apache 2.0 许可证。详见 [LICENSE](LICENSE) 文件。

## 支持

- 项目主页: https://github.com/gagliardetto/solana-go
- 问题报告: https://github.com/gagliardetto/solana-go/issues
- 文档: https://pkg.go.dev/github.com/gagliardetto/solana-go
