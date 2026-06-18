package cmd

import (
	"context"
	"fmt"
	"strings"

	"bot/repo"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type tableFetch func(context.Context) ([]table.Row, error)

type tableModel struct {
	ctx     context.Context
	title   string
	filter  string
	table   table.Model
	fetch   tableFetch
	loading bool
	err     error
}

type tableRowsMsg struct {
	rows []table.Row
	err  error
}

var (
	tuiTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7dd3fc"))
	tuiHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94a3b8"))
	tuiErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f87171"))
)

func RunAuthTUI(ctx context.Context, client *apiClient, filter authFilter) error {
	model := newTableModel(ctx, "auth tokens", describeAuthFilter(filter), authColumns(), func(ctx context.Context) ([]table.Row, error) {
		rows, err := client.ListAuth(ctx, filter)
		if err != nil {
			return nil, err
		}
		return authTableRows(rows), nil
	})
	_, err := tea.NewProgram(model).Run()
	return err
}

func RunOperateLogTUI(ctx context.Context, client *apiClient, filter operateLogFilter) error {
	model := newTableModel(ctx, "operate logs", describeOperateLogFilter(filter), operateLogColumns(), func(ctx context.Context) ([]table.Row, error) {
		rows, err := client.ListOperateLog(ctx, filter)
		if err != nil {
			return nil, err
		}
		return operateLogTableRows(rows), nil
	})
	_, err := tea.NewProgram(model).Run()
	return err
}

func newTableModel(ctx context.Context, title string, filter string, columns []table.Column, fetch tableFetch) tableModel {
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(18),
	)
	tbl.Focus()
	return tableModel{
		ctx:     ctx,
		title:   title,
		filter:  filter,
		table:   tbl,
		fetch:   fetch,
		loading: true,
	}
}

func (m tableModel) Init() tea.Cmd {
	return m.load()
}

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tableRowsMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.table.SetRows(msg.rows)
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.table.SetWidth(max(40, msg.Width-4))
		m.table.SetHeight(max(8, msg.Height-7))
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.err = nil
			return m, m.load()
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableModel) View() tea.View {
	lines := []string{
		tuiTitleStyle.Render(m.title),
		tuiHintStyle.Render("↑/↓ move  r refresh  q quit"),
	}
	if m.filter != "" {
		lines = append(lines, tuiHintStyle.Render(m.filter))
	}
	if m.loading {
		lines = append(lines, "loading...")
	}
	if m.err != nil {
		lines = append(lines, tuiErrorStyle.Render(m.err.Error()))
	}
	lines = append(lines, "", m.table.View())

	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, lines...))
	view.AltScreen = true
	return view
}

func (m tableModel) load() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.fetch(m.ctx)
		return tableRowsMsg{rows: rows, err: err}
	}
}

func authColumns() []table.Column {
	return []table.Column{
		{Title: "USERNAME", Width: 15},
		{Title: "TOKEN", Width: 48},
		{Title: "STATUS", Width: 10},
		{Title: "CREATED_AT", Width: 19},
	}
}

func authTableRows(rows []*repo.Auth) []table.Row {
	result := make([]table.Row, 0, len(rows))
	for _, row := range rows {
		result = append(result, table.Row{
			row.Username,
			row.Token,
			authStatusText(row.Status),
			formatUnix(row.CreatedAt),
		})
	}
	return result
}

func operateLogColumns() []table.Column {
	return []table.Column{
		{Title: "ID", Width: 8},
		{Title: "OPERATOR", Width: 14},
		{Title: "MODULE", Width: 14},
		{Title: "TARGET", Width: 16},
		{Title: "TYPE", Width: 8},
		{Title: "IP", Width: 15},
		{Title: "OPERATE_AT", Width: 19},
		{Title: "CONTENT", Width: 44},
	}
}

func operateLogTableRows(rows []*repo.OperateLog) []table.Row {
	result := make([]table.Row, 0, len(rows))
	for _, row := range rows {
		result = append(result, table.Row{
			fmt.Sprint(row.Id),
			row.Operator,
			row.ModuleName,
			row.TargetId,
			opTypeText(row.OpType),
			row.IpAddress,
			formatUnix(row.OperateAt),
			truncate(row.Content, 44),
		})
	}
	return result
}

func describeAuthFilter(filter authFilter) string {
	parts := make([]string, 0, 3)
	if filter.Username != "" {
		parts = append(parts, "username="+filter.Username)
	}
	if filter.Token != "" {
		parts = append(parts, "token="+maskToken(filter.Token))
	}
	if filter.Status != 0 {
		parts = append(parts, "status="+authStatusText(filter.Status))
	}
	return strings.Join(parts, "  ")
}

func describeOperateLogFilter(filter operateLogFilter) string {
	parts := make([]string, 0, 4)
	if filter.ModuleName != "" {
		parts = append(parts, "module_name="+filter.ModuleName)
	}
	if filter.TargetId != "" {
		parts = append(parts, "target_id="+filter.TargetId)
	}
	if filter.Page > 0 {
		parts = append(parts, fmt.Sprintf("page=%d", filter.Page))
	}
	if filter.PageSize > 0 {
		parts = append(parts, fmt.Sprintf("page_size=%d", filter.PageSize))
	}
	return strings.Join(parts, "  ")
}
