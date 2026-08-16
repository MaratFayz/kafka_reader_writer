package components

import (
	"fmt"
	"marat/fayz/kafka_reader_writer/internal/windows"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type SpinnerWhenLoadMessagesFromKafkaComponent struct {
	spinner      spinner.Model
	textStyle    string
	spinnerStyle string
	helpStyle    string
}

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}

	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

func NewSpinnerWhenLoadMessagesFromKafkaComponent() *SpinnerWhenLoadMessagesFromKafkaComponent {
	m := SpinnerWhenLoadMessagesFromKafkaComponent{}

	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[5]

	return &m
}

func (s *SpinnerWhenLoadMessagesFromKafkaComponent) update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (s *SpinnerWhenLoadMessagesFromKafkaComponent) View() string {
	var viewStringOfSpinner string
	viewStringOfSpinner += fmt.Sprintf("\n %s%s%s\n\n", s.spinner.View(), textStyle("Spinning..."))
	viewStringOfSpinner += helpStyle("h/l, ←/→: change spinner • q: exit\n")
	return viewStringOfSpinner
}

func (s *SpinnerWhenLoadMessagesFromKafkaComponent) startTick() tea.Msg {
	return spinner.TickMsg{}
}
