package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	menuWidth     = 22
	lockAfter     = 10 * time.Minute
	defaultHeight = 24
	defaultWidth  = 100
)

type screen int

const (
	screenLogin screen = iota
	screenMain
	screenRowAction
	screenForm
	screenConfirm
	screenLocked
)

type appModel struct {
	cfgPath string
	cfg     *appConfig
	client  *apiClient

	resources     []resourceSpec
	resourceIndex int
	records       []record
	table         table.Model

	screen      screen
	previous    screen
	loginForm   inputForm
	lockForm    inputForm
	actionForm  inputForm
	actionOp    operation
	actionRec   record
	rowAction   int
	confirmText string

	width        int
	height       int
	loading      bool
	message      string
	err          string
	lastActivity time.Time
}

type loadMsg struct {
	records []record
	err     error
}

type loginMsg struct {
	cfg    *appConfig
	client *apiClient
	err    error
}

type actionMsg struct {
	message string
	err     error
}

type tickMsg time.Time

func Run(addrOverride string) error {
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}
	cfg, err := loadConfig(path)
	if err != nil {
		return err
	}
	if strings.TrimSpace(addrOverride) != "" {
		cfg.Addr = normalizeBaseURL(addrOverride)
	}
	model := newAppModel(path, cfg)
	_, err = tea.NewProgram(model).Run()
	return err
}

func newAppModel(path string, cfg *appConfig) appModel {
	resources := resources()
	tbl := newTable(resources[0])
	return appModel{
		cfgPath:      path,
		cfg:          cfg,
		resources:    resources,
		table:        tbl,
		screen:       screenLogin,
		loginForm:    newLoginForm(cfg),
		lockForm:     newLockForm(),
		width:        defaultWidth,
		height:       defaultHeight,
		lastActivity: time.Now(),
	}
}

func (m appModel) Init() tea.Cmd {
	return tick()
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.shouldLock(time.Time(msg)) {
			m.lock()
		}
		return m, tick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeTable()
		return m, nil
	case loadMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.err = ""
		m.records = msg.records
		m.table.SetRows(recordRows(msg.records))
		m.resizeTable()
		return m, nil
	case loginMsg:
		if msg.err != nil {
			m.loginForm.Err = msg.err.Error()
			return m, nil
		}
		m.cfg = msg.cfg
		m.client = msg.client
		m.screen = screenMain
		m.message = "登录成功"
		m.err = ""
		m.lastActivity = time.Now()
		m.loading = true
		return m, m.loadCurrent()
	case actionMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			m.screen = screenMain
			return m, nil
		}
		m.message = msg.message
		m.err = ""
		m.screen = screenMain
		m.loading = true
		return m, m.loadCurrent()
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.screen != screenLogin && m.screen != screenLocked {
			m.lastActivity = time.Now()
		}
	}

	switch m.screen {
	case screenLogin:
		return m.updateLogin(msg)
	case screenLocked:
		return m.updateLocked(msg)
	case screenRowAction:
		return m.updateRowAction(msg)
	case screenForm:
		return m.updateForm(msg)
	case screenConfirm:
		return m.updateConfirm(msg)
	default:
		return m.updateMain(msg)
	}
}

func (m appModel) View() tea.View {
	var content string
	switch m.screen {
	case screenLogin:
		content = m.center(m.loginForm.View(m.width))
	case screenLocked:
		content = m.center(m.lockForm.View(m.width))
	case screenRowAction:
		content = m.mainLayout(m.rowActionView())
	case screenForm:
		content = m.mainLayout(m.actionForm.CompactView(m.displayWidth()))
	case screenConfirm:
		content = m.mainLayout(m.confirmView())
	default:
		content = m.mainLayout(m.tableView())
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m appModel) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "q", "esc":
			return m, tea.Quit
		case "enter":
			if err := m.loginForm.Validate(); err != nil {
				m.loginForm.Err = err.Error()
				return m, nil
			}
			values := m.loginForm.Values()
			token := values["token"]
			lockPassword := values["lock_password"]
			if m.cfg.LockPassword != "" && lockPassword != m.cfg.LockPassword {
				m.loginForm.Err = "锁屏密码错误"
				return m, nil
			}
			cfg := *m.cfg
			cfg.Token = token
			cfg.LockPassword = lockPassword
			cfg.Addr = normalizeBaseURL(cfg.Addr)
			return m, func() tea.Msg {
				client := newAPIClient(cfg.Addr, cfg.Token)
				_, err := client.ListBot(context.Background(), botFilter{})
				if err != nil {
					return loginMsg{err: err}
				}
				if err := saveConfig(m.cfgPath, &cfg); err != nil {
					return loginMsg{err: err}
				}
				return loginMsg{cfg: &cfg, client: client}
			}
		}
	}
	var cmd tea.Cmd
	m.loginForm, cmd = m.loginForm.Update(msg)
	return m, cmd
}

