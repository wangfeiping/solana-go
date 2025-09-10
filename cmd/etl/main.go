package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	MintAccount  string
	Commitment   string
	Verbose      bool
}

var (
	cfg = &Config{}
	rootCmd = &cobra.Command{
		Use:   "solana-etl",
		Short: "Solana WebSocket event monitor for mint account transactions",
		Long: `A real-time Solana transaction monitor that uses WebSocket subscriptions
to track transactions related to a specific mint account.

This tool connects to a Solana WebSocket provider and monitors all transactions
that mention the specified mint account, providing real-time notifications
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
	
	rootCmd.PersistentFlags().StringVarP(&cfg.MintAccount, "mint-account", "m", "", 
		"Mint account address to monitor (required)")
	
	rootCmd.PersistentFlags().StringVarP(&cfg.Commitment, "commitment", "c", "confirmed", 
		"Commitment level: processed, confirmed, or finalized")
	
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, 
		"Enable verbose logging")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start WebSocket subscription and monitor transactions",
	Long: `Start the WebSocket subscription to monitor transactions related to a specific mint account.
This command will connect to the Solana WebSocket provider and continuously parse
and analyze transactions mentioning the specified mint account.`,
	Run: runStartCommand,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of Solana network and mint account",
	Long: `Check the current status of the Solana network and validate the specified mint account.
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
	// Validate required parameters
	if cfg.MintAccount == "" {
		log.Fatal("Error: mint-account is required. Use -m or --mint-account flag.")
	}

	// Validate mint account
	mintPubkey, err := solana.PublicKeyFromBase58(cfg.MintAccount)
	if err != nil {
		log.Fatalf("Invalid mint account address: %v", err)
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

	// Subscribe to logs mentioning the mint account
	log.Printf("Subscribing to transactions mentioning mint account: %s", cfg.MintAccount)
	subscription, err := client.LogsSubscribeMentions(mintPubkey, commitment)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer subscription.Unsubscribe()

	log.Printf("Starting to monitor transactions from block %d...", cfg.StartBlock)
	log.Println("Press Ctrl+C to stop monitoring")

	// Process incoming transactions
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping monitor")
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
					log.Println("Subscription cancelled")
					return
				}
				log.Printf("Error receiving message: %v", err)
				continue
			}

			// Process the log result
			processTransaction(logResult, cfg)
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

	// Check mint account if provided
	if cfg.MintAccount != "" {
		log.Printf("\nChecking mint account: %s", cfg.MintAccount)
		
		mintPubkey, err := solana.PublicKeyFromBase58(cfg.MintAccount)
		if err != nil {
			log.Printf("❌ Invalid mint account address: %v", err)
			return
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
	} else {
		log.Println("\n💡 Use -m or --mint-account to check a specific mint account")
	}
}

func processTransaction(logResult *ws.LogResult, config *Config) {
	// Check if transaction was successful
	if logResult.Value.Err != nil {
		if config.Verbose {
			log.Printf("❌ Failed transaction: %s (Error: %v)", 
				logResult.Value.Signature.String(), logResult.Value.Err)
		}
		return
	}

	// Check if we're past the start block
	if logResult.Context.Slot < config.StartBlock {
		return
	}

	// Log the transaction details
	log.Printf("✅ Transaction found for mint %s:", config.MintAccount)
	log.Printf("   Signature: %s", logResult.Value.Signature.String())
	log.Printf("   Slot: %d", logResult.Context.Slot)
	log.Printf("   Timestamp: %s", time.Now().Format(time.RFC3339))

	if config.Verbose && len(logResult.Value.Logs) > 0 {
		log.Printf("   Logs:")
		for i, logMsg := range logResult.Value.Logs {
			log.Printf("     [%d] %s", i+1, logMsg)
		}
	}

	// Here you can add additional processing logic
	// For example, save to database, send notifications, etc.
	analyzeTransactionLogs(logResult.Value.Logs, config)
}

func analyzeTransactionLogs(logs []string, config *Config) {
	// Analyze logs for specific patterns related to the mint account
	for _, logMsg := range logs {
		// Look for transfer patterns
		if contains(logMsg, "Transfer") || contains(logMsg, "transfer") {
			log.Printf("   🔄 Transfer detected in logs")
		}
		
		// Look for mint patterns
		if contains(logMsg, "Mint") || contains(logMsg, "mint") {
			log.Printf("   🪙 Mint operation detected in logs")
		}
		
		// Look for burn patterns
		if contains(logMsg, "Burn") || contains(logMsg, "burn") {
			log.Printf("   🔥 Burn operation detected in logs")
		}
	}
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
