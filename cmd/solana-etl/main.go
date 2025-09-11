package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/etl/exporter"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

type Config struct {
	ProviderURL  string
	StartBlock   uint64
	MintAccount  string   // 原始输入字符串
	MintAccounts []string // 解析后的 mint 账户列表
	Commitment   string
	Verbose      bool
	Exporter     string // Prometheus exporter 地址
}

// Global metrics instance
var metrics *exporter.Metrics

// Global RPC client for getting transaction details
var rpcClient *rpc.Client

var (
	cfg     = &Config{}
	rootCmd = &cobra.Command{
		Use:   "solana-etl",
		Short: "Solana WebSocket event monitor for mint account transactions",
		Long: `A real-time Solana transaction monitor that uses WebSocket subscriptions
to track transactions related to specific mint accounts.

This tool connects to a Solana WebSocket provider and monitors all transactions
that mention the specified mint accounts, providing real-time notifications
with transaction details, logs, and operation analysis.`,
	}
)

func init() {
	// Add commands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)

	// Global flags for all commands
	rootCmd.PersistentFlags().StringVarP(&cfg.ProviderURL, "provider", "p",
		"wss://api.mainnet-beta.solana.com",
		"Solana WebSocket provider URL")

	rootCmd.PersistentFlags().Uint64VarP(&cfg.StartBlock, "start-block", "s", 0,
		"Start monitoring from this block number (slot)")

	// 支持多个 mint 账户，用逗号分隔
	rootCmd.PersistentFlags().StringVarP(&cfg.MintAccount, "mint-account", "m", "",
		"Mint account address(es) to monitor (comma-separated for multiple accounts)")

	rootCmd.PersistentFlags().StringVarP(&cfg.Commitment, "commitment", "c", "confirmed",
		"Commitment level: processed, confirmed, or finalized")

	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false,
		"Enable verbose logging")

	// 添加 exporter 参数
	rootCmd.PersistentFlags().StringVar(&cfg.Exporter, "exporter", ":20000",
		"Prometheus exporter address in format ip:port for metrics export")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start WebSocket subscription and monitor transactions",
	Long: `Start the WebSocket subscription to monitor transactions related to specific mint accounts.
This command will connect to the Solana WebSocket provider and continuously parse
and analyze transactions mentioning the specified mint accounts.`,
	Run: runStartCommand,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of Solana network and mint accounts",
	Long: `Check the current status of the Solana network and validate the specified mint accounts.
This command will connect to the RPC endpoint and retrieve information about
the network status and mint account details.`,
	Run: runStatusCommand,
}

func main() {
	// Initialize metrics
	metrics = exporter.NewMetrics()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// startMetricsServer 启动 Prometheus metrics HTTP 服务器
func startMetricsServer(ctx context.Context, addr string) {
	if addr == "" {
		return
	}

	log.Printf("Starting Prometheus metrics server on %s", addr)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down metrics server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down metrics server: %v", err)
		}
	}()
}

