# Solana Wallet Token Balance Tracker

A real-time monitoring system for tracking SPL token balances in Solana wallets.

## Features

- Real-time tracking of SPL token balances for Solana wallets
- Event-driven architecture using WebSocket connections
- Low resource usage and high stability
- Configurable monitoring parameters
- Automatic retry and reconnection mechanisms
- Hybrid subscription + polling approach for maximum reliability

## Implementation Details

The tracker uses a hybrid approach to ensure reliable and real-time token balance updates:

1. **WebSocket Subscriptions**: Subscribes to the Solana Token Program for real-time updates.
2. **Periodic Polling**: Performs regular polling as a fallback to ensure no updates are missed.
3. **State Management**: Maintains an in-memory state of token balances and detects changes.
4. **Event Handlers**: Provides an event-driven system to react to balance changes.

### Architecture

```
┌─────────────────┐      ┌─────────────────┐
│                 │      │                 │
│  Solana Chain   │◄────►│ RPC/WebSocket   │
│                 │      │    Endpoints    │
└─────────────────┘      └────────┬────────┘
                                  │
                                  │
                         ┌────────▼────────┐
                         │                 │
                         │  Solana Client  │
                         │                 │
                         └────────┬────────┘
                                  │
                                  │
                         ┌────────▼────────┐
                         │                 │
                         │ Wallet Monitor  │
                         │                 │
                         └────────┬────────┘
                                  │
                                  │
                         ┌────────▼────────┐
                         │                 │
                         │ Event Handlers  │
                         │                 │
                         └─────────────────┘
```

## Requirements

- Go 1.18+
- Solana RPC endpoint with WebSocket support

## Setup and Configuration

1. Clone the repository
2. Copy `config.example.json` to `config.json` and set your configuration
3. Build the application with `go build -o tracker ./cmd/tracker`
4. Run with `./tracker`

## Configuration Options

- `rpc_endpoint`: Solana RPC endpoint URL
- `ws_endpoint`: Solana WebSocket endpoint URL
- `wallets`: Array of wallet addresses to monitor
- `tokens`: Array of token mint addresses to track (leave empty to track all tokens)
- `log_level`: Logging level (debug, info, warn, error)

## Docker Support

Build and run with Docker:

```bash
# Build the Docker image
docker build -t solana-wallet-tracker .

# Run the container with environment variables
docker run -d \
  -e MONITOR_WALLETS=wallet1,wallet2 \
  -e MONITOR_TOKENS=token1,token2 \
  -e SOLANA_RPC_ENDPOINT=https://your-rpc-endpoint \
  -e SOLANA_WS_ENDPOINT=wss://your-ws-endpoint \
  solana-wallet-tracker
```

## Extending the Application

To add custom handlers for token balance changes, modify the `RegisterHandler` function call in `cmd/tracker/main.go`:

```go
walletMonitor.RegisterHandler(func(accountInfo solana.TokenAccountInfo) {
    // Your custom handling code here
    // Examples:
    // - Send to message queue
    // - Update database
    // - Call external API
    // - Generate alerts
})
```