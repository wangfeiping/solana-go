package config

import (
	"context"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/time/rate"
)

// Config contains all configuration options for the Solana ETL tool
type Config struct {
	WSSURL       string   // WebSocket URL
	RPCURL       string   // RPC URL
	StartBlock   uint64   // Start monitoring from this block number (slot)
	MintAccount  string   // 原始输入字符串
	MintAccounts []string // 解析后的 mint 账户列表
	Commitment   string   // Commitment level: processed, confirmed, or finalized
	Verbose      bool     // Enable verbose logging
	Exporter     string   // Prometheus exporter 地址
	MonitorSOL   bool     // 是否监听SOL转账
	RPCLimitRate int      // RPC调用频率限制 (每秒请求数)
	QueueSize    int      // 消息队列大小
}

// System Program ID for SOL transfers
const SystemProgramID = "11111111111111111111111111111111"

// NewConfig creates a new config with default values
func NewConfig() *Config {
	return &Config{
		WSSURL:       "wss://api.mainnet-beta.solana.com",
		RPCURL:       "https://api.mainnet-beta.solana.com",
		StartBlock:   0,
		Commitment:   "confirmed",
		Verbose:      false,
		Exporter:     ":20000",
		MonitorSOL:   false,
		RPCLimitRate: 10,
		QueueSize:    2000,
	}
}

// ParseMintAccounts 解析逗号分隔的 mint 账户字符串
func (c *Config) ParseMintAccounts() {
	if c.MintAccount == "" {
		return
	}

	// 按逗号分割并清理空白字符
	accounts := strings.Split(c.MintAccount, ",")
	result := make([]string, 0, len(accounts))

	for _, account := range accounts {
		account = strings.TrimSpace(account)
		if account != "" {
			result = append(result, account)
		}
	}

	c.MintAccounts = result
}

// CheckForSystemProgram 检查是否包含System Program ID，如果包含则自动启用SOL监听
func (c *Config) CheckForSystemProgram() {
	for _, mintAccount := range c.MintAccounts {
		if mintAccount == SystemProgramID {
			c.MonitorSOL = true
			break
		}
	}
}

// RateLimitedRPCClient wraps the RPC client with rate limiting
type RateLimitedRPCClient struct {
	client  *rpc.Client
	limiter *rate.Limiter
}

// NewRateLimitedRPCClient creates a new rate-limited RPC client
func NewRateLimitedRPCClient(rpcURL string, rps int) *RateLimitedRPCClient {
	client := rpc.New(rpcURL)
	limiter := rate.NewLimiter(rate.Limit(rps), rps) // Allow burst up to rps

	return &RateLimitedRPCClient{
		client:  client,
		limiter: limiter,
	}
}

// GetTransaction wraps the RPC call with rate limiting
func (r *RateLimitedRPCClient) GetTransaction(ctx context.Context, signature solana.Signature, opts *rpc.GetTransactionOpts) (*rpc.GetTransactionResult, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return r.client.GetTransaction(ctx, signature, opts)
}

// GetAccountInfo wraps the RPC call with rate limiting
func (r *RateLimitedRPCClient) GetAccountInfo(ctx context.Context, account solana.PublicKey) (*rpc.GetAccountInfoResult, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return r.client.GetAccountInfo(ctx, account)
}

// GetHealth wraps the RPC call with rate limiting
func (r *RateLimitedRPCClient) GetHealth(ctx context.Context) (string, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return "", err
	}
	return r.client.GetHealth(ctx)
}

// GetSlot wraps the RPC call with rate limiting
func (r *RateLimitedRPCClient) GetSlot(ctx context.Context, commitment rpc.CommitmentType) (uint64, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return 0, err
	}
	return r.client.GetSlot(ctx, commitment)
}

// GetEpochInfo wraps the RPC call with rate limiting
func (r *RateLimitedRPCClient) GetEpochInfo(ctx context.Context, commitment rpc.CommitmentType) (*rpc.GetEpochInfoResult, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return r.client.GetEpochInfo(ctx, commitment)
}

// Close closes the underlying RPC client
func (r *RateLimitedRPCClient) Close() error {
	return r.client.Close()
}
