package solana

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/sirupsen/logrus"
)

// Client represents a Solana RPC and WebSocket client
type Client struct {
	RPCClient   *rpc.Client
	WSClient    *ws.Client
	RPCEndpoint string
	WSEndpoint  string
}

// TokenAccountInfo contains token account data
type TokenAccountInfo struct {
	Address       string
	Owner         string
	Mint          string
	Balance       uint64
	Decimals      uint8
	ProgramID     string
	LastUpdatedAt time.Time
}

// NewClient creates a new Solana client
func NewClient(rpcEndpoint, wsEndpoint string) (*Client, error) {
	rpcClient := rpc.New(rpcEndpoint)

	// Initialize the WebSocket client
	wsClient, err := ws.Connect(context.Background(), wsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	return &Client{
		RPCClient:   rpcClient,
		WSClient:    wsClient,
		RPCEndpoint: rpcEndpoint,
		WSEndpoint:  wsEndpoint,
	}, nil
}

// Close closes the WebSocket connection
func (c *Client) Close() {
	if c.WSClient != nil {
		c.WSClient.Close()
	}
}

// GetTokenAccounts retrieves all SPL token accounts for a given wallet address
func (c *Client) GetTokenAccounts(ctx context.Context, walletAddress string) ([]TokenAccountInfo, error) {
	// Parse the public key from string
	pubkey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet address: %w", err)
	}

	// Request token accounts
	res, err := c.RPCClient.GetTokenAccountsByOwner(
		ctx,
		pubkey,
		&rpc.GetTokenAccountsByOwnerOpts{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{
			Encoding: rpc.AccountEncodingJSONParsed,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token accounts: %w", err)
	}

	var accounts []TokenAccountInfo
	for _, item := range res.Value {
		// Parse account info
		info, ok := item.Account.Data.GetJSONParsed().(*rpc.ParsedTokenAccount)
		if !ok {
			logrus.Warnf("Failed to parse token account data for %s", item.Pubkey)
			continue
		}

		// Create token account info
		tokenInfo := TokenAccountInfo{
			Address:       item.Pubkey.String(),
			Owner:         walletAddress,
			Mint:          info.Data.Parsed.Info.Mint,
			ProgramID:     info.Data.Program,
			LastUpdatedAt: time.Now(),
		}

		// Parse balance and decimals
		tokenInfo.Balance = info.Data.Parsed.Info.TokenAmount.Amount.Uint64()
		tokenInfo.Decimals = uint8(info.Data.Parsed.Info.TokenAmount.Decimals)

		accounts = append(accounts, tokenInfo)
	}

	return accounts, nil
}

// SubscribeToTokenAccountUpdates subscribes to token account updates for a given wallet
func (c *Client) SubscribeToTokenAccountUpdates(
	ctx context.Context,
	walletAddress string,
	callback func(TokenAccountInfo),
) error {
	// Parse the wallet address
	pubkey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return fmt.Errorf("invalid wallet address: %w", err)
	}

	// Subscribe to account updates using the WebSocket client
	_, err = c.WSClient.ProgramSubscribe(
		ctx,
		solana.TokenProgramID,
		rpc.CommitmentConfirmed,
		func(res ws.ProgramNotification) {
			// The subscription callback will receive updates for all token program operations
			// We need to filter for our wallet address

			// Check if the update is for our wallet
			accountInfo, err := parseTokenAccountFromSubscription(res, walletAddress)
			if err != nil {
				logrus.Warnf("Failed to parse token account update: %v", err)
				return
			}

			// If we successfully parsed an account belonging to our wallet, call the callback
			if accountInfo != nil {
				callback(*accountInfo)
			}
		},
	)

	if err != nil {
		return fmt.Errorf("failed to subscribe to program updates: %w", err)
	}

	return nil
}

// parseTokenAccountFromSubscription parses token account info from WebSocket notification
func parseTokenAccountFromSubscription(
	notification ws.ProgramNotification,
	walletAddress string,
) (*TokenAccountInfo, error) {
	// Skip notifications that are not related to token accounts
	if notification.Result.Value.Account.Owner != solana.TokenProgramID.String() {
		return nil, nil
	}

	// Attempt to parse the account data
	var tokenAccount struct {
		Owner string `json:"owner"`
		Data  struct {
			Parsed struct {
				Type string `json:"type"`
				Info struct {
					Mint        string `json:"mint"`
					Owner       string `json:"owner"`
					TokenAmount struct {
						Amount   string `json:"amount"`
						Decimals uint8  `json:"decimals"`
					} `json:"tokenAmount"`
				} `json:"info"`
			} `json:"parsed"`
		} `json:"data"`
	}

	// Convert the account data to JSON and parse it
	accountData, err := json.Marshal(notification.Result.Value.Account)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(accountData, &tokenAccount); err != nil {
		return nil, err
	}

	// Check if this account belongs to our wallet
	if tokenAccount.Data.Parsed.Info.Owner != walletAddress {
		return nil, nil
	}

	// Create token account info
	amount, ok := new(solana.U64).SetString(tokenAccount.Data.Parsed.Info.TokenAmount.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid token amount: %s", tokenAccount.Data.Parsed.Info.TokenAmount.Amount)
	}

	return &TokenAccountInfo{
		Address:       notification.Result.Value.Pubkey.String(),
		Owner:         walletAddress,
		Mint:          tokenAccount.Data.Parsed.Info.Mint,
		Balance:       amount.Uint64(),
		Decimals:      tokenAccount.Data.Parsed.Info.TokenAmount.Decimals,
		LastUpdatedAt: time.Now(),
	}, nil
}
