package config

import (
	"os"
)

type Config struct {
	Web3ProviderURL string
	EtherscanAPIKey string
	MongoURI        string
	Port            string
	WSPort          string
}

func Load() (*Config, error) {
	return &Config{
		Web3ProviderURL: os.Getenv("WEB3_PROVIDER_URL"),
		EtherscanAPIKey: os.Getenv("ETHERSCAN_API_KEY"),
		MongoURI:        os.Getenv("MONGO_URI"),
		Port:            os.Getenv("PORT"),
		WSPort:          os.Getenv("WS_PORT"),
	}, nil
}