package api

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Addr           string   `json:"addr"`
	AllowedOrigins []string `json:"allowedOrigins"`
	FrontendDir    string   `json:"frontendDir"`
	TlsCertFile    string   `json:"tlsCertFile"`
	TlsKeyFile     string   `json:"tlsKeyFile"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}
