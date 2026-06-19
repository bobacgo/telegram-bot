package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultAPIAddr = "http://127.0.0.1:8080"

type appConfig struct {
	Addr         string `yaml:"addr"`
	Token        string `yaml:"token"`
	LockPassword string `yaml:"lock_password"`
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, "botctl", "config.yaml"), nil
}

func loadConfig(path string) (*appConfig, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &appConfig{Addr: defaultAPIAddr}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := &appConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = defaultAPIAddr
	}
	cfg.Addr = normalizeBaseURL(cfg.Addr)
	cfg.Token = strings.TrimSpace(cfg.Token)
	cfg.LockPassword = strings.TrimSpace(cfg.LockPassword)
	return cfg, nil
}

func saveConfig(path string, cfg *appConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

func normalizeBaseURL(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return defaultAPIAddr
	}
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/")
}
