package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config holds the application configuration
type Config struct {
	RPCEndpoint string   `json:"rpc_endpoint"`
	WSEndpoint  string   `json:"ws_endpoint"`
	Wallets     []string `json:"wallets"`
	Tokens      []string `json:"tokens"`
	LogLevel    string   `json:"log_level"`
}

// LoadConfig loads configuration from config.json and environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Default config
	config := &Config{
		RPCEndpoint: "https://api.mainnet-beta.solana.com",
		WSEndpoint:  "wss://api.mainnet-beta.solana.com",
		LogLevel:    "info",
	}

	// Check if config file exists
	configFile := "config.json"
	if _, err := os.Stat(configFile); err == nil {
		// Read config file
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, err
		}

		// Parse JSON
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	if endpoint := os.Getenv("SOLANA_RPC_ENDPOINT"); endpoint != "" {
		config.RPCEndpoint = endpoint
	}

	if endpoint := os.Getenv("SOLANA_WS_ENDPOINT"); endpoint != "" {
		config.WSEndpoint = endpoint
	}

	if wallets := os.Getenv("MONITOR_WALLETS"); wallets != "" {
		config.Wallets = strings.Split(wallets, ",")
	}

	if tokens := os.Getenv("MONITOR_TOKENS"); tokens != "" {
		config.Tokens = strings.Split(tokens, ",")
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// Setup logger
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return config, nil
}

// CreateDefaultConfigFile creates a default config.json file if it doesn't exist
func CreateDefaultConfigFile() error {
	configFile := "config.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		config := &Config{
			RPCEndpoint: "https://api.mainnet-beta.solana.com",
			WSEndpoint:  "wss://api.mainnet-beta.solana.com",
			Wallets:     []string{"ExampleWallet1", "ExampleWallet2"},
			Tokens:      []string{"ExampleTokenMint1", "ExampleTokenMint2"},
			LogLevel:    "info",
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
			return err
		}

		return ioutil.WriteFile(configFile, data, 0644)
	}
	return nil
}
