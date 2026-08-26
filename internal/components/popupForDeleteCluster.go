package components

import (
	"bytes"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"

	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	buttonStyle = lipgloss.NewStyle().
			Padding(0, 3).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	activeButtonStyle = lipgloss.NewStyle().
				Padding(0, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#ff6b6b")).
				Foreground(lipgloss.Color("#ff6b6b")).
				Bold(true)

	// Стиль для подсказки
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type OverlayWindowForDeleteNewServer struct {
	selected int // 0 - левая кнопка, 1 - правая кнопка
	Name     string

	key      string
	body     string
	offset   string
	datetime string
}

func (*OverlayWindowForDeleteNewServer) Init() tea.Cmd { return nil }

func (m *OverlayWindowForDeleteNewServer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "in OverlayWindowForDeleteNewServer: event %v", msg)
	slog.Info(buf.String())

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "in OverlayWindowForDeleteNewServer: event KeyPressMsg: %s", msg.String())
		slog.Info(buf.String())

		switch msg.String() {
		case "esc":

		case "left":
			// переключиться на левую кнопку (индекс 0)
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case "right":
			// переключиться на правую кнопку (индекс 1)
			if m.selected < 1 {
				m.selected++
			}
			return m, nil

		case "enter", " ":
			// обработка нажатия выбранной кнопки
			if m.selected == 0 {
				slog.Info("Клиент хочет удалить кластер " + m.Name)
				return m, func() tea.Msg { return contracts.ClientAgreeDeleteCluster{Name: m.Name} }
			} else {
				slog.Info("Клиент отказывается удалять кластер " + m.Name)
				return m, func() tea.Msg { return contracts.ClientDisagreeDeleteCluster{} }
			}
		}
	}
	return m, nil
}

func (o *OverlayWindowForDeleteNewServer) View() tea.View {
	text := "Удалить кластер " + o.Name

	// Создаём две кнопки с учётом выделения
	var leftBtn, rightBtn string
	if o.selected == 0 {
		leftBtn = activeButtonStyle.Render("Левая")
		rightBtn = buttonStyle.Render("Правая")
	} else {
		leftBtn = buttonStyle.Render("Левая")
		rightBtn = activeButtonStyle.Render("Правая")
	}

	// Размещаем кнопки рядом
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, leftBtn, "  ", rightBtn)

	// Подсказка по управлению
	help := helpStyle.Render("←/→ для переключения, Enter/Пробел для нажатия, q/ctrl+c для выхода")

	// Объединяем в вертикальный блок
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Center, text, buttons, help))
}

type OverlayWindowForDeleteNewServerProvider struct {
}

func CreateNewOverlayWindowForDeleteNewServerProvider() *OverlayWindowForDeleteNewServerProvider {
	return &OverlayWindowForDeleteNewServerProvider{}
}

func (o *OverlayWindowForDeleteNewServerProvider) ProvideNewOverlayWindowForDeleteNewServer(clusterName string) tea.Model {
	return &OverlayWindowForDeleteNewServer{Name: clusterName}
}