func runStartCommand(cmd *cobra.Command, args []string) {
	// Parse mint accounts
	if cfg.MintAccount == "" {
		log.Fatal("Error: mint-account is required. Use -m or --mint-account flag.")
	}

	// 解析多个 mint 账户
	cfg.MintAccounts = parseMintAccounts(cfg.MintAccount)
	if len(cfg.MintAccounts) == 0 {
		log.Fatal("Error: no valid mint accounts provided")
	}

	// Validate mint accounts
	mintPubkeys := make([]solana.PublicKey, 0, len(cfg.MintAccounts))
	for i, mintAccount := range cfg.MintAccounts {
		mintPubkey, err := solana.PublicKeyFromBase58(mintAccount)
		if err != nil {
			log.Fatalf("Invalid mint account address at position %d: %s - %v", i+1, mintAccount, err)
		}
		mintPubkeys = append(mintPubkeys, mintPubkey)
	}

	log.Printf("Monitoring %d mint account(s):", len(cfg.MintAccounts))
	for i, mintAccount := range cfg.MintAccounts {
		log.Printf("  [%d] %s", i+1, mintAccount)
	}

	// Initialize RPC client for getting transaction details
	rpcURL := cfg.ProviderURL
	if rpcURL[:3] == "wss" {
		rpcURL = "https" + rpcURL[3:]
	} else if rpcURL[:2] == "ws" {
		rpcURL = "http" + rpcURL[2:]
	}
	rpcClient = rpc.New(rpcURL)
	log.Printf("Initialized RPC client: %s", rpcURL)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics server if exporter is configured
	if cfg.Exporter != "" {
		startMetricsServer(ctx, cfg.Exporter)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, closing connections...")
		cancel()
	}()

	// Connect to WebSocket
	log.Printf("Connecting to Solana WebSocket at: %s", cfg.ProviderURL)
	client, err := ws.Connect(ctx, cfg.ProviderURL)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
		metrics.SetConnectionStatus("solana", 0)
	} else {
		metrics.SetConnectionStatus("solana", 1)
	}
	defer client.Close()

	// Set commitment level
	commitment := rpc.CommitmentConfirmed
	if cfg.Commitment != "" {
		switch cfg.Commitment {
		case "processed":
			commitment = rpc.CommitmentProcessed
		case "confirmed":
			commitment = rpc.CommitmentConfirmed
		case "finalized":
			commitment = rpc.CommitmentFinalized
		default:
			log.Printf("Unknown commitment level: %s, using confirmed", cfg.Commitment)
		}
	}

	// 为每个 mint 账户创建订阅
	var wg sync.WaitGroup
	subscriptions := make([]*ws.LogSubscription, 0, len(mintPubkeys))

	for i, mintPubkey := range mintPubkeys {
		log.Printf("Subscribing to transactions mentioning mint account [%d]: %s", i+1, cfg.MintAccounts[i])
		subscription, err := client.LogsSubscribeMentions(mintPubkey, commitment)
		if err != nil {
			log.Fatalf("Failed to subscribe to logs for mint %s: %v", cfg.MintAccounts[i], err)
		}
		subscriptions = append(subscriptions, subscription)
		defer subscription.Unsubscribe()

		// Update metrics
		metrics.SetSubscriptionActive("solana", 1)

		// 为每个订阅启动一个 goroutine
		wg.Add(1)
		go func(sub *ws.LogSubscription, mintAccount string, mintIndex int) {
			defer wg.Done()
			defer metrics.SetSubscriptionActive("solana", 0)
			processSubscription(ctx, sub, mintAccount, mintIndex)
		}(subscription, cfg.MintAccounts[i], i+1)
	}

	log.Printf("Starting to monitor transactions from block %d...", cfg.StartBlock)
	log.Println("Press Ctrl+C to stop monitoring")

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("All subscriptions stopped")
}

func processSubscription(ctx context.Context, subscription *ws.LogSubscription, mintAccount string, mintIndex int) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Subscription [%d] cancelled for mint %s", mintIndex, mintAccount)
			return
		default:
			// Set a timeout for receiving messages
			recvCtx, recvCancel := context.WithTimeout(ctx, 30*time.Second)
			logResult, err := subscription.Recv(recvCtx)
			recvCancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					// No message received within timeout, continue
					continue
				}
				if err == context.Canceled {
					log.Printf("Subscription [%d] cancelled for mint %s", mintIndex, mintAccount)
					return
				}
				log.Printf("Error receiving message for mint [%d] %s: %v", mintIndex, mintAccount, err)
				continue
			}

			// Process the log result
			processTransaction(logResult, mintAccount, mintIndex)
		}
	}
}

