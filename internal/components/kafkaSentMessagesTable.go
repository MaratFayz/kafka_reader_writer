package components

import (
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaSendMessageTableComponent struct {
	table *table.Model
	model *windows.Model
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m *KafkaSendMessageTableComponent) Init() tea.Cmd { return nil }

func (m *KafkaSendMessageTableComponent) Update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return model, tea.Quit
		case "enter":
			return model, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[2]),
			)
		}
	}
	t, cmd := m.table.Update(msg)
	m.table = &t
	return model, cmd
}

func (m *KafkaSendMessageTableComponent) View() string {
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}

func (m *KafkaSendMessageTableComponent) GetActiveRow() []string {
	return m.table.SelectedRow()
}

func (m *KafkaSendMessageTableComponent) SetValues(messages []*contracts.SentMessagesRow) error {
	rows := make([]table.Row, 0, len(messages))

	for _, v := range messages {
		rows = append(rows, v.Row)
	}

	m.table.SetRows(rows)

	return nil
}

func CreateKafkaSendMessageTable(m *windows.Model) *KafkaSendMessageTableComponent {
	columns := []table.Column{
		{Title: "Body", Width: 10},
		{Title: "Status", Width: 5},
		{Title: "Time", Width: 5},
	}

	// rows := []table.Row{
	// 	{"1", "Tokyo", "asd"},
	// 	{"2", "Delhi", "asd"},
	// 	{"3", "Shanghai", "China"},
	// 	{"4", "Dhaka", "Bangladesh"},
	// }

	t := table.New(
		table.WithColumns(columns),
		// table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
		table.WithWidth(42),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	k := &KafkaSendMessageTableComponent{table: &t, model: m}
	return k
}
