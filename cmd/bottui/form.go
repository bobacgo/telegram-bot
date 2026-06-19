package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type inputForm struct {
	Title  string
	Fields []fieldSpec
	Inputs []textinput.Model
	Focus  int
	Err    string
}

func newInputForm(title string, fields []fieldSpec) inputForm {
	inputs := make([]textinput.Model, 0, len(fields))
	for i, field := range fields {
		input := textinput.New()
		input.Prompt = ""
		input.Placeholder = field.Placeholder
		input.SetValue(field.Value)
		input.SetWidth(48)
		if field.Password {
			input.EchoMode = textinput.EchoPassword
		}
		if i == 0 {
			input.Focus()
		}
		inputs = append(inputs, input)
	}
	return inputForm{
		Title:  title,
		Fields: fields,
		Inputs: inputs,
	}
}

func (f inputForm) Values() map[string]string {
	values := make(map[string]string, len(f.Fields))
	for i, field := range f.Fields {
		values[field.Key] = strings.TrimSpace(f.Inputs[i].Value())
	}
	return values
}

func (f inputForm) Validate() error {
	for i, field := range f.Fields {
		if field.Required && strings.TrimSpace(f.Inputs[i].Value()) == "" {
			return fmt.Errorf("%s is required", field.Label)
		}
	}
	return nil
}

func (f inputForm) Update(msg tea.Msg) (inputForm, tea.Cmd) {
	var cmds []tea.Cmd
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "tab", "down":
			f.moveFocus(1)
			return f, nil
		case "shift+tab", "up":
			f.moveFocus(-1)
			return f, nil
		}
	}
	for i := range f.Inputs {
		var cmd tea.Cmd
		f.Inputs[i], cmd = f.Inputs[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return f, tea.Batch(cmds...)
}

func (f inputForm) View(width int) string {
	title := modalTitleStyle.Render(f.Title)
	lines := []string{title}
	if f.Err != "" {
		lines = append(lines, errorStyle.Render(f.Err))
	}
	for i, field := range f.Fields {
		label := field.Label
		if field.Required {
			label += " *"
		}
		labelText := formLabelStyle.Render(label)
		inputText := f.Inputs[i].View()
		if i == f.Focus {
			inputText = focusedInputStyle.Render(inputText)
		} else {
			inputText = inputStyle.Render(inputText)
		}
		lines = append(lines, labelText, inputText)
	}
	lines = append(lines, hintStyle.Render("Tab/↑/↓ 切换  Enter 提交  Esc 返回"))
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if width > 0 {
		return modalBoxStyle.Width(min(width-8, 66)).Render(body)
	}
	return modalBoxStyle.Render(body)
}

func (f inputForm) CompactView(width int) string {
	panelWidth := compactFormWidth(width)
	inputWidth := max(18, panelWidth-28)
	lines := []string{
		compactPromptStyle.Render("> ") + titleStyle.Render(f.Title),
		"",
	}
	if f.Err != "" {
		lines = append(lines, errorStyle.Render(f.Err), "")
	}
	for i, field := range f.Fields {
		lines = append(lines, f.compactFieldView(i, field, inputWidth))
	}
	lines = append(lines, "", hintStyle.Render("tab/↑/↓ next  enter submit  esc back"))
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return compactFormStyle.Width(panelWidth).Render(body)
}

func (f *inputForm) moveFocus(delta int) {
	if len(f.Inputs) == 0 {
		return
	}
	f.Inputs[f.Focus].Blur()
	f.Focus = (f.Focus + delta + len(f.Inputs)) % len(f.Inputs)
	f.Inputs[f.Focus].Focus()
}

func (f inputForm) compactFieldView(index int, field fieldSpec, inputWidth int) string {
	label := field.Label
	if field.Required {
		label += " *"
	}
	input := f.Inputs[index]
	input.SetWidth(inputWidth)
	labelText := lipgloss.NewStyle().Width(18).Render(label)
	inputText := input.View()
	if index == f.Focus {
		return compactRailStyle.Render("│ ") +
			compactActiveLabelStyle.Render(labelText) +
			compactActiveInputStyle.Render(inputText)
	}
	return "  " +
		compactLabelStyle.Render(labelText) +
		compactInputStyle.Render(inputText)
}

func compactFormWidth(width int) int {
	if width <= 0 {
		return 72
	}
	return min(max(56, width-8), 92)
}