func processTransaction(logResult *ws.LogResult, mintAccount string, mintIndex int) {
	// 更新区块高度监控指标
	// 在 Solana 中，slot 和 block height 是相关的概念
	// 这里使用 slot 作为区块高度的近似值
	metrics.UpdateBlockHeight("solana", float64(logResult.Context.Slot))

	// Check if transaction was successful
	if logResult.Value.Err != nil {
		metrics.RecordTransaction("solana", mintAccount, "failed", "unknown")
		if cfg.Verbose {
			log.Printf("❌ Failed transaction for mint [%d] %s: %s (Error: %v)",
				mintIndex, mintAccount, logResult.Value.Signature.String(), logResult.Value.Err)
		}
		return
	}

	// // Check if we're past the start block
	// if logResult.Context.Slot < cfg.StartBlock {
	// 	return
	// }

	// Update successful transaction metric
	metrics.RecordTransaction("solana", mintAccount, "success", "unknown")

	// Log the transaction details
	// log.Printf("✅ Transaction found for mint [%d] %s:", mintIndex, mintAccount)
	log.Printf("   Signature: %s", logResult.Value.Signature.String())
	// log.Printf("   Slot: %d", logResult.Context.Slot)
	log.Printf("   Block Height: %d", logResult.Context.Slot)
	// log.Printf("   Timestamp: %s", time.Now().Format(time.RFC3339))

	// Parse transfer information from logs
	// Get complete transaction details including from/to addresses
	transfers := getTransactionDetails(logResult.Value.Signature, logResult.Value.Logs)

	// Add debug logging
	if cfg.Verbose {
		log.Printf("   🔍 Debug: Found %d transfers from parsing", len(transfers))
		if len(transfers) == 0 {
			log.Printf("   🔍 Debug: No transfers found, checking for TransferChecked instruction...")
			for i, logMsg := range logResult.Value.Logs {
				if strings.Contains(logMsg, "TransferChecked") {
					log.Printf("    Debug: Found TransferChecked in log [%d]: %s", i+1, logMsg)
				}
			}
		}
	}

	if len(transfers) > 0 {
		// log.Printf("%d ", logResult.Context.Slot)
		log.Printf("   💸 Transfer Details:")
		for i, transfer := range transfers {
			log.Printf("     Transfer [%d]:", i+1)
			log.Printf("       From: %s (%s)", formatAddress(transfer.From), transfer.From)
			if transfer.FromOwner != "" && transfer.FromOwner != "unknown" && transfer.FromOwner != "error" && transfer.FromOwner != "not_found" && transfer.FromOwner != "invalid" && transfer.FromOwner != "not_token_account" && transfer.FromOwner != "invalid_data" {
				log.Printf("         Owner: %s (%s)", formatAddress(transfer.FromOwner), transfer.FromOwner)
			}
			log.Printf("       To: %s (%s)", formatAddress(transfer.To), transfer.To)
			if transfer.ToOwner != "" && transfer.ToOwner != "unknown" && transfer.ToOwner != "error" && transfer.ToOwner != "not_found" && transfer.ToOwner != "invalid" && transfer.ToOwner != "not_token_account" && transfer.ToOwner != "invalid_data" {
				log.Printf("         Owner: %s (%s)", formatAddress(transfer.ToOwner), transfer.ToOwner)
			}
			if transfer.Amount != "" {
				log.Printf("       Amount: %s", transfer.Amount)
			}
			if transfer.Mint != "" {
				log.Printf("       Mint: %s (%s)", formatAddress(transfer.Mint), transfer.Mint)
			}
		}
	}

	// Parse and display all account addresses involved
	addresses := parseAccountAddresses(logResult.Value.Logs)
	if len(addresses) > 0 {
		log.Printf("   📋 Account Addresses:")
		for i, addr := range addresses {
			log.Printf("     [%d] %s (%s)", i+1, formatAddress(addr), addr)
		}
	}

	if cfg.Verbose && len(logResult.Value.Logs) > 0 {
		log.Printf("   Logs:")
		for i, logMsg := range logResult.Value.Logs {
			log.Printf("     [%d] %s", i+1, logMsg)
		}
	}

	// Here you can add additional processing logic
	// For example, save to database, send notifications, etc.
	analyzeTransactionLogs(logResult.Value.Logs, mintAccount, mintIndex)
}

