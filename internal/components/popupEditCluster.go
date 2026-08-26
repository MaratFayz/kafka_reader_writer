package components

import (
	"bytes"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"

	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type OverlayWindowForEditCluster struct {
	key      string
	body     string
	offset   string
	datetime string
	cluster  *contracts.KafkaCluster
}

func (*OverlayWindowForEditCluster) Init() tea.Cmd { return nil }

func (m *OverlayWindowForEditCluster) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "in OverlayWindowForEditCluster: event %v", msg)
	slog.Info(buf.String())

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "in OverlayWindowForEditCluster: event KeyPressMsg: %s", msg.String())
		slog.Info(buf.String())

		if msg.String() == "esc" {
			return m, func() tea.Msg { return contracts.CloseWindowForEditCluster{} }
		}
	}
	return m, nil
}

func (o *OverlayWindowForEditCluster) View() tea.View {
	s := lipgloss.NewStyle().
		Width(34).
		Height(3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center, lipgloss.Center)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Center, "Key", s.Render(o.key)), lipgloss.JoinHorizontal(lipgloss.Center, "Body", s.Render(o.body)), lipgloss.JoinHorizontal(lipgloss.Center, "Offset", s.Render(o.offset)), lipgloss.JoinHorizontal(lipgloss.Center, "DateTime", s.Render(o.datetime))))
}

type OverlayWindowForEditClusterProvider struct {
}

func CreateNewOverlayWindowForEditClusterProvider() *OverlayWindowForEditClusterProvider {
	return &OverlayWindowForEditClusterProvider{}
}

func (o *OverlayWindowForEditClusterProvider) ProvideNewOverlayWindowForEditCluster(cluster *contracts.KafkaCluster) tea.Model {
	return &OverlayWindowForEditCluster{cluster: cluster}
}