func (m appModel) updateLocked(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			if err := m.lockForm.Validate(); err != nil {
				m.lockForm.Err = err.Error()
				return m, nil
			}
			if m.lockForm.Values()["lock_password"] != m.cfg.LockPassword {
				m.lockForm.Err = "锁屏密码错误"
				return m, nil
			}
			m.screen = screenMain
			m.lockForm = newLockForm()
			m.message = "已解锁"
			m.lastActivity = time.Now()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.lockForm, cmd = m.lockForm.Update(msg)
	return m, cmd
}

func (m appModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "enter":
			if err := m.actionForm.Validate(); err != nil {
				m.actionForm.Err = err.Error()
				return m, nil
			}
			spec := m.currentResource()
			op := m.actionOp
			values := m.actionForm.Values()
			rec := m.actionRec
			m.loading = true
			return m, func() tea.Msg {
				message, err := spec.Submit(context.Background(), m.client, op, values, rec)
				return actionMsg{message: message, err: err}
			}
		}
	}
	var cmd tea.Cmd
	m.actionForm, cmd = m.actionForm.Update(msg)
	return m, cmd
}

func (m appModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc", "n":
		m.screen = screenMain
		return m, nil
	case "y", "enter":
		spec := m.currentResource()
		rec := m.actionRec
		m.loading = true
		return m, func() tea.Msg {
			message, err := spec.Submit(context.Background(), m.client, operationDelete, nil, rec)
			return actionMsg{message: message, err: err}
		}
	}
	return m, nil
}

func (m appModel) updateRowAction(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	options := m.rowActionOptions()
	switch key.String() {
	case "esc", "q":
		m.screen = screenMain
		return m, nil
	case "left", "h", "up", "k":
		m.rowAction = wrapIndex(m.rowAction-1, len(options))
		return m, nil
	case "right", "l", "down", "j", "tab":
		m.rowAction = wrapIndex(m.rowAction+1, len(options))
		return m, nil
	case "u", "e":
		if hasOperation(options, operationUpdate) {
			return m.openActionForm(operationUpdate)
		}
		return m, nil
	case "d":
		if hasOperation(options, operationDelete) {
			return m.openDelete()
		}
		return m, nil
	case "enter":
		if len(options) == 0 {
			m.screen = screenMain
			return m, nil
		}
		switch options[min(m.rowAction, len(options)-1)] {
		case operationUpdate:
			return m.openActionForm(operationUpdate)
		case operationDelete:
			return m.openDelete()
		}
	}
	return m, nil
}

func (m appModel) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "q":
			return m, tea.Quit
		case "o":
			m.screen = screenLogin
			m.loginForm = newLoginForm(m.cfg)
			return m, nil
		case "l":
			m.lock()
			return m, nil
		case "r":
			m.loading = true
			return m, m.loadCurrent()
		case "left", "h":
			m = m.switchResource(-1)
			return m, m.loadCurrent()
		case "right", "tab":
			m = m.switchResource(1)
			return m, m.loadCurrent()
		case "n":
			return m.openActionForm(operationCreate)
		case "a", "e", "d", "enter":
			return m.openRowAction()
		}
		if idx := numberKeyIndex(key.String()); idx >= 0 && idx < len(m.resources) {
			m.resourceIndex = idx
			m.resetTable()
			m.loading = true
			return m, m.loadCurrent()
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m appModel) openRowAction() (tea.Model, tea.Cmd) {
	spec := m.currentResource()
	if !spec.CanUpdate && !spec.CanDelete {
		m.message = "当前模块没有行操作"
		return m, nil
	}
	rec, ok := m.selectedRecord()
	if !ok {
		m.message = "请先选择一行"
		return m, nil
	}
	m.actionRec = rec
	m.rowAction = 0
	m.screen = screenRowAction
	return m, nil
}

func (m appModel) openActionForm(op operation) (tea.Model, tea.Cmd) {
	spec := m.currentResource()
	if op == operationCreate && !spec.CanCreate {
		m.message = "当前模块不支持创建"
		return m, nil
	}
	if op == operationUpdate && !spec.CanUpdate {
		m.message = "当前模块不支持更新"
		return m, nil
	}
	rec, ok := m.selectedRecord()
	if op == operationUpdate && !ok {
		m.message = "请先选择一行"
		return m, nil
	}
	m.actionOp = op
	m.actionRec = rec
	m.actionForm = newInputForm(actionTitle(spec.Title, op), spec.Fields(op, rec))
	m.screen = screenForm
	return m, nil
}

