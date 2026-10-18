package components

import (
	"bytes"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"

	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/textinput"
)

type activeField int

const (
	title activeField = iota
	url
	username
	password
	trust_store_path
	yesOrNo
)

type OverlayWindowForAppendNewCluster struct {
	title            textinput.Model
	url              textinput.Model
	username         textinput.Model
	password         textinput.Model
	trust_store_path textinput.Model
	yesOrNo          int // 0 - Yes, 1 - No
	whatFieldActive  activeField
}

func (o *OverlayWindowForAppendNewCluster) Init() tea.Cmd {
	return tea.Batch(o.title.Focus(), textinput.Blink)
}

func (m *OverlayWindowForAppendNewCluster) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "in OverlayWindowForAppendNewCluster: event %v", msg)
	slog.Info(buf.String())

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "in OverlayWindowForAppendNewCluster: event KeyPressMsg: %s", msg.String())
		slog.Info(buf.String())

		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return contracts.UserRefuseAddClusterToList{} }

		case "up":
			if m.whatFieldActive >= 0 && m.whatFieldActive <= 5 {
				m.whatFieldActive--

				switch m.whatFieldActive {
				case title:
					cmd := m.title.Focus()
					m.url.Blur()
					return m, cmd
				case url:
					m.username.Blur()
					cmd := m.url.Focus()
					return m, cmd
				case username:
					m.password.Blur()
					cmd := m.username.Focus()
					return m, cmd
				case password:
					m.trust_store_path.Blur()
					cmd := m.password.Focus()
					return m, cmd
				case trust_store_path:
					m.password.Blur()
					cmd := m.trust_store_path.Focus()
					return m, cmd
				case yesOrNo:
					return m, nil
				}
			}

			return m, nil

		case "down":
			if m.whatFieldActive < 5 {
				m.whatFieldActive++

				switch m.whatFieldActive {
				case title:
					cmd := m.title.Focus()
					return m, cmd
				case url:
					m.title.Blur()
					cmd := m.url.Focus()
					return m, cmd
				case username:
					m.url.Blur()
					cmd := m.username.Focus()
					return m, cmd
				case password:
					m.username.Blur()
					cmd := m.password.Focus()
					return m, cmd
				case trust_store_path:
					m.password.Blur()
					cmd := m.trust_store_path.Focus()
					return m, cmd
				case yesOrNo:
					m.trust_store_path.Blur()
					return m, nil
				}
			}

		case "left":
			// переключиться на левую кнопку (индекс 0)
			if m.whatFieldActive == yesOrNo && m.yesOrNo == 1 {
				m.yesOrNo--
			}
			return m, nil

		case "right":
			// переключиться на правую кнопку (индекс 1)
			if m.whatFieldActive == yesOrNo && m.yesOrNo == 0 {
				m.yesOrNo++
			}
			return m, nil

		case "enter":
			// обработка нажатия выбранной кнопки
			if m.whatFieldActive == yesOrNo && m.yesOrNo == 0 {
				slog.Info("Клиент хочет добавить кластер", "c", m)
				return m, func() tea.Msg {
					return contracts.UserConfirmAddNewClusterToList{
						Title:            m.title.View(),
						Url:              m.url.View(),
						Username:         m.username.View(),
						Password:         m.password.View(),
						Trust_store_path: m.trust_store_path.View(),
					}
				}
			} else if m.whatFieldActive == yesOrNo {
				slog.Info("Клиент отказывается добавить кластер", "c", m)
				return m, func() tea.Msg { return contracts.UserRefuseAddClusterToList{} }
			}
		default:
			switch m.whatFieldActive {
			case title:
				tm, cmd := m.title.Update(msg)
				m.title = tm
				return m, cmd
			case url:
				tm, cmd := m.url.Update(msg)
				m.url = tm
				return m, cmd
			case username:
				tm, cmd := m.username.Update(msg)
				m.username = tm
				return m, cmd
			case password:
				tm, cmd := m.password.Update(msg)
				m.password = tm
				return m, cmd
			case trust_store_path:
				tm, cmd := m.trust_store_path.Update(msg)
				m.trust_store_path = tm
				return m, cmd
			}
		}
	}
	return m, nil
}

func (o *OverlayWindowForAppendNewCluster) View() tea.View {
	s := lipgloss.NewStyle().
		Width(34).
		Height(3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center, lipgloss.Center)

	var leftBtn, rightBtn string
	if o.yesOrNo == 0 {
		leftBtn = activeButtonStyle.Render("Да")
		rightBtn = buttonStyle.Render("Нет")
	} else {
		leftBtn = buttonStyle.Render("Да")
		rightBtn = activeButtonStyle.Render("Нет")
	}

	return tea.NewView(
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.JoinHorizontal(lipgloss.Center, "title", s.Render(o.title.View())),
			lipgloss.JoinHorizontal(lipgloss.Center, "url", s.Render(o.url.View())),
			lipgloss.JoinHorizontal(lipgloss.Center, "username", s.Render(o.username.View())),
			lipgloss.JoinHorizontal(lipgloss.Center, "password", s.Render(o.password.View())),
			lipgloss.JoinHorizontal(lipgloss.Center, "trust_store_path", s.Render(o.trust_store_path.View())),
			lipgloss.JoinHorizontal(lipgloss.Center, leftBtn, "  ", rightBtn),
		),
	)
}

type OverlayWindowForAppendNewServerProvider struct {
}

func CreateNewOverlayWindowForAppendNewServerProvider() *OverlayWindowForAppendNewServerProvider {
	return &OverlayWindowForAppendNewServerProvider{}
}

func (o *OverlayWindowForAppendNewServerProvider) ProvideNew() tea.Model {
	oo := &OverlayWindowForAppendNewCluster{
		title:            textinput.New(),
		url:              textinput.New(),
		username:         textinput.New(),
		password:         textinput.New(),
		trust_store_path: textinput.New(),
	}

	oo.title.Prompt = ""

	return oo
}
