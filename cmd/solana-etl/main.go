package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/spf13/cobra"
)

type Config struct {
	ProviderURL  string
	StartBlock   uint64
	MintAccount  string   // 原始输入字符串
	MintAccounts []string // 解析后的 mint 账户列表
	Commitment   string
	Verbose      bool
}

var (
	cfg = &Config{}
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

		// 为每个订阅启动一个 goroutine
		wg.Add(1)
		go func(sub *ws.LogSubscription, mintAccount string, mintIndex int) {
			defer wg.Done()
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
	// Check if transaction was successful
	if logResult.Value.Err != nil {
		if cfg.Verbose {
			log.Printf("❌ Failed transaction for mint [%d] %s: %s (Error: %v)", 
				mintIndex, mintAccount, logResult.Value.Signature.String(), logResult.Value.Err)
		}
		return
	}

	// Check if we're past the start block
	if logResult.Context.Slot < cfg.StartBlock {
		return
	}

	// Log the transaction details
	log.Printf("✅ Transaction found for mint [%d] %s:", mintIndex, mintAccount)
	log.Printf("   Signature: %s", logResult.Value.Signature.String())
	log.Printf("   Slot: %d", logResult.Context.Slot)
	log.Printf("   Timestamp: %s", time.Now().Format(time.RFC3339))

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
		}
		
		// Look for mint patterns
		if contains(logMsg, "Mint") || contains(logMsg, "mint") {
			log.Printf("   🪙 Mint operation detected in logs for mint [%d] %s", mintIndex, mintAccount)
		}
		
		// Look for burn patterns
		if contains(logMsg, "Burn") || contains(logMsg, "burn") {
			log.Printf("   🔥 Burn operation detected in logs for mint [%d] %s", mintIndex, mintAccount)
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
