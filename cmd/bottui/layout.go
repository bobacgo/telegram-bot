package main

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	softBorder = lipgloss.RoundedBorder()

	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5e7eb")).
			Background(lipgloss.Color("#0f172a"))
	panelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5e7eb")).
			Background(lipgloss.Color("#0f172a")).
			Border(softBorder).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 1)
	menuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5e7eb")).
			Background(lipgloss.Color("#0f172a")).
			Width(menuWidth).
			Border(softBorder, false, true, false, false).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(1, 1)
	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cbd5e1")).
			Padding(0, 1)
	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8fafc")).
				Background(lipgloss.Color("#0f766e")).
				Bold(true).
				Padding(0, 1)
	tabBorderColor    = lipgloss.Color("#8b5cf6")
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	tabStyle          = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cbd5e1")).
				Background(lipgloss.Color("#0f172a")).
				Border(inactiveTabBorder).
				BorderForeground(tabBorderColor).
				Padding(0, 2)
	selectedTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8fafc")).
				Background(lipgloss.Color("#0f172a")).
				Border(activeTabBorder).
				BorderForeground(lipgloss.Color("#a78bfa")).
				Bold(true).
				Padding(0, 2)
	tabRuleStyle = lipgloss.NewStyle().
			Foreground(tabBorderColor).
			Background(lipgloss.Color("#0f172a"))
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7dd3fc"))
	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94a3b8"))
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#86efac"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fca5a5"))
	modalBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5e7eb")).
			Background(lipgloss.Color("#111827")).
			Border(softBorder).
			BorderForeground(lipgloss.Color("#38bdf8")).
			Padding(1, 2)
	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fde68a")).
			MarginBottom(1)
	formLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cbd5e1")).
			MarginTop(1)
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5e7eb")).
			Background(lipgloss.Color("#111827")).
			Border(softBorder).
			BorderForeground(lipgloss.Color("#475569")).
			Padding(0, 1)
	focusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8fafc")).
				Background(lipgloss.Color("#111827")).
				Border(softBorder).
				BorderForeground(lipgloss.Color("#38bdf8")).
				Padding(0, 1)
	compactFormStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e5e7eb")).
				Background(lipgloss.Color("#111827")).
				Padding(1, 3)
	compactPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8b5cf6")).
				Bold(true)
	compactRailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8b5cf6"))
	compactLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8b949e"))
	compactActiveLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8fafc")).
				Bold(true)
	compactInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cbd5e1"))
	compactActiveInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8fafc"))
	actionButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#111827")).
				Background(lipgloss.Color("#94a3b8")).
				Border(softBorder).
				BorderForeground(lipgloss.Color("#94a3b8")).
				Bold(true).
				Align(lipgloss.Center).
				Width(12).
				Padding(0, 1)
	focusedActionButtonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#111827")).
					Background(lipgloss.Color("#f472b6")).
					Border(softBorder).
					BorderForeground(lipgloss.Color("#f472b6")).
					Bold(true).
					Underline(true).
					Align(lipgloss.Center).
					Width(12).
					Padding(0, 1)
	focusedDeleteButtonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#111827")).
					Background(lipgloss.Color("#fb7185")).
					Border(softBorder).
					BorderForeground(lipgloss.Color("#fb7185")).
					Bold(true).
					Underline(true).
					Align(lipgloss.Center).
					Width(12).
					Padding(0, 1)
)

func (m appModel) mainLayout(body string) string {
	top := m.tabsView()
	bottom := m.commandBar()
	contentWidth := m.displayWidth()
	contentHeight := m.displayHeightFrom(top, bottom)
	sidebar := m.menuView(contentHeight)
	centeredBody := lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Center,
		lipgloss.Top,
		body,
		lipgloss.WithWhitespaceStyle(baseStyle),
	)
	content := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Render(centeredBody)
	main := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	page := lipgloss.JoinVertical(lipgloss.Left, top, main, bottom)
	return baseStyle.Width(m.width).Height(m.height).Render(page)
}

func (m appModel) displayWidth() int {
	return max(40, m.width-menuWidth-2)
}

func (m appModel) displayHeight() int {
	return m.displayHeightFrom(m.tabsView(), m.commandBar())
}

func (m appModel) displayHeightFrom(top string, bottom string) int {
	return max(1, m.height-lipgloss.Height(top)-lipgloss.Height(bottom))
}