func analyzeTransactionLogs(logs []string, mintAccount string, mintIndex int) {
	// Analyze logs for specific patterns related to the mint account
	for _, logMsg := range logs {
		// Look for transfer patterns
		if contains(logMsg, "Transfer") || contains(logMsg, "transfer") {
			log.Printf("   🔄 Transfer detected in logs for mint [%d] %s", mintIndex, mintAccount)
			metrics.RecordTransaction("solana", mintAccount, "success", "transfer")
		}

		// Look for mint patterns
		if contains(logMsg, "Mint") || contains(logMsg, "mint") {
			log.Printf("   🪙 Mint operation detected in logs for mint [%d] %s", mintIndex, mintAccount)
			metrics.RecordTransaction("solana", mintAccount, "success", "mint")
		}

		// Look for burn patterns
		if contains(logMsg, "Burn") || contains(logMsg, "burn") {
			log.Printf("   �� Burn operation detected in logs for mint [%d] %s", mintIndex, mintAccount)
			metrics.RecordTransaction("solana", mintAccount, "success", "burn")
		}
	}
}

func runStatusCommand(cmd *cobra.Command, args []string) {
	// Convert WebSocket URL to HTTP RPC URL
	rpcURL := cfg.ProviderURL
	if rpcURL[:3] == "wss" {
		rpcURL = "https" + rpcURL[3:]
	} else if rpcURL[:2] == "ws" {
		rpcURL = "http" + rpcURL[2:]
	}

	log.Printf("Checking Solana network status...")
	log.Printf("RPC URL: %s", rpcURL)

	// Create RPC client
	client := rpc.New(rpcURL)

	// Get network status
	ctx := context.Background()
	health, err := client.GetHealth(ctx)
	if err != nil {
		log.Printf("❌ Failed to get network health: %v", err)
	} else {
		log.Printf("✅ Network health: %s", health)
	}

	// Get current slot
	slot, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Printf("❌ Failed to get current slot: %v", err)
	} else {
		log.Printf("📊 Current slot: %d", slot)
	}

	// Get epoch info
	epochInfo, err := client.GetEpochInfo(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Printf("❌ Failed to get epoch info: %v", err)
	} else {
		log.Printf("📅 Epoch: %d, Slot Index: %d, Slots in Epoch: %d",
			epochInfo.Epoch, epochInfo.SlotIndex, epochInfo.SlotsInEpoch)
	}

	// Check mint accounts if provided
	if cfg.MintAccount != "" {
		// Parse mint accounts
		cfg.MintAccounts = parseMintAccounts(cfg.MintAccount)

		log.Printf("\nChecking %d mint account(s):", len(cfg.MintAccounts))
		for i, mintAccount := range cfg.MintAccounts {
			log.Printf("\n[%d] Checking mint account: %s", i+1, mintAccount)

			mintPubkey, err := solana.PublicKeyFromBase58(mintAccount)
			if err != nil {
				log.Printf("❌ Invalid mint account address: %v", err)
				continue
			}

			accountInfo, err := client.GetAccountInfo(ctx, mintPubkey)
			if err != nil {
				log.Printf("❌ Failed to get account info: %v", err)
			} else if accountInfo.Value == nil {
				log.Printf("❌ Account not found")
			} else {
				log.Printf("✅ Account found")
				log.Printf("   Owner: %s", accountInfo.Value.Owner.String())
				log.Printf("   Lamports: %d", accountInfo.Value.Lamports)
				log.Printf("   Executable: %t", accountInfo.Value.Executable)
				log.Printf("   Rent Epoch: %d", accountInfo.Value.RentEpoch)
			}
		}
	} else {
		log.Println("\n💡 Use -m or --mint-account to check specific mint accounts")
	}

	// Display exporter info
	if cfg.Exporter != "" {
		log.Printf("\n📊 Prometheus metrics available at: http://%s/metrics", cfg.Exporter)
	}
}

