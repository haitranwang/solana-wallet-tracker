package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/solana-wallet-tracker/pkg/solana"
)

// BalanceChangeHandler is a function that handles token balance changes
type BalanceChangeHandler func(accountInfo solana.TokenAccountInfo)

// Monitor handles monitoring of token balances for Solana wallets
type Monitor struct {
	client     *solana.Client
	wallets    []string
	tokens     []string
	handlers   []BalanceChangeHandler
	state      map[string]solana.TokenAccountInfo
	stateMutex sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMonitor creates a new wallet monitor
func NewMonitor(client *solana.Client, wallets, tokens []string) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Monitor{
		client:  client,
		wallets: wallets,
		tokens:  tokens,
		state:   make(map[string]solana.TokenAccountInfo),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterHandler registers a handler for balance change events
func (m *Monitor) RegisterHandler(handler BalanceChangeHandler) {
	m.handlers = append(m.handlers, handler)
}

// Start begins monitoring the wallets
func (m *Monitor) Start() error {
	// First, load the initial state
	if err := m.updateInitialState(); err != nil {
		return err
	}

	// Subscribe to updates for each wallet
	for _, wallet := range m.wallets {
		go func(walletAddress string) {
			if err := m.subscribeToWalletUpdates(walletAddress); err != nil {
				logrus.Errorf("Failed to subscribe to wallet updates for %s: %v", walletAddress, err)
			}
		}(wallet)
	}

	// Start periodic polling to ensure we don't miss any updates
	go m.startPeriodicPolling()

	return nil
}

// Stop stops the monitoring
func (m *Monitor) Stop() {
	m.cancel()
}

// GetCurrentState returns the current state of all tracked token accounts
func (m *Monitor) GetCurrentState() map[string]solana.TokenAccountInfo {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	// Create a copy of the state
	stateCopy := make(map[string]solana.TokenAccountInfo, len(m.state))
	for k, v := range m.state {
		stateCopy[k] = v
	}

	return stateCopy
}

// updateInitialState loads the initial token account state for all wallets
func (m *Monitor) updateInitialState() error {
	for _, wallet := range m.wallets {
		accounts, err := m.client.GetTokenAccounts(m.ctx, wallet)
		if err != nil {
			return err
		}

		// Filter by tokens if specified
		for _, account := range accounts {
			if m.shouldTrackToken(account.Mint) {
				m.processAccountUpdate(account)
			}
		}
	}

	return nil
}

// subscribeToWalletUpdates subscribes to token account updates for a wallet
func (m *Monitor) subscribeToWalletUpdates(walletAddress string) error {
	return m.client.SubscribeToTokenAccountUpdates(
		m.ctx,
		walletAddress,
		func(account solana.TokenAccountInfo) {
			// Check if we should track this token
			if !m.shouldTrackToken(account.Mint) {
				return
			}

			// Update the state and notify handlers if balance changed
			m.processAccountUpdate(account)
		},
	)
}

// startPeriodicPolling starts a periodic polling to update token account states
func (m *Monitor) startPeriodicPolling() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update token balances for all wallets
			for _, wallet := range m.wallets {
				accounts, err := m.client.GetTokenAccounts(m.ctx, wallet)
				if err != nil {
					logrus.Errorf("Failed to poll token accounts for %s: %v", wallet, err)
					continue
				}

				for _, account := range accounts {
					if m.shouldTrackToken(account.Mint) {
						m.processAccountUpdate(account)
					}
				}
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// processAccountUpdate processes a token account update
func (m *Monitor) processAccountUpdate(account solana.TokenAccountInfo) {
	// Lock for state update
	m.stateMutex.Lock()

	// Generate a unique key for this token account
	key := account.Owner + ":" + account.Mint

	// Check if this is a new account or if the balance has changed
	oldAccount, exists := m.state[key]
	balanceChanged := !exists || oldAccount.Balance != account.Balance

	// Update the state
	m.state[key] = account

	// Unlock after state update
	m.stateMutex.Unlock()

	// Notify handlers if balance changed
	if balanceChanged {
		logrus.WithFields(logrus.Fields{
			"wallet":  account.Owner,
			"mint":    account.Mint,
			"balance": account.Balance,
		}).Info("Token balance changed")

		// Notify all registered handlers
		for _, handler := range m.handlers {
			go handler(account)
		}
	}
}

// shouldTrackToken determines if a token should be tracked
func (m *Monitor) shouldTrackToken(mint string) bool {
	// If no tokens are specified, track all tokens
	if len(m.tokens) == 0 {
		return true
	}

	// Check if the token is in the list
	for _, token := range m.tokens {
		if token == mint {
			return true
		}
	}

	return false
}
