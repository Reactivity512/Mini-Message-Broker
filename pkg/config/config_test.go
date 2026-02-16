package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	// Загрузка с использованием несуществующего файла: следует использовать значения по умолчанию
	cfg, err := Load(filepath.Join(os.TempDir(), "nonexistent-config-12345.yaml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.GRPCPort != 50051 {
		t.Errorf("default grpc_port want 50051, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Broker.AckTimeoutSeconds != 30 {
		t.Errorf("default ack_timeout_seconds want 30, got %d", cfg.Broker.AckTimeoutSeconds)
	}
	if cfg.Broker.MaxMessageSize != 1048576 {
		t.Errorf("default max_message_size want 1048576, got %d", cfg.Broker.MaxMessageSize)
	}
}

func TestLoad_fromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`
server:
  grpc_port: 9000
  http_port: 9080
broker:
  default_retention_messages: 5000
  ack_timeout_seconds: 60
  max_message_size: 2048
logging:
  level: debug
`), 0644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.GRPCPort != 9000 {
		t.Errorf("grpc_port want 9000, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Broker.AckTimeoutSeconds != 60 {
		t.Errorf("ack_timeout want 60, got %d", cfg.Broker.AckTimeoutSeconds)
	}
	if cfg.Broker.MaxMessageSize != 2048 {
		t.Errorf("max_message_size want 2048, got %d", cfg.Broker.MaxMessageSize)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("logging.level want debug, got %s", cfg.Logging.Level)
	}
}