// parseMintAccounts 解析逗号分隔的 mint 账户字符串
func parseMintAccounts(mintAccountStr string) []string {
	if mintAccountStr == "" {
		return nil
	}

	// 按逗号分割并清理空白字符
	accounts := strings.Split(mintAccountStr, ",")
	result := make([]string, 0, len(accounts))

	for _, account := range accounts {
		account = strings.TrimSpace(account)
		if account != "" {
			result = append(result, account)
		}
	}

	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TransferInfo contains parsed transfer information
type TransferInfo struct {
	From      string
	To        string
	Amount    string
	Mint      string
	FromOwner string // Owner of the source token account
	ToOwner   string // Owner of the destination token account
}

// parseTransferFromLogs extracts transfer information from transaction logs
func parseTransferFromLogs(logs []string) []TransferInfo {
	var transfers []TransferInfo

	// Common patterns for transfer logs in Solana
	patterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{
			name:    "spl_token_transfer",
			pattern: regexp.MustCompile(`Transfer (\d+) tokens from ([A-Za-z0-9]{32,44}) to ([A-Za-z0-9]{32,44})`),
		},
		{
			name:    "spl_token_transfer_with_mint",
			pattern: regexp.MustCompile(`Transfer (\d+) tokens from ([A-Za-z0-9]{32,44}) to ([A-Za-z0-9]{32,44}) for mint ([A-Za-z0-9]{32,44})`),
		},
		{
			name:    "sol_transfer",
			pattern: regexp.MustCompile(`Transfer (\d+) lamports from ([A-Za-z0-9]{32,44}) to ([A-Za-z0-9]{32,44})`),
		},
		{
			name:    "token_transfer_instruction",
			pattern: regexp.MustCompile(`Program log: Transfer: from ([A-Za-z0-9]{32,44}) to ([A-Za-z0-9]{32,44}) amount (\d+)`),
		},
		{
			name:    "spl_token_instruction",
			pattern: regexp.MustCompile(`Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke \[1\]: Transfer from ([A-Za-z0-9]{32,44}) to ([A-Za-z0-9]{32,44}) amount (\d+)`),
		},
		// New: TransferChecked instruction pattern
		{
			name:    "transfer_checked_instruction",
			pattern: regexp.MustCompile(`Program log: Instruction: TransferChecked`),
		},
	}

	for _, logMsg := range logs {
		for _, pattern := range patterns {
			matches := pattern.pattern.FindStringSubmatch(logMsg)
			if len(matches) >= 4 {
				transfer := TransferInfo{}

				switch pattern.name {
				case "spl_token_transfer":
					transfer.Amount = matches[1]
					transfer.From = matches[2]
					transfer.To = matches[3]
				case "spl_token_transfer_with_mint":
					transfer.Amount = matches[1]
					transfer.From = matches[2]
					transfer.To = matches[3]
					transfer.Mint = matches[4]
				case "sol_transfer":
					transfer.Amount = matches[1]
					transfer.From = matches[2]
					transfer.To = matches[3]
				case "token_transfer_instruction":
					transfer.From = matches[1]
					transfer.To = matches[2]
					transfer.Amount = matches[3]
				case "spl_token_instruction":
					transfer.From = matches[1]
					transfer.To = matches[2]
					transfer.Amount = matches[3]
				case "transfer_checked_instruction":
					// TransferChecked instruction detected, need to get details from RPC
					transfer.From = "detected"
					transfer.To = "detected"
					transfer.Amount = "detected"
				}

				if transfer.From != "" && transfer.To != "" {
					transfers = append(transfers, transfer)
				}
			}
		}
	}

	return transfers
}

// parseAccountAddresses extracts account addresses from transaction logs
func parseAccountAddresses(logs []string) []string {
	var addresses []string
	addressSet := make(map[string]bool)

	// Pattern to match Solana addresses (32-44 characters, base58)
	addressPattern := regexp.MustCompile(`[1-9A-HJ-NP-Za-km-z]{32,44}`)

	for _, logMsg := range logs {
		matches := addressPattern.FindAllString(logMsg, -1)
		for _, match := range matches {
			// Filter out common non-address patterns
			if !isCommonNonAddress(match) && !addressSet[match] {
				addresses = append(addresses, match)
				addressSet[match] = true
			}
		}
	}

	return addresses
}

// isCommonNonAddress checks if a string is likely not an address
func isCommonNonAddress(s string) bool {
	commonNonAddresses := []string{
		"11111111111111111111111111111111",             // System program
		"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",  // Token program
		"ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL", // Associated token program
		"11111111111111111111111111111112",             // System program (alternative)
	}

	for _, nonAddr := range commonNonAddresses {
		if s == nonAddr {
			return true
		}
	}

	// Check if it's too short or contains invalid characters
	if len(s) < 32 || len(s) > 44 {
		return true
	}

	return false
}