func (m appModel) tabsView() string {
	tabs := make([]string, 0, len(m.resources))
	for i, resource := range m.resources {
		style := tabStyle
		if i == m.resourceIndex {
			style = selectedTabStyle
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d %s", i+1, resource.Title)))
	}
	tabLine := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	ruleWidth := max(0, m.width-lipgloss.Width(tabLine)-4)
	rule := tabRuleStyle.Render(strings.Repeat("─", ruleWidth))
	line := lipgloss.JoinHorizontal(lipgloss.Bottom, tabLine, rule)
	return baseStyle.Width(m.width).Render(lipgloss.NewStyle().PaddingLeft(2).Render(line))
}

func (m appModel) menuView(height int) string {
	lines := []string{titleStyle.Render("bottui")}
	for i, resource := range m.resources {
		style := menuItemStyle
		if i == m.resourceIndex {
			style = selectedMenuItemStyle
		}
		lines = append(lines, style.Render(fmt.Sprintf("%d. %s", i+1, resource.Title)))
	}
	lines = append(lines, "", hintStyle.Render("登录用户: token"))
	return menuStyle.Height(max(1, height)).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m appModel) tableView() string {
	spec := m.currentResource()
	lines := []string{titleStyle.Render(spec.Title)}
	if m.message != "" {
		lines = append(lines, successStyle.Render(m.message))
	}
	if m.loading {
		lines = append(lines, hintStyle.Render("加载中..."))
	}
	lines = append(lines, "", m.table.View())
	if len(m.records) == 0 && !m.loading && m.err == "" {
		lines = append(lines, hintStyle.Render("暂无数据"))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m appModel) confirmView() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		modalTitleStyle.Render("确认操作"),
		m.confirmText,
		"",
		hintStyle.Render("Enter/y 确认  n/Esc 取消"),
	)
	return modalBoxStyle.Width(min(m.displayWidth()-4, 70)).Render(body)
}

func (m appModel) rowActionView() string {
	displayWidth := m.displayWidth()
	options := m.rowActionOptions()
	buttons := make([]string, 0, len(options))
	for i, op := range options {
		buttons = append(buttons, m.rowActionButton(op, i == m.rowAction))
	}
	buttonLine := lipgloss.JoinHorizontal(lipgloss.Top, buttons...)
	buttonLine = lipgloss.PlaceHorizontal(min(displayWidth-10, 56), lipgloss.Center, buttonLine)
	lines := []string{
		modalTitleStyle.Render("行操作"),
		titleStyle.Render(m.actionRec.Label),
		"",
		buttonLine,
	}
	lines = append(lines, "", hintStyle.Render("←/→/Tab 切换  Enter 确认  Esc/q 返回"))
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return modalBoxStyle.Width(min(displayWidth-4, 64)).Render(body)
}

func (m appModel) rowActionButton(op operation, focused bool) string {
	label := "更新"
	if op == operationDelete {
		label = "删除"
	}
	style := actionButtonStyle
	if focused {
		style = focusedActionButtonStyle
		if op == operationDelete {
			style = focusedDeleteButtonStyle
		}
	}
	return style.MarginRight(2).Render(label)
}

func (m appModel) commandBar() string {
	width := max(20, m.width-2)
	if m.err != "" {
		text := ansi.Truncate("错误: "+m.err, max(1, width-4), "…")
		return panelStyle.Width(width).Render(errorStyle.Render(text))
	}

	spec := m.currentResource()
	commands := []string{"1-6 切换模块", "←/→ 切 tab", "↑/↓ 选行", "Enter/a 行操作", "r 刷新"}
	if spec.CanCreate {
		commands = append(commands, "n 新建")
	}
	if spec.CanUpdate {
		commands = append(commands, "e 操作")
	}
	if spec.CanDelete {
		commands = append(commands, "d 操作")
	}
	commands = append(commands, "l 锁屏", "o 退出登录", "q 退出")
	return panelStyle.Width(width).Render(hintStyle.Render(strings.Join(commands, "  ")))
}

func (m appModel) center(content string) string {
	placed := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
		lipgloss.WithWhitespaceStyle(baseStyle),
	)
	return baseStyle.Width(m.width).Height(m.height).Render(placed)
}

func tabBorderWithBottom(left string, middle string, right string) lipgloss.Border {
	border := softBorder
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}
