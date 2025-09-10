# Solana ETL - WebSocket Event Monitor (Multiple Mint Accounts)

This tool monitors Solana transactions related to multiple mint accounts using WebSocket subscriptions with a modern Cobra-based command-line interface supporting multiple commands and concurrent monitoring.

## Features

- **Multiple Mint Account Support**: Monitor multiple mint accounts simultaneously
- **Concurrent Processing**: Each mint account runs in its own goroutine for optimal performance
- **Real-time Monitoring**: WebSocket subscription to track mint account transactions
- **Network Status**: Check Solana network health and account information
- **Modern CLI**: Cobra-based interface with both short and long parameter names
- **Flexible Configuration**: Support for different commitment levels and providers
- **Verbose Logging**: Detailed transaction analysis and operation detection

## Installation

```bash
go build -o solana-etl main.go
```

## Commands

### Available Commands

| Command | Description |
|---------|-------------|
| `start` | Start WebSocket subscription and monitor transactions for multiple mint accounts |
| `status` | Check the status of Solana network and multiple mint accounts |
| `help` | Show help information |

## Usage

### 1. Check Network Status

```bash
# Check general network status
./solana-etl status

# Check multiple mint accounts status
./solana-etl status -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
```

### 2. Start Transaction Monitoring

#### Single Mint Account
```bash
# Basic monitoring with USDC mint
./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -v
```

#### Multiple Mint Accounts
```bash
# Monitor USDC and USDT simultaneously
./solana-etl start -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB" -v

# Monitor popular tokens (USDC, USDT, SOL)
./solana-etl start -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB,So11111111111111111111111111111111111111112" -v

# Monitor with custom provider and start block
./solana-etl start -p wss://api.mainnet-beta.solana.com -s 200000000 -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB" -v
```

### 3. Help and Documentation

```bash
# Show general help
./solana-etl --help

# Show help for specific command
./solana-etl start --help
./solana-etl status --help
```

## Command Line Arguments

| Short | Long | Description | Default | Required |
|-------|------|-------------|---------|----------|
| `-p` | `--provider` | Solana WebSocket provider URL | "wss://api.mainnet-beta.solana.com" | No |
| `-s` | `--start-block` | Start monitoring from this block number/slot | 0 | No |
| `-m` | `--mint-account` | Mint account address(es) to monitor (comma-separated) | - | Yes (for start command) |
| `-c` | `--commitment` | Commitment level: "processed", "confirmed", or "finalized" | "confirmed" | No |
| `-v` | `--verbose` | Enable verbose logging | false | No |
| `-h` | `--help` | Show help information | - | No |

## Examples

### Network Status Examples

```bash
# Check network health
./solana-etl status

# Check multiple mint accounts
./solana-etl status -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"

# Check with custom provider
./solana-etl status -p wss://api.devnet.solana.com -m "<MINT_ADDRESS1>,<MINT_ADDRESS2>"
```

### Monitoring Examples

#### Single Mint Account
```bash
# Monitor USDC with verbose output
./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -v

# Monitor USDT mint account
./solana-etl start -m Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB -v
```

#### Multiple Mint Accounts
```bash
# Monitor USDC and USDT simultaneously
./solana-etl start -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB" -v

# Monitor popular tokens
./solana-etl start -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB,So11111111111111111111111111111111111111112" -v

# Monitor from specific block
./solana-etl start -s 200000000 -m "<MINT_ADDRESS1>,<MINT_ADDRESS2>"

# Monitor with finalized commitment
./solana-etl start -c finalized -m "<MINT_ADDRESS1>,<MINT_ADDRESS2>" -v
```

## Popular Mint Addresses

| Token | Mint Address |
|-------|--------------|
| USDC | `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v` |
| USDT | `Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB` |
| SOL | `So11111111111111111111111111111111111111112` |
| BONK | `DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263` |

## Output Examples

### Status Command Output (Multiple Mints)

