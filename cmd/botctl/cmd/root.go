package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type cliConfig struct {
	Addr       string
	Token      string
	ConfigPath string
	Timeout    time.Duration
	loadErr    error
}

func NewRootCommand() *cobra.Command {
	configPath, configPathErr := defaultCLIConfigPath()
	fileCfg, loadErr := loadCLIConfig(configPath)
	cfg := &cliConfig{
		Addr:       defaultAPIAddr,
		ConfigPath: configPath,
		Timeout:    10 * time.Second,
		loadErr:    loadErr,
	}
	if configPathErr != nil {
		cfg.loadErr = configPathErr
	}
	if loadErr == nil {
		cfg.applyFileConfig(fileCfg)
	}

	root := &cobra.Command{
		Use:           "botctl",
		Short:         "Telegram bot 运维工具",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&cfg.Addr, "addr", cfg.Addr, "API 地址，默认读取 ~/botctl/config.yaml")
	root.PersistentFlags().StringVar(&cfg.Token, "token", cfg.Token, "API 鉴权 token，默认读取 ~/botctl/config.yaml")
	root.PersistentFlags().DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "HTTP 请求超时时间")

	root.AddCommand(NewConfigCommand(cfg))
	root.AddCommand(NewAuthCommand(cfg))
	root.AddCommand(NewOperateLogCommand(cfg))
	return root
}

func Execute() error {
	return NewRootCommand().Execute()
}

func (cfg *cliConfig) Client() (*apiClient, error) {
	if cfg.loadErr != nil {
		return nil, cfg.loadErr
	}
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil, missingTokenError(cfg.ConfigPath)
	}
	return newAPIClient(cfg.Addr, token, cfg.Timeout), nil
}

func (cfg *cliConfig) applyFileConfig(fileCfg *fileConfig) {
	if fileCfg == nil {
		return
	}
	if strings.TrimSpace(fileCfg.Addr) != "" {
		cfg.Addr = strings.TrimSpace(fileCfg.Addr)
	}
	if strings.TrimSpace(fileCfg.Token) != "" {
		cfg.Token = strings.TrimSpace(fileCfg.Token)
	}
}
