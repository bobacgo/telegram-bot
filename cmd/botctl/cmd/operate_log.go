package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"bot/repo"

	"github.com/spf13/cobra"
)

func NewOperateLogCommand(cfg *cliConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "operate-log",
		Aliases: []string{"oplog"},
		Short:   "查看操作日志",
	}
	cmd.AddCommand(newOperateLogListCommand(cfg))
	cmd.AddCommand(newOperateLogTUICommand(cfg))
	return cmd
}

func newOperateLogListCommand(cfg *cliConfig) *cobra.Command {
	filter := operateLogFilter{Page: 1, PageSize: 20}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "查看操作日志列表",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cfg.Client()
			if err != nil {
				return err
			}
			rows, err := client.ListOperateLog(cmd.Context(), filter)
			if err != nil {
				return err
			}
			printOperateLogRows(rows)
			return nil
		},
	}
	addOperateLogFilterFlags(cmd, &filter)
	return cmd
}

func newOperateLogTUICommand(cfg *cliConfig) *cobra.Command {
	filter := operateLogFilter{Page: 1, PageSize: 50}
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "打开操作日志界面",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cfg.Client()
			if err != nil {
				return err
			}
			return RunOperateLogTUI(cmd.Context(), client, filter)
		},
	}
	addOperateLogFilterFlags(cmd, &filter)
	return cmd
}

func addOperateLogFilterFlags(cmd *cobra.Command, filter *operateLogFilter) {
	cmd.Flags().StringVar(&filter.ModuleName, "module-name", "", "按模块名称过滤")
	cmd.Flags().StringVar(&filter.TargetId, "target-id", "", "按目标 ID 过滤")
	cmd.Flags().IntVar(&filter.Page, "page", filter.Page, "页码")
	cmd.Flags().IntVar(&filter.PageSize, "page-size", filter.PageSize, "每页数量")
}

func printOperateLogRows(rows []*repo.OperateLog) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tOPERATOR\tMODULE\tTARGET\tTYPE\tIP\tOPERATE_AT\tCONTENT")
	for _, row := range rows {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.Id,
			row.Operator,
			row.ModuleName,
			row.TargetId,
			opTypeText(row.OpType),
			row.IpAddress,
			formatUnix(row.OperateAt),
			truncate(row.Content, 80),
		)
	}
	_ = w.Flush()
}

func opTypeText(opType int) string {
	switch opType {
	case repo.OpAdd:
		return "create"
	case repo.OpUpdate:
		return "update"
	case repo.OpDelete:
		return "delete"
	default:
		return fmt.Sprintf("unknown(%d)", opType)
	}
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
