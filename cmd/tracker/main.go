package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/solana-wallet-tracker/pkg/config"
	"github.com/yourusername/solana-wallet-tracker/pkg/monitor"
	"github.com/yourusername/solana-wallet-tracker/pkg/solana"
)

func main() {
	// Create default config file if not exists
	if err := config.CreateDefaultConfigFile(); err != nil {
		logrus.Fatalf("Failed to create default config file: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Solana client
	client, err := solana.NewClient(cfg.RPCEndpoint, cfg.WSEndpoint)
	if err != nil {
		logrus.Fatalf("Failed to initialize Solana client: %v", err)
	}
	defer client.Close()

	// Check if we have wallets to monitor
	if len(cfg.Wallets) == 0 {
		logrus.Fatal("No wallets configured to monitor. Add wallets to config.json or set MONITOR_WALLETS environment variable.")
	}

	// Initialize monitor
	walletMonitor := monitor.NewMonitor(client, cfg.Wallets, cfg.Tokens)

	// Register a handler for balance changes
	walletMonitor.RegisterHandler(func(accountInfo solana.TokenAccountInfo) {
		logrus.WithFields(logrus.Fields{
			"address":  accountInfo.Address,
			"owner":    accountInfo.Owner,
			"mint":     accountInfo.Mint,
			"balance":  accountInfo.Balance,
			"decimals": accountInfo.Decimals,
		}).Info("Token balance updated")

		// Here you can add code to notify other systems:
		// - Send message to message queue
		// - Call webhook
		// - Update database
		// - etc.
	})

	// Start the monitor
	if err := walletMonitor.Start(); err != nil {
		logrus.Fatalf("Failed to start monitor: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"wallets": cfg.Wallets,
		"tokens":  cfg.Tokens,
	}).Info("Started monitoring token balances")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown gracefully
	logrus.Info("Shutting down...")

	walletMonitor.Stop()
	logrus.Info("Solana wallet tracker stopped")
}
