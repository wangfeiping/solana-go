#!/bin/bash

# Example usage of Solana ETL WebSocket monitor with Cobra CLI commands

echo "=== Solana ETL WebSocket Monitor Examples (Cobra Commands) ==="
echo ""

echo "Available Commands:"
echo "  start  - Start WebSocket subscription and monitor transactions"
echo "  status - Check the status of Solana network and mint account"
echo ""

echo "1. Check network status:"
echo "./solana-etl status"
echo ""

echo "2. Check network status with mint account:"
echo "./solana-etl status -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
echo ""

echo "3. Start monitoring USDC mint account:"
echo "./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -v"
echo ""

echo "4. Start monitoring with custom provider and start block:"
echo "./solana-etl start -p wss://api.mainnet-beta.solana.com -s 200000000 -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
echo ""

echo "5. Start monitoring with finalized commitment level:"
echo "./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v -c finalized -v"
echo ""

echo "6. Start monitoring USDT mint account:"
echo "./solana-etl start -m Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB -v"
echo ""

echo "7. Mixed short and long flags:"
echo "./solana-etl start -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v --commitment confirmed --verbose"
echo ""

echo "8. Show help for specific command:"
echo "./solana-etl start --help"
echo "./solana-etl status --help"
echo ""

echo "9. Show general help:"
echo "./solana-etl --help"
echo ""

echo "Parameter mapping:"
echo "  -p, --provider     : WebSocket provider URL"
echo "  -s, --start-block  : Start monitoring from block number"
echo "  -m, --mint-account : Mint account address (required for start command)"
echo "  -c, --commitment   : Commitment level"
echo "  -v, --verbose      : Enable verbose logging"
echo ""

echo "Note: Press Ctrl+C to stop monitoring when using start command"
echo ""

# Uncomment the lines below to run tests (will require network connection)
# echo "Testing status command..."
# ./solana-etl status
# echo ""
# echo "Testing status with USDC mint..."
# ./solana-etl status -m EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
