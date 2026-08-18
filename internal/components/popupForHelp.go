package components

import (
	"bytes"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/debug"

	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type OverlayWindowForHelp struct {
	Values []contracts.KeyValue
}

func (*OverlayWindowForHelp) Init() tea.Cmd { return nil }

func (m *OverlayWindowForHelp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "in OverlayWindowForHelp: event %v", msg)
	debug.WriteDebugFile(buf.String())

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "in OverlayWindowForHelp: event KeyPressMsg: %s", msg.String())
		debug.WriteDebugFile(buf.String())

		if msg.String() == "esc" {
			return m, func() tea.Msg { return contracts.CloseWindowForHelp{} }
		}
	}
	return m, nil
}

func (o *OverlayWindowForHelp) View() tea.View {
	s := lipgloss.NewStyle().
		Width(34).
		Height(3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center, lipgloss.Center)

	helps := make([]string, len(o.Values))

	for _, h := range o.Values {
		helps = append(helps, lipgloss.JoinHorizontal(lipgloss.Center, h.Key, s.Render(h.Value)))
	}

	return tea.NewView(
		lipgloss.JoinVertical(
			lipgloss.Center,
			// lipgloss.JoinHorizontal(lipgloss.Center, "Key", s.Render(o.key)),
			// lipgloss.JoinHorizontal(lipgloss.Center, "Body", s.Render(o.body)),
			// lipgloss.JoinHorizontal(lipgloss.Center, "Offset", s.Render(o.offset)),
			helps...))
}

type OverlayWindowForHelpProvider struct {
}

func CreateNewOverlayWindowForHelpProvider() *OverlayWindowForHelpProvider {
	return &OverlayWindowForHelpProvider{}
}

func (o *OverlayWindowForHelpProvider) ProvideNewOverlayWindowForHelp(values []contracts.KeyValue) tea.Model {
	return &OverlayWindowForHelp{Values: values}
}