func (m appModel) openDelete() (tea.Model, tea.Cmd) {
	spec := m.currentResource()
	if !spec.CanDelete {
		m.message = "当前模块不支持删除"
		return m, nil
	}
	rec, ok := m.selectedRecord()
	if !ok && !spec.DeleteNeedsInput {
		m.message = "请先选择一行"
		return m, nil
	}
	m.actionOp = operationDelete
	m.actionRec = rec
	if spec.DeleteNeedsInput {
		m.actionForm = newInputForm(actionTitle(spec.Title, operationDelete), spec.Fields(operationDelete, rec))
		m.screen = screenForm
		return m, nil
	}
	m.confirmText = "确认删除 " + rec.Label + " ?"
	m.screen = screenConfirm
	return m, nil
}

func (m appModel) switchResource(delta int) appModel {
	m.resourceIndex = (m.resourceIndex + delta + len(m.resources)) % len(m.resources)
	m.resetTable()
	m.loading = true
	return m
}

func (m *appModel) lock() {
	m.previous = m.screen
	m.screen = screenLocked
	m.lockForm = newLockForm()
	m.message = "已锁屏"
}

func (m appModel) shouldLock(now time.Time) bool {
	if m.screen == screenLogin || m.screen == screenLocked {
		return false
	}
	return now.Sub(m.lastActivity) >= lockAfter
}

func (m appModel) loadCurrent() tea.Cmd {
	spec := m.currentResource()
	client := m.client
	return func() tea.Msg {
		if client == nil {
			return loadMsg{err: fmt.Errorf("not logged in")}
		}
		rows, err := spec.List(context.Background(), client)
		return loadMsg{records: rows, err: err}
	}
}

func (m *appModel) resizeTable() {
	spec := m.currentResource()
	m.table.SetWidth(min(tableNaturalWidth(spec), max(20, m.displayWidth()-4)))
	m.table.SetHeight(tableVisibleHeight(len(m.records), m.displayHeight()))
}

func (m *appModel) resetTable() {
	m.table = newTable(m.currentResource())
	m.resizeTable()
	m.records = nil
}

func (m appModel) currentResource() resourceSpec {
	return m.resources[m.resourceIndex]
}

func (m appModel) selectedRecord() (record, bool) {
	if len(m.records) == 0 {
		return record{}, false
	}
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.records) {
		return record{}, false
	}
	return m.records[idx], true
}

func (m appModel) rowActionOptions() []operation {
	spec := m.currentResource()
	options := make([]operation, 0, 2)
	if spec.CanUpdate {
		options = append(options, operationUpdate)
	}
	if spec.CanDelete {
		options = append(options, operationDelete)
	}
	return options
}

func newTable(spec resourceSpec) table.Model {
	styles := table.DefaultStyles()
	styles.Header = styles.Header.Bold(true).Foreground(lipgloss.Color("#cbd5e1"))
	styles.Selected = styles.Selected.Bold(true).Foreground(lipgloss.Color("#f8fafc")).Background(lipgloss.Color("#2563eb"))
	tbl := table.New(
		table.WithColumns(spec.Columns),
		table.WithFocused(true),
		table.WithHeight(16),
		table.WithStyles(styles),
	)
	tbl.Focus()
	return tbl
}

func recordRows(records []record) []table.Row {
	rows := make([]table.Row, 0, len(records))
	for _, rec := range records {
		rows = append(rows, table.Row(rec.Cells))
	}
	return rows
}

func tableNaturalWidth(spec resourceSpec) int {
	width := 0
	for _, col := range spec.Columns {
		if col.Width > 0 {
			width += col.Width + 2
		}
	}
	return max(20, width)
}

func tableVisibleHeight(rowCount int, displayHeight int) int {
	visibleRows := max(1, rowCount)
	visibleRows = min(visibleRows, max(1, displayHeight-6))
	return visibleRows + 1
}

func newLoginForm(cfg *appConfig) inputForm {
	token := ""
	if cfg != nil {
		token = cfg.Token
	}
	return newInputForm("登录 bottui", []fieldSpec{
		{Key: "token", Label: "token", Value: token, Password: true, Required: true},
		{Key: "lock_password", Label: "锁屏密码", Password: true, Required: true},
	})
}

func newLockForm() inputForm {
	return newInputForm("已锁屏", []fieldSpec{
		{Key: "lock_password", Label: "锁屏密码", Password: true, Required: true},
	})
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func numberKeyIndex(key string) int {
	if len(key) != 1 || key[0] < '1' || key[0] > '9' {
		return -1
	}
	return int(key[0] - '1')
}

func wrapIndex(idx int, count int) int {
	if count <= 0 {
		return 0
	}
	return (idx + count) % count
}

func hasOperation(options []operation, target operation) bool {
	for _, op := range options {
		if op == target {
			return true
		}
	}
	return false
}

func actionTitle(title string, op operation) string {
	switch op {
	case operationCreate:
		return "创建 " + title
	case operationUpdate:
		return "更新 " + title
	case operationDelete:
		return "删除 " + title
	default:
		return title
	}
}
