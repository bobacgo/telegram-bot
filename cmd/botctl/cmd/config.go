package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const defaultAPIAddr = "http://127.0.0.1:8080"

type fileConfig struct {
	Addr  string `yaml:"addr"`
	Token string `yaml:"token"`
}

func NewConfigCommand(cfg *cliConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理 botctl 本地配置",
	}
	cmd.AddCommand(newConfigSetCommand(cfg))
	cmd.AddCommand(newConfigShowCommand(cfg))
	cmd.AddCommand(newConfigPathCommand(cfg))
	return cmd
}

func newConfigSetCommand(cfg *cliConfig) *cobra.Command {
	req := fileConfig{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "写入本地配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			current, err := loadCLIConfig(cfg.ConfigPath)
			if err != nil {
				return err
			}
			if current != nil {
				req = *current
			}
			if flagChanged(cmd, "addr") {
				req.Addr, _ = cmd.Flags().GetString("addr")
			}
			if flagChanged(cmd, "token") {
				req.Token, _ = cmd.Flags().GetString("token")
			}
			if strings.TrimSpace(req.Addr) == "" {
				req.Addr = defaultAPIAddr
			}
			if strings.TrimSpace(req.Token) == "" {
				if err := promptConfigSet(&req); err != nil {
					return err
				}
			}
			req.Addr = normalizeBaseURL(req.Addr)
			req.Token = strings.TrimSpace(req.Token)
			if req.Token == "" {
				return fmt.Errorf("token is required")
			}
			if err := saveCLIConfig(cfg.ConfigPath, &req); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "配置保存成功: %s\n", cfg.ConfigPath)
			return nil
		},
	}
	cmd.Flags().String("addr", "", "API 地址")
	cmd.Flags().String("token", "", "API 鉴权 token")
	return cmd
}

func newConfigShowCommand(cfg *cliConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "查看本地配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			fileCfg, err := loadCLIConfig(cfg.ConfigPath)
			if err != nil {
				return err
			}
			if fileCfg == nil {
				fmt.Fprintf(os.Stdout, "config file not found: %s\n", cfg.ConfigPath)
				return nil
			}
			fmt.Fprintf(os.Stdout, "path: %s\n", cfg.ConfigPath)
			fmt.Fprintf(os.Stdout, "addr: %s\n", fileCfg.Addr)
			fmt.Fprintf(os.Stdout, "token: %s\n", maskToken(fileCfg.Token))
			return nil
		},
	}
}

func newConfigPathCommand(cfg *cliConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "输出配置文件路径",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stdout, cfg.ConfigPath)
			return nil
		},
	}
}

func defaultCLIConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, "botctl", "config.yaml"), nil
}

func loadCLIConfig(path string) (*fileConfig, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := &fileConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func saveCLIConfig(path string, cfg *fileConfig) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("config path is empty")
	}
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

func promptConfigSet(req *fileConfig) error {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("addr").
				Value(&req.Addr).
				Validate(requiredText),
			huh.NewInput().
				Title("token").
				Value(&req.Token).
				EchoMode(huh.EchoModePassword).
				Validate(requiredText),
		),
	).Run()
}

func flagChanged(cmd *cobra.Command, name string) bool {
	flag := cmd.Flags().Lookup(name)
	return flag != nil && flag.Changed
}
