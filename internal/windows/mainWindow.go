package windows

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	leftStyle = lipgloss.NewStyle().
			Width(30).
			Height(20).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1)

	rightStyle = lipgloss.NewStyle().
			Width(50).
			Height(20).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

type MainWindow struct {
	mainModel Model
}

// Рендер левой панели со списком
func (m MainWindow) renderListOfServers() string {
	var items []string

	for i, kafkaServer := range m.mainModel.kafkaServers {
		cursor := "  "

		if m.mainModel.selectedKafkaServer == i {
			cursor = "➤ "
		}

		line := fmt.Sprintf("%s %s", cursor, kafkaServer.title)

		if m.mainModel.selectedKafkaServer == i {
			line = selectedStyle.Render(line)
		}

		items = append(items, line)
	}

	return strings.Join(items, "\n")
}

// Рендер правой панели с неизменяемыми полями
func (m MainWindow) renderInfoPanel() string {
	// // Находим выбранный элемент
	// var selectedData InfoFields
	// for _, item := range m.items {
	// 	if item.ID == m.selectedID {
	// 		selectedData = item.Data
	// 		break
	// 	}
	// }

	// // Формируем поля
	// fields := []string{
	// 	fmt.Sprintf("%s: %s", fieldStyle.Render("Название"), valueStyle.Render(selectedData.Name)),
	// 	fmt.Sprintf("%s: %s", fieldStyle.Render("Описание"), valueStyle.Render(selectedData.Description)),
	// 	fmt.Sprintf("%s: %s", fieldStyle.Render("Статус"), valueStyle.Render(selectedData.Status)),
	// 	fmt.Sprintf("%s: %s", fieldStyle.Render("Версия"), valueStyle.Render(selectedData.Version)),
	// 	fmt.Sprintf("%s: %s", fieldStyle.Render("Создан"), valueStyle.Render(selectedData.Created)),
	// }

	// // Заголовок панели
	// header := fieldStyle.Render("📋 ИНФОРМАЦИЯ О ПРОЕКТЕ")
	// separator := infoStyle.Render(strings.Repeat("─", 40))

	// return fmt.Sprintf("%s\n%s\n\n%s", header, separator, strings.Join(fields, "\n\n"))
	return ""
}

func (m MainWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	// 	m.width = msg.Width
	// 	m.height = msg.Height
	// 	return m, nil

	// case tea.KeyMsg:
	// 	switch msg.String() {
	// 	case "ctrl+c", "q", "esc":
	// 		return m, tea.Quit

	// 	case "up", "k":
	// 		if m.cursor > 0 {
	// 			m.cursor--
	// 		}

	// 	case "down", "j":
	// 		if m.cursor < len(m.items)-1 {
	// 			m.cursor++
	// 		}

	// 	case "enter", " ":
	// 		// Выбираем элемент для отображения в правой панели
	// 		m.selectedID = m.items[m.cursor].ID
	// 	}
	// }

	return m.mainModel, nil
}

func (m MainWindow) View() string {
	// Рендерим левую и правую панели
	listView := m.renderListOfServers()
	infoView := m.renderInfoPanel()

	// Применяем стили к панелям
	leftPane := leftStyle.Render(listView)
	rightPane := rightStyle.Render(infoView)

	// Объединяем панели горизонтально
	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Добавляем инструкции внизу
	help := infoStyle.Render("\n\n↑/↓ или j/k - навигация • Enter - выбрать • q/esc - выход")

	return layout + help
}
