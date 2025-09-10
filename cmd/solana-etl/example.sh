#!/bin/bash

# Example usage of Solana ETL WebSocket monitor with multiple mint accounts

echo "=== Solana ETL WebSocket Monitor Examples (Multiple Mint Accounts) ==="
echo ""

echo "Available Commands:"
echo "  start  - Start WebSocket subscription and monitor transactions"
echo "  status - Check the status of Solana network and mint accounts"
echo ""

echo "1. Check network status:"
echo "./solana-etl status"
echo ""

echo "2. Check multiple mint accounts status:"
echo "./solana-etl status -m \"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB\""
echo ""

echo "3. Start monitoring single mint account:"
echo "./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -v"
echo ""

echo "4. Start monitoring multiple mint accounts (USDC + USDT):"
echo "./solana-etl start -m \"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB\" -v"
echo ""

echo "5. Start monitoring multiple mint accounts with custom provider:"
echo "./solana-etl start -p wss://api.mainnet-beta.solana.com -m \"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB\" -v"
echo ""

echo "6. Start monitoring multiple mint accounts with start block:"
echo "./solana-etl start -s 200000000 -m \"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB\" -v"
echo ""

echo "7. Start monitoring multiple mint accounts with finalized commitment:"
echo "./solana-etl start -c finalized -m \"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB\" -v"
echo ""

echo "8. Monitor popular tokens (USDC, USDT, SOL):"
echo "./solana-etl start -m \"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB,So11111111111111111111111111111111111111112\" -v"
echo ""

echo "9. Show help for specific command:"
echo "./solana-etl start --help"
echo "./solana-etl status --help"
echo ""

echo "10. Show general help:"
echo "./solana-etl --help"
echo ""

echo "Parameter mapping:"
echo "  -p, --provider     : WebSocket provider URL"
echo "  -s, --start-block  : Start monitoring from block number"
echo "  -m, --mint-account : Mint account address(es) (comma-separated for multiple)"
echo "  -c, --commitment   : Commitment level"
echo "  -v, --verbose      : Enable verbose logging"
echo ""

echo "Multiple Mint Account Examples:"
echo "  USDC: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
echo "  USDT: Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
echo "  SOL:  So11111111111111111111111111111111111111112"
echo "  BONK: DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263"
echo ""

echo "Note: Press Ctrl+C to stop monitoring when using start command"
echo ""

# Uncomment the lines below to run tests (will require network connection)
# echo "Testing status with multiple mints..."
# ./solana-etl status -m "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v,Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
