package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	var addr string
	root := &cobra.Command{
		Use:           "bottui",
		Short:         "Telegram bot 后台管理 TUI",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(addr)
		},
	}
	root.Flags().StringVar(&addr, "addr", "", "API 地址，默认读取 ~/botctl/config.yaml")
	return root
}
