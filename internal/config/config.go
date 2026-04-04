package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig  `mapstructure:"server"`
	Clients     ClientsConfig `mapstructure:"clients"`
	Storage     StorageConfig `mapstructure:"storage"`
	Log         LogConfig     `mapstructure:"log"`
	Environment string        `mapstructure:"environment"`
	QRTimeout   time.Duration `mapstructure:"qr_timeout"`
}

type TelegramConfig struct {
	APIID      int           `mapstructure:"api_id"`
	APIHash    string        `mapstructure:"api_hash"`
	SessionDir string        `mapstructure:"session_dir"`
	QRTimeout  time.Duration `mapstructure:"qr_timeout"`
}

type ServerConfig struct {
	Port              string        `mapstructure:"port"`
	MaxConnectionIdle time.Duration `mapstructure:"max_connection_idle"`
	MaxConnectionAge  time.Duration `mapstructure:"max_connection_age"`
}

type ClientsConfig struct {
	GRPC           GRPCClientConfig `mapstructure:"grpc"`
	TelegramClient TelegramConfig   `mapstructure:"telegram"`
}

type GRPCClientConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	Timeout      time.Duration `mapstructure:"timeout"`
	RetriesCount int           `mapstructure:"retries_count"`
}

type StorageConfig struct {
	Type          string        `mapstructure:"type"`
	SessionTTL    time.Duration `mapstructure:"session_ttl"`
	MessagesLimit int           `mapstructure:"messages_limit"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
	Format string `mapstructure:"format"`
}

// Load загружает конфигурацию
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.telegram-service")

	v.SetEnvPrefix("TELEGRAM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(cfg.Clients.TelegramClient.SessionDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session dir: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", "50051")
	v.SetDefault("server.max_connection_idle", "5m")
	v.SetDefault("server.max_connection_age", "10m")

	v.SetDefault("clients.grpc.host", "localhost")
	v.SetDefault("clients.grpc.port", "50051")
	v.SetDefault("clients.grpc.timeout", "30s")
	v.SetDefault("clients.grpc.retries_count", 3)

	v.SetDefault("clients.telegram.session_dir", "./sessions")
	v.SetDefault("clients.telegram.qr_timeout", "60s")

	v.SetDefault("storage.type", "memory")
	v.SetDefault("storage.session_ttl", "24h")
	v.SetDefault("storage.messages_limit", 1000)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.format", "json")

	v.SetDefault("environment", "development")
}

func validate(cfg *Config) error {
	if cfg.Clients.TelegramClient.APIID == 0 {
		return fmt.Errorf("telegram.api_id is required (set in config.yaml or env TELEGRAM_TELEGRAM_API_ID)")
	}

	if cfg.Clients.TelegramClient.APIHash == "" {
		return fmt.Errorf("telegram.api_hash is required (set in config.yaml or env TELEGRAM_TELEGRAM_API_HASH)")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Log.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, error)", cfg.Log.Level)
	}

	validEnv := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnv[cfg.Environment] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, production)", cfg.Environment)
	}

	return nil
}

// GetGRPCClientAddress возвращает адрес для gRPC клиента
func (c *Config) GetGRPCClientAddress() string {
	return fmt.Sprintf("%s:%s", c.Clients.GRPC.Host, c.Clients.GRPC.Port)
}
