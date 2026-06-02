package windows

import (
	"marat/fayz/kafka_reader_writer/internal/components"
	"marat/fayz/kafka_reader_writer/internal/config"
	"marat/fayz/kafka_reader_writer/internal/localstorage"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/key"
	list "charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	broker              string
	topic               string
	partition           int
	textarea            textarea.Model
	senderStyle         lipgloss.Style
	messages            []string
	config              *config.Config
	kafkaServers        []kafkaServer
	selectedKafkaServer int
	activeWindow        Window
}

type kafkaServer struct {
	title string
}

func InitialModel(ls localstorage.LocalStorage) *Model {
	ta := components.CreateTextArea()

	return &Model{
		// broker:      config.broker,
		// topic:       config.topic,
		// partition:   config.partition,
		textarea: ta,
		// senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		// messages:    []string{},
		// config:      config,
	}
}

func (m Model) Init() tea.Cmd {
	//TODO add query to database
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyPressMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			mes := m.textarea.Value()
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+mes)
			m.textarea.Reset()
			// sendMessage(m.config, mes)
			return m, nil

		default:
			// Send all other keypresses to the textarea.
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
			// // The "up" and "k" keys move the cursor up
			// case "up", "k":
			// 	if m.cursor > 0 {
			// 		m.cursor--
			// 	}

			// // The "down" and "j" keys move the cursor down
			// case "down", "j":
			// 	if m.cursor < len(m.choices)-1 {
			// 		m.cursor++
			// 	}

			// // The "enter" key and the space bar toggle the selected state
			// // for the item that the cursor is pointing at.
			// case "enter", "space":
			// 	_, ok := m.selected[m.cursor]
			// 	if ok {
			// 		delete(m.selected, m.cursor)
			// 	} else {
			// 		m.selected[m.cursor] = struct{}{}
			// 	}
		}

	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() tea.View {
	// // The header
	// s := "Параметры\n\n"

	// s += fmt.Sprintf("Brokers: %s\nTopic: %s\nPartition:%d\n", m.broker, m.topic, m.partition)

	// for _, m := range m.messages {
	// 	s += m
	// 	s += "\n"
	// }

	// // The footer
	// s += "\nPress q to quit.\n"

	// v := tea.NewView(s + "\n" + m.textarea.View())
	// c := m.textarea.Cursor()
	// v.Cursor = c
	// v.AltScreen = true
	styles := components.NewStyles(false) // default to dark background styles

	delegateKeys := components.NewDelegateKeyMap()
	listKeys := components.NewListKeyMap()

	items := make([]list.Item, 0)
	items = append(items, components.NewItem("OOOO", "DDDDD"), components.NewItem("OOOO2", "DDDDD2"))

	delegate := components.NewItemDelegate(delegateKeys, &styles)
	list := list.New(items, delegate, 0, 0)
	list.Title = "YYY"
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ToggleSpinner,
			listKeys.InsertItem,
			listKeys.ToggleTitleBar,
			listKeys.ToggleStatusBar,
			listKeys.TogglePagination,
			listKeys.ToggleHelpMenu,
		}
	}

	// Update list size.
	h, v := styles.App.GetFrameSize()
	list.SetSize(100-h, 100-v)

	// Update the model and list styles.
	list.Styles.Title = styles.Title

	v2 := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, list.View(), list.View()))
	v2.AltScreen = true

	return v2
}
