package components

import (
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaReadMessagesTableComponent struct {
	table *table.Model
	model *windows.Model
}

var baseStyleKafkaReadMessagesTableComponent = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m *KafkaReadMessagesTableComponent) Init() tea.Cmd { return nil }

func (m *KafkaReadMessagesTableComponent) Update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd) {
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

func (m *KafkaReadMessagesTableComponent) View() string {
	return baseStyleKafkaReadMessagesTableComponent.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
	// return "asd"
}

func (m *KafkaReadMessagesTableComponent) GetActiveRow() []string {
	return m.table.SelectedRow()
}

func (m *KafkaReadMessagesTableComponent) SetValues(messages []*contracts.ReadMessagesRow) error {
	rows := make([]table.Row, 0, len(messages))

	for _, v := range messages {
		rows = append(rows, v.Row)
	}

	m.table.SetRows(rows)

	return nil
}

func CreateKafkaReadMessagesTable(m *windows.Model) *KafkaReadMessagesTableComponent {
	columns := []table.Column{
		{Title: "Offset", Width: 10},
		{Title: "Time", Width: 10},
		{Title: "Body", Width: 10},
	}

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
	t.SetRows(make([]table.Row, 0))
	k := &KafkaReadMessagesTableComponent{table: &t, model: m}
	return k
}