```
2025/09/11 00:02:12 Checking Solana network status...
2025/09/11 00:02:12 RPC URL: https://api.mainnet-beta.solana.com
2025/09/11 00:02:12 ✅ Network health: ok
2025/09/11 00:02:12 📊 Current slot: 365926982
2025/09/11 00:02:12 📅 Epoch: 847, Slot Index: 22982, Slots in Epoch: 432000
2025/09/11 00:02:12 
Checking 2 mint account(s):
2025/09/11 00:02:12 
[1] Checking mint account: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
2025/09/11 00:02:13 ✅ Account found
2025/09/11 00:02:13    Owner: TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA
2025/09/11 00:02:13    Lamports: 418324908117
2025/09/11 00:02:13    Executable: false
2025/09/11 00:02:13    Rent Epoch: 18446744073709551615
2025/09/11 00:02:13 
[2] Checking mint account: Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB
2025/09/11 00:02:13 ✅ Account found
2025/09/11 00:02:13    Owner: TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA
2025/09/11 00:02:13    Lamports: 141322334077
2025/09/11 00:02:13    Executable: false
2025/09/11 00:02:13    Rent Epoch: 18446744073709551615
```

### Start Command Output (Multiple Mints)

```
2025/09/11 00:02:15 Monitoring 2 mint account(s):
2025/09/11 00:02:15   [1] EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
2025/09/11 00:02:15   [2] Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB
2025/09/11 00:02:15 Connecting to Solana WebSocket at: wss://api.mainnet-beta.solana.com
2025/09/11 00:02:15 Subscribing to transactions mentioning mint account [1]: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
2025/09/11 00:02:15 Subscribing to transactions mentioning mint account [2]: Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB
2025/09/11 00:02:15 Starting to monitor transactions from block 0...
2025/09/11 00:02:15 Press Ctrl+C to stop monitoring
✅ Transaction found for mint [1] EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v:
   Signature: 5K7x8...abc123
   Slot: 365926800
   Timestamp: 2025-09-11T00:02:20Z
   🔄 Transfer detected in logs for mint [1] EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
```

## Architecture

### Concurrent Processing

The tool uses goroutines to handle multiple mint account subscriptions concurrently:

1. **Main Thread**: Manages WebSocket connection and graceful shutdown
2. **Subscription Goroutines**: Each mint account gets its own goroutine for processing
3. **WaitGroup**: Ensures all goroutines complete before program exit

### Memory Management

- Each subscription maintains its own message processing loop
- Automatic cleanup of resources on shutdown
- Efficient parsing of comma-separated mint account strings

## Adding New Commands

The modular structure makes it easy to add new commands. To add a new command:

1. Create a new command variable:
```go
var newCmd = &cobra.Command{
    Use:   "new-command",
    Short: "Description of the new command",
    Long:  `Detailed description...`,
    Run:   runNewCommand,
}
```

2. Add it to the root command in `init()`:
```go
rootCmd.AddCommand(newCmd)
```

3. Implement the command function:
```go
func runNewCommand(cmd *cobra.Command, args []string) {
    // Command implementation
}
```

## Stopping the Monitor

Press `Ctrl+C` to gracefully stop the monitoring and close WebSocket connections. All subscription goroutines will be properly terminated.

## Error Handling

The program includes comprehensive error handling for:
- Invalid mint account addresses
- WebSocket connection failures
- RPC connection failures
- Subscription errors
- Network timeouts
- Graceful shutdown
- Required parameter validation
- Multiple mint account parsing

## Dependencies

- `github.com/gagliardetto/solana-go` - Solana Go SDK
- `github.com/spf13/cobra` - Modern CLI framework

## Building

To build the executable:

```bash
go mod tidy
go build -o solana-etl main.go
```

Then run:

```bash
# Check status
./solana-etl status

# Start monitoring multiple mints
./solana-etl start -m "<MINT_ADDRESS1>,<MINT_ADDRESS2>"
```
