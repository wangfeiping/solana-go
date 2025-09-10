# Solana ETL - WebSocket Event Monitor (Cobra Commands)

This tool monitors Solana transactions related to a specific mint account using WebSocket subscriptions with a modern Cobra-based command-line interface supporting multiple commands.

## Features

- **Multiple Commands**: Modular command structure for different operations
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
| `start` | Start WebSocket subscription and monitor transactions |
| `status` | Check the status of Solana network and mint account |
| `help` | Show help information |

## Usage

### 1. Check Network Status

```bash
# Check general network status
./solana-etl status

# Check network status with specific mint account
./solana-etl status -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
```

### 2. Start Transaction Monitoring

```bash
# Basic monitoring with USDC mint
./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -v

# Monitoring with custom provider and start block
./solana-etl start -p wss://api.mainnet-beta.solana.com -s 200000000 -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v

# Monitoring with finalized commitment level
./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -c finalized -v
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
| `-m` | `--mint-account` | Mint account address to monitor | - | Yes (for start command) |
| `-c` | `--commitment` | Commitment level: "processed", "confirmed", or "finalized" | "confirmed" | No |
| `-v` | `--verbose` | Enable verbose logging | false | No |
| `-h` | `--help` | Show help information | - | No |

## Examples

### Network Status Examples

```bash
# Check network health
./solana-etl status

# Check USDC mint account
./solana-etl status -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v

# Check with custom provider
./solana-etl status -p wss://api.devnet.solana.com -m <MINT_ADDRESS>
```

### Monitoring Examples

```bash
# Monitor USDC with verbose output
./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -v

# Monitor USDT mint account
./solana-etl start -m Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB -v

# Monitor from specific block
./solana-etl start -s 200000000 -m <MINT_ADDRESS>

# Monitor with finalized commitment
./solana-etl start -c finalized -m <MINT_ADDRESS> -v
```

## Output Examples

### Status Command Output

```
2025/09/10 18:33:00 Checking Solana network status...
2025/09/10 18:33:00 RPC URL: https://api.mainnet-beta.solana.com
2025/09/10 18:33:00 ✅ Network health: ok
2025/09/10 18:33:00 📊 Current slot: 365877748
2025/09/10 18:33:00 📅 Epoch: 846, Slot Index: 405749, Slots in Epoch: 432000
2025/09/10 18:33:00 
Checking mint account: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
2025/09/10 18:33:00 ✅ Account found
2025/09/10 18:33:00    Owner: TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA
2025/09/10 18:33:00    Lamports: 418324908065
2025/09/10 18:33:00    Executable: false
2025/09/10 18:33:00    Rent Epoch: 18446744073709551615
```

### Start Command Output

```
2025/09/10 18:33:00 Connecting to Solana WebSocket at: wss://api.mainnet-beta.solana.com
2025/09/10 18:33:00 Subscribing to transactions mentioning mint account: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
2025/09/10 18:33:00 Starting to monitor transactions from block 0...
2025/09/10 18:33:00 Press Ctrl+C to stop monitoring
✅ Transaction found for mint EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v:
   Signature: 5K7x8...abc123
   Slot: 365877800
   Timestamp: 2025-09-10T18:33:15Z
   🔄 Transfer detected in logs
```

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

Press `Ctrl+C` to gracefully stop the monitoring and close WebSocket connections.

## Error Handling

The program includes comprehensive error handling for:
- Invalid mint account addresses
- WebSocket connection failures
- RPC connection failures
- Subscription errors
- Network timeouts
- Graceful shutdown
- Required parameter validation

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

# Start monitoring
./solana-etl start -m <MINT_ADDRESS>
```
