package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"bot/repo"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func NewAuthCommand(cfg *cliConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "管理 API 访问 token",
	}
	cmd.AddCommand(newAuthCreateCommand(cfg))
	cmd.AddCommand(newAuthUpdateCommand(cfg))
	cmd.AddCommand(newAuthDeleteCommand(cfg))
	cmd.AddCommand(newAuthListCommand(cfg))
	cmd.AddCommand(newAuthTUICommand(cfg))
	return cmd
}

func newAuthCreateCommand(cfg *cliConfig) *cobra.Command {
	req := &repo.Auth{Status: repo.AuthStatusUsable}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建 API 访问 token",
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Username = strings.TrimSpace(req.Username)
			req.Token = strings.TrimSpace(req.Token)
			if req.Username == "" || req.Token == "" || req.Status == 0 {
				if err := promptAuthCreate(req); err != nil {
					return err
				}
			}
			req.Username = strings.TrimSpace(req.Username)
			req.Token = strings.TrimSpace(req.Token)
			if req.Username == "" || req.Token == "" {
				return fmt.Errorf("username and token are required")
			}
			if req.Status == 0 {
				req.Status = repo.AuthStatusUsable
			}

			client, err := cfg.Client()
			if err != nil {
				return err
			}
			row, err := client.CreateAuth(cmd.Context(), req)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "创建成功: username=%s\n", row.Username)
			printAuthRows([]*repo.Auth{row})
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Username, "username", "", "用户名")
	cmd.Flags().StringVar(&req.Token, "auth-token", "", "新 token")
	cmd.Flags().IntVar(&req.Status, "status", repo.AuthStatusUsable, "状态：1 启用，2 禁用")
	return cmd
}

func newAuthUpdateCommand(cfg *cliConfig) *cobra.Command {
	req := &repo.AuthUpdateReq{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "更新 API 访问 token",
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Username = strings.TrimSpace(req.Username)
			req.Token = strings.TrimSpace(req.Token)
			if req.Username == "" || (req.Token == "" && req.Status == 0) {
				if err := promptAuthUpdate(req); err != nil {
					return err
				}
			}
			req.Username = strings.TrimSpace(req.Username)
			req.Token = strings.TrimSpace(req.Token)
			if req.Username == "" {
				return fmt.Errorf("username is required")
			}
			if req.Token == "" && req.Status == 0 {
				return fmt.Errorf("token or status is required")
			}

			client, err := cfg.Client()
			if err != nil {
				return err
			}
			row, err := client.UpdateAuth(cmd.Context(), req)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "更新成功: username=%s\n", row.Username)
			printAuthRows([]*repo.Auth{row})
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Username, "username", "", "用户名")
	cmd.Flags().StringVar(&req.Token, "auth-token", "", "新 token，不传则不更新")
	cmd.Flags().IntVar(&req.Status, "status", 0, "状态：0 不更新，1 启用，2 禁用")
	return cmd
}

func newAuthDeleteCommand(cfg *cliConfig) *cobra.Command {
	var username string
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除 API 访问 token",
		RunE: func(cmd *cobra.Command, args []string) error {
			username = strings.TrimSpace(username)
			if username == "" {
				if err := promptRequiredText("username", "要删除的用户名", &username); err != nil {
					return err
				}
			}
			username = strings.TrimSpace(username)
			if username == "" {
				return fmt.Errorf("username is required")
			}
			if !yes {
				confirmed, err := promptAuthDelete(username)
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(os.Stderr, "canceled")
					return nil
				}
			}

			client, err := cfg.Client()
			if err != nil {
				return err
			}
			if err := client.DeleteAuth(cmd.Context(), username); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "删除成功: username=%s\n", username)
			return nil
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "用户名")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认")
	return cmd
}

func newAuthListCommand(cfg *cliConfig) *cobra.Command {
	filter := authFilter{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "查看 API 访问 token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cfg.Client()
			if err != nil {
				return err
			}
			rows, err := client.ListAuth(cmd.Context(), filter)
			if err != nil {
				return err
			}
			printAuthRows(rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&filter.Username, "username", "", "按用户名过滤")
	cmd.Flags().StringVar(&filter.Token, "auth-token", "", "按 token 过滤")
	cmd.Flags().IntVar(&filter.Status, "status", 0, "按状态过滤：1 启用，2 禁用")
	return cmd
}

func newAuthTUICommand(cfg *cliConfig) *cobra.Command {
	filter := authFilter{}
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "打开 token 管理界面",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cfg.Client()
			if err != nil {
				return err
			}
			return RunAuthTUI(cmd.Context(), client, filter)
		},
	}
	cmd.Flags().StringVar(&filter.Username, "username", "", "按用户名过滤")
	cmd.Flags().StringVar(&filter.Token, "auth-token", "", "按 token 过滤")
	cmd.Flags().IntVar(&filter.Status, "status", 0, "按状态过滤：1 启用，2 禁用")
	return cmd
}

func promptAuthCreate(req *repo.Auth) error {
	if req.Status == 0 {
		req.Status = repo.AuthStatusUsable
	}
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("username").
				Value(&req.Username).
				Validate(requiredText),
			huh.NewInput().
				Title("token").
				Value(&req.Token).
				EchoMode(huh.EchoModePassword).
				Validate(requiredText),
			huh.NewSelect[int]().
				Title("status").
				Value(&req.Status).
				Options(authStatusOptions(false)...),
		),
	).Run()
}

func promptAuthUpdate(req *repo.AuthUpdateReq) error {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("username").
				Value(&req.Username).
				Validate(requiredText),
			huh.NewInput().
				Title("new token").
				Value(&req.Token).
				EchoMode(huh.EchoModePassword),
			huh.NewSelect[int]().
				Title("status").
				Value(&req.Status).
				Options(authStatusOptions(true)...),
		),
	).Run()
}

func promptAuthDelete(username string) (bool, error) {
	confirmed := false
	err := huh.NewConfirm().
		Title("delete auth token?").
		Description(username).
		Affirmative("delete").
		Negative("cancel").
		Value(&confirmed).
		Run()
	return confirmed, err
}

func promptRequiredText(title string, placeholder string, value *string) error {
	return huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(value).
		Validate(requiredText).
		Run()
}

func printAuthRows(rows []*repo.Auth) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "USERNAME\tTOKEN\tSTATUS\tCREATED_AT")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			row.Username,
			row.Token,
			authStatusText(row.Status),
			formatUnix(row.CreatedAt),
		)
	}
	_ = w.Flush()
}

func authStatusOptions(withSkip bool) []huh.Option[int] {
	options := make([]huh.Option[int], 0, 3)
	if withSkip {
		options = append(options, huh.NewOption("不更新", 0))
	}
	options = append(options,
		huh.NewOption("启用", repo.AuthStatusUsable),
		huh.NewOption("禁用", repo.AuthStatusDisabled),
	)
	return options
}

func authStatusText(status int) string {
	switch status {
	case repo.AuthStatusUsable:
		return "enabled"
	case repo.AuthStatusDisabled:
		return "disabled"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}

func requiredText(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("required")
	}
	return nil
}

func maskToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return token
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func formatUnix(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
