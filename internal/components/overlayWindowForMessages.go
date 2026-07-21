package components

import (
	"bytes"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/debug"

	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type OverlayWindowForMessages struct {
	key      string
	body     string
	offset   string
	datetime string
}

func (*OverlayWindowForMessages) Init() tea.Cmd { return nil }

func (m *OverlayWindowForMessages) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "in OverlayWindowForMessages: event %v", msg)
	debug.WriteDebugFile(buf.String())

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "in OverlayWindowForMessages: event KeyPressMsg: %s", msg.String())
		debug.WriteDebugFile(buf.String())

		if msg.String() == "esc" {
			return m, func() tea.Msg { return contracts.CloseOverlayWindow{} }
		}
	}
	return m, nil
}

func (o *OverlayWindowForMessages) View() tea.View {
	s := lipgloss.NewStyle().
		Width(34).
		Height(3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center, lipgloss.Center)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Center, "Key", s.Render(o.key)), lipgloss.JoinHorizontal(lipgloss.Center, "Body", s.Render(o.body)), lipgloss.JoinHorizontal(lipgloss.Center, "Offset", s.Render(o.offset)), lipgloss.JoinHorizontal(lipgloss.Center, "DateTime", s.Render(o.datetime))))
}

type OverlayWindowForMessagesProvider struct {
}

func CreateNewOverlayWindowForMessagesProvider() *OverlayWindowForMessagesProvider {
	return &OverlayWindowForMessagesProvider{}
}

func (o *OverlayWindowForMessagesProvider) ProvideNew(body string, key string, offset string, datetime string) tea.Model {
	return &OverlayWindowForMessages{key: key, body: body, offset: offset, datetime: datetime}
}
