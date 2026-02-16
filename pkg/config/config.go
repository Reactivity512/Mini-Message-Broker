package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	Broker  BrokerConfig
	Logging LoggingConfig
}

type ServerConfig struct {
	GRPCPort int
	HTTPPort int
}

type BrokerConfig struct {
	DefaultRetentionMessages int
	AckTimeoutSeconds        int
	MaxMessageSize           int
}

type LoggingConfig struct {
	Level string
}

func Load(path string) (*Config, error) {
	v := viper.New()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok || errors.Is(err, os.ErrNotExist) {
			v.SetDefault("server.grpc_port", 50051)
			v.SetDefault("server.http_port", 8080)
			v.SetDefault("broker.default_retention_messages", 10000)
			v.SetDefault("broker.ack_timeout_seconds", 30)
			v.SetDefault("broker.max_message_size", 1048576)
			v.SetDefault("logging.level", "info")
		} else {
			return nil, fmt.Errorf("config: %w", err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			GRPCPort: v.GetInt("server.grpc_port"),
			HTTPPort: v.GetInt("server.http_port"),
		},
		Broker: BrokerConfig{
			DefaultRetentionMessages: v.GetInt("broker.default_retention_messages"),
			AckTimeoutSeconds:        v.GetInt("broker.ack_timeout_seconds"),
			MaxMessageSize:           v.GetInt("broker.max_message_size"),
		},
		Logging: LoggingConfig{
			Level: v.GetString("logging.level"),
		},
	}

	return cfg, nil
}