// formatAddress shortens an address for display
func formatAddress(address string) string {
	// if len(address) <= 12 {
	// 	return address
	// }
	// return address[:4] + "..." + address[len(address)-4:]
	if len(address) <= 6 {
		return address
	}
	return address[len(address)-6:]
}

// getTransactionDetails gets complete transfer details from logs and RPC
func getTransactionDetails(signature solana.Signature, logs []string) []TransferInfo {
	var transfers []TransferInfo

	// First try to parse from logs
	transfers = parseTransferFromLogs(logs)

	// If log parsing failed or found TransferChecked, try to get details from RPC
	if (len(transfers) == 0 || hasTransferCheckedInstruction(logs)) && rpcClient != nil {
		if cfg.Verbose {
			log.Printf("    Debug: Attempting to get transaction details from RPC...")
		}
		rpcTransfers := getTransferDetailsFromRPC(signature)
		if len(rpcTransfers) > 0 {
			transfers = rpcTransfers
			if cfg.Verbose {
				log.Printf("   🔍 Debug: Successfully got %d transfers from RPC", len(transfers))
			}
		} else {
			if cfg.Verbose {
				log.Printf("   🔍 Debug: Failed to get transfers from RPC")
			}
		}
	}

	return transfers
}

// hasTransferCheckedInstruction checks if logs contain TransferChecked instruction
func hasTransferCheckedInstruction(logs []string) bool {
	for _, logMsg := range logs {
		if strings.Contains(logMsg, "Instruction: TransferChecked") {
			return true
		}
	}
	return false
}

// getTransferDetailsFromRPC gets transfer details from RPC transaction data
func getTransferDetailsFromRPC(signature solana.Signature) []TransferInfo {
	var transfers []TransferInfo

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get complete transaction information
	tx, err := rpcClient.GetTransaction(ctx, signature, &rpc.GetTransactionOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     rpc.CommitmentConfirmed,
		MaxSupportedTransactionVersion: &[]uint64{0}[0],
	})
	if err != nil {
		if cfg.Verbose {
			log.Printf("   ⚠️  Failed to get transaction details: %v", err)
		}
		return transfers
	}

	if tx == nil || tx.Transaction == nil {
		if cfg.Verbose {
			log.Printf("   ⚠️  Transaction envelope is nil")
		}
		return transfers
	}

	// Parse transaction account information
	transaction, err := tx.Transaction.GetTransaction()
	if err != nil {
		if cfg.Verbose {
			log.Printf("   ⚠️  Failed to get transaction from envelope: %v", err)
		}
		return transfers
	}

	if transaction == nil {
		if cfg.Verbose {
			log.Printf("   ⚠️  Transaction is nil")
		}
		return transfers
	}

	// Message is a struct, not a pointer, so we don't need to check for nil
	// Get account list
	accounts := transaction.Message.AccountKeys
	if len(accounts) < 2 {
		if cfg.Verbose {
			log.Printf("   ⚠️  Not enough accounts in transaction: %d", len(accounts))
		}
		return transfers
	}

	if cfg.Verbose {
		log.Printf("    Debug: Transaction has %d accounts", len(accounts))
		for i, account := range accounts {
			log.Printf("   🔍 Debug: Account [%d]: %s", i, account.String())
		}
	}

	// Find SPL Token transfer instructions
	for i, instruction := range transaction.Message.Instructions {
		if instruction.ProgramIDIndex < uint16(len(accounts)) {
			programID := accounts[instruction.ProgramIDIndex]

			if cfg.Verbose {
				log.Printf("   🔍 Debug: Instruction [%d] program: %s", i, programID.String())
			}

			// Check if it's SPL Token program
			if programID.String() == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
				if cfg.Verbose {
					log.Printf("   🔍 Debug: Found SPL Token instruction with %d accounts", len(instruction.Accounts))
				}

				// Convert solana.CompiledInstruction to rpc.CompiledInstruction
				rpcInstruction := rpc.CompiledInstruction{
					ProgramIDIndex: instruction.ProgramIDIndex,
					Accounts:       instruction.Accounts,
					Data:           instruction.Data,
				}
				transfer := parseSPLTokenInstruction(rpcInstruction, accounts)
				if transfer.From != "" && transfer.To != "" {
					transfers = append(transfers, transfer)
					if cfg.Verbose {
						log.Printf("   🔍 Debug: Parsed transfer: %s -> %s, amount: %s",
							transfer.From, transfer.To, transfer.Amount)
					}
				}
			}
		}
	}

	return transfers
}

// parseSPLTokenInstruction parses SPL Token instruction to extract transfer info
func parseSPLTokenInstruction(instruction rpc.CompiledInstruction, accounts []solana.PublicKey) TransferInfo {
	transfer := TransferInfo{}

	// SPL Token TransferChecked instruction format:
	// 0: source (token account)
	// 1: mint (mint account)
	// 2: destination (token account)
	// 3: authority (owner)
	// 4: signers (if any)

	if len(instruction.Accounts) >= 3 {
		// Get source and destination accounts
		if int(instruction.Accounts[0]) < len(accounts) {
			transfer.From = accounts[instruction.Accounts[0]].String()
		}
		if int(instruction.Accounts[2]) < len(accounts) {
			transfer.To = accounts[instruction.Accounts[2]].String()
		}
		if len(instruction.Accounts) >= 2 && int(instruction.Accounts[1]) < len(accounts) {
			transfer.Mint = accounts[instruction.Accounts[1]].String()
		}

		// Try to parse amount from instruction data
		if len(instruction.Data) >= 9 {
			// TransferChecked instruction data format: instruction type(1 byte) + amount(8 bytes)
			amount := uint64(instruction.Data[1]) |
				uint64(instruction.Data[2])<<8 |
				uint64(instruction.Data[3])<<16 |
				uint64(instruction.Data[4])<<24 |
				uint64(instruction.Data[5])<<32 |
				uint64(instruction.Data[6])<<40 |
				uint64(instruction.Data[7])<<48 |
				uint64(instruction.Data[8])<<56
			transfer.Amount = fmt.Sprintf("%d", amount)
		}

		// Get token account owners
		if transfer.From != "" {
			transfer.FromOwner = getTokenAccountOwner(transfer.From)
		}
		if transfer.To != "" {
			transfer.ToOwner = getTokenAccountOwner(transfer.To)
		}
	}

	return transfer
}

// getTokenAccountOwner gets the owner of a token account
func getTokenAccountOwner(tokenAccount string) string {
	if rpcClient == nil {
		return "unknown"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenAccountPubkey, err := solana.PublicKeyFromBase58(tokenAccount)
	if err != nil {
		if cfg.Verbose {
			log.Printf("   ⚠️  Invalid token account address: %s", tokenAccount)
		}
		return "invalid"
	}

	// Get account info
	accountInfo, err := rpcClient.GetAccountInfo(ctx, tokenAccountPubkey)
	if err != nil {
		if cfg.Verbose {
			log.Printf("   ⚠️  Failed to get token account info for %s: %v", tokenAccount, err)
		}
		return "error"
	}

	if accountInfo.Value == nil {
		if cfg.Verbose {
			log.Printf("   ⚠️  Token account not found: %s", tokenAccount)
		}
		return "not_found"
	}

	// Check if it's a token account (owned by Token program)
	if accountInfo.Value.Owner.String() != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		if cfg.Verbose {
			log.Printf("   ⚠️  Account %s is not a token account (owner: %s)", tokenAccount, accountInfo.Value.Owner.String())
		}
		return "not_token_account"
	}

	// Parse token account data to get owner
	// Token account data format: mint(32) + owner(32) + amount(8) + delegate(32) + state(1) + is_native(1) + delegated_amount(8) + close_authority(32)
	if len(accountInfo.Value.Data.GetBinary()) < 64 {
		if cfg.Verbose {
			log.Printf("   ⚠️  Token account data too short: %d bytes", len(accountInfo.Value.Data.GetBinary()))
		}
		return "invalid_data"
	}

	// Extract owner (bytes 32-64)
	ownerBytes := accountInfo.Value.Data.GetBinary()[32:64]
	ownerPubkey := solana.PublicKeyFromBytes(ownerBytes)

	return ownerPubkey.String()
}
