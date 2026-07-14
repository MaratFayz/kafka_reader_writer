package components

import (
	"log"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"strings"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaReadWriteTabsstyles struct {
	doc         lipgloss.Style
	highlight   lipgloss.Style
	inactiveTab lipgloss.Style
	activeTab   lipgloss.Style
	window      lipgloss.Style
}

func newStyles(bgIsDark bool) *KafkaReadWriteTabsstyles {
	lightDark := lipgloss.LightDark(bgIsDark)

	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")
	highlightColor := lightDark(lipgloss.Color("#874BFD"), lipgloss.Color("#7D56F4"))

	s := new(KafkaReadWriteTabsstyles)
	s.doc = lipgloss.NewStyle().
		Padding(1, 2, 1, 2)
	s.inactiveTab = lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(highlightColor).
		Padding(0, 1)
	s.activeTab = s.inactiveTab.
		Border(activeTabBorder, true)
	s.window = lipgloss.NewStyle().
		BorderForeground(highlightColor).
		Padding(2, 0).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop()
	return s
}

type activeTab int

const (
	READ activeTab = iota
	WRITE
)

type activeComponentWriteTab int

const (
	TABS activeComponentWriteTab = iota
	TEXT_AREA
	TABLE
)

type KafkaReadWriteTabsComponent struct {
	styles                            *KafkaReadWriteTabsstyles
	Tabs                              []string
	TabContent                        []string
	activeTab                         activeTab
	activeComponentWriteTab           activeComponentWriteTab
	model                             *windows.Model
	kafkaSendMessageTextAreaComponent KafkaSendMessageTextArea
	kafkaSendMessageTableComponent    KafkaSendMessageTable
	kafkaProducerConsumer             kafkaConsumerProducer
	historySentMessages               historySentMessages
	kafkaReadMessageTable             KafkaReadMessageTable
	isLoadHistorySentMessages         atomic.Bool
}

type kafkaConsumerProducer interface {
	Send(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, kafkaPartition string, textToSend string) error
	Read(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, kafkaPartition string, numberOfReadMessages int) ([]*contracts.ReadMessagesRow, error)
}

type historySentMessages interface {
	SaveMessage(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, messageToSave string) error
	GetMessages(cluster *contracts.KafkaCluster, topic string) ([]*contracts.SentMessagesRow, error)
}

type KafkaSendMessageTextArea interface {
	Update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd)
	View() string
	Init() tea.Cmd
	SetText(text string)
	GetText() string
}

type KafkaSendMessageTable interface {
	Update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd)
	View() string
	Init() tea.Cmd
	GetActiveRow() []string
	SetValues(messages []*contracts.SentMessagesRow) error
}

type KafkaReadMessageTable interface {
	Update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd)
	View() string
	Init() tea.Cmd
	GetActiveRow() []string
	SetValues(messages []*contracts.ReadMessagesRow) error
}

type errorWhenSendMessageToKafka struct {
	error error
}

type successWhenSendMessageToKafka struct {
}

type errorWhenSaveMessageToHistory struct {
	error error
}

type successWhenSaveMessageToHistory struct {
}

type errorWhenGetMessagesFromHistory struct {
	error error
}

type successWhenGetMessagesFromHistory struct {
	messages []*contracts.SentMessagesRow
}

type errorWhenSetMessagesHistoryMessagesTableCmd struct {
	error error
}

type successWhenSetMessagesHistoryMessagesTableCmd struct {
}

type successWhenReadMessagesFromKafka struct {
	messages []*contracts.ReadMessagesRow
}

type errorWhenReadMessagesFromKafka struct {
	error error
}

type successWhenSetMessagesToReadMessagesTableCmd struct {
}

type errorWhenSetMessagesToReadMessagesTableCmd struct {
	error error
}

func (k *KafkaReadWriteTabsComponent) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	getMessagesFromHistoryCmd := func() tea.Msg {
		cluster := m.GetClusterByTitle(m.SelectedKafkaCluster)
		topic := m.SelectedKafkaTopic

		messages, err := k.historySentMessages.GetMessages(cluster, topic)

		if err != nil {
			return errorWhenGetMessagesFromHistory{error: err}
		}

		return successWhenGetMessagesFromHistory{messages: messages}
	}

	//добавить проверку, какая вкладка видима и направлять туда события
	cmds := make([]tea.Cmd, 0)

	if m.ActivePane == 3 {
		if k.isLoadHistorySentMessages.Load() == false {
			cmds = append(cmds, getMessagesFromHistoryCmd)
			k.isLoadHistorySentMessages.Store(true)
		}

		switch msg := msg.(type) {
		case successWhenSaveMessageToHistory:
			cmds = append(cmds, getMessagesFromHistoryCmd)
			return m, tea.Batch(cmds...)

		case successWhenGetMessagesFromHistory:
			setMessagesHistoryMessagesTableCmd := func() tea.Msg {
				err := k.kafkaSendMessageTableComponent.SetValues(msg.messages)

				if err != nil {
					return errorWhenSetMessagesHistoryMessagesTableCmd{error: err}
				}

				return successWhenSetMessagesHistoryMessagesTableCmd{}
			}
			cmds = append(cmds, setMessagesHistoryMessagesTableCmd)

			return m, tea.Batch(cmds...)

		case successWhenReadMessagesFromKafka:
			setMessagesToReadMessagesTableCmd := func() tea.Msg {
				err := k.kafkaReadMessageTable.SetValues(msg.messages)

				if err != nil {
					return errorWhenSetMessagesToReadMessagesTableCmd{error: err}
				}

				return successWhenSetMessagesToReadMessagesTableCmd{}
			}
			cmds = append(cmds, setMessagesToReadMessagesTableCmd)

			return m, tea.Batch(cmds...)

		case tea.KeyPressMsg:
			switch k.activeComponentWriteTab {
			case TABS:
				switch keypress := msg.String(); keypress {
				case "ctrl+c", "q":
					cmds = append(cmds, tea.Quit)
					return m, tea.Batch(cmds...)
				case "right", "l", "n", "tab":
					k.activeTab = min(k.activeTab+1, activeTab(len(k.Tabs)-1))
					return m, tea.Batch(cmds...)
				case "left", "h", "p", "shift+tab":
					k.activeTab = max(k.activeTab-1, 0)
					return m, tea.Batch(cmds...)
				case "down":
					k.activeComponentWriteTab = min(k.activeComponentWriteTab+1, 3)
					return m, tea.Batch(cmds...)
				}
			}

			switch k.activeTab {
			case WRITE:
				switch k.activeComponentWriteTab {
				case TEXT_AREA:
					switch keypress := msg.String(); keypress {
					case "down":
						k.activeComponentWriteTab = min(k.activeComponentWriteTab+1, 3)
						return m, tea.Batch(cmds...)
					case "ctrl+s":
						textToSend := k.kafkaSendMessageTextAreaComponent.GetText()
						// k.kafkaSendMessageTextAreaComponent.SetText("")
						cluster := m.GetClusterByTitle(m.SelectedKafkaCluster)
						topic := m.SelectedKafkaTopic
						partition := m.SelectedKafkaPartition

						sendMessageCmd := func() tea.Msg {
							err := k.kafkaProducerConsumer.Send(cluster, topic, partition, textToSend)

							if err != nil {
								return errorWhenSendMessageToKafka{error: err}
							}

							return successWhenSendMessageToKafka{}
						}

						saveMessageToHistoryCmd := func() tea.Msg {
							err := k.historySentMessages.SaveMessage(cluster, topic, textToSend)

							if err != nil {
								return errorWhenSaveMessageToHistory{error: err}
							}

							return successWhenSaveMessageToHistory{}
						}

						cmds = append(cmds, sendMessageCmd, saveMessageToHistoryCmd)
						return m, tea.Batch(cmds...)
					default:
						var ok bool
						m2, cmd := k.kafkaSendMessageTextAreaComponent.Update(msg, m)

						m, ok = m2.(*windows.Model)
						if !ok {
							// Handle the case where m2 is not *windows.Model
							log.Fatal("m2 is not a *windows.Model")
						}

						cmds = append(cmds, cmd)
						return m, tea.Batch(cmds...)
					}
				case TABLE:
					switch keypress := msg.String(); keypress {
					case "enter":
						columns := k.kafkaSendMessageTableComponent.GetActiveRow()
						if len(columns) > 0 {
							k.kafkaSendMessageTextAreaComponent.SetText(k.kafkaSendMessageTableComponent.GetActiveRow()[0])
						}
						k.activeComponentWriteTab = TEXT_AREA
						return m, tea.Batch(cmds...)
					default:
						var ok bool
						m2, cmd := k.kafkaSendMessageTableComponent.Update(msg, m)

						m, ok = m2.(*windows.Model)
						if !ok {
							// Handle the case where m2 is not *windows.Model
							log.Fatal("m2 is not a *windows.Model")
						}

						cmds = append(cmds, cmd)
						return m, tea.Batch(cmds...)
					}
				}
			case READ:
				switch keypress := msg.String(); keypress {
				case "r":
					cluster := m.GetClusterByTitle(m.SelectedKafkaCluster)
					topic := m.SelectedKafkaTopic
					partition := m.SelectedKafkaPartition

					readMessagesFromKafkaCmd := func() tea.Msg {
						numberOfReadMessages := 10
						messages, err := k.kafkaProducerConsumer.Read(cluster, topic, partition, numberOfReadMessages)

						if err != nil {
							return errorWhenReadMessagesFromKafka{error: err}
						}

						return successWhenReadMessagesFromKafka{messages: messages}
					}

					cmds = append(cmds, readMessagesFromKafkaCmd)
					return m, tea.Batch(cmds...)
				default:
					// var ok bool
					// m2, cmd := k.kafkaSendMessageTableComponent.Update(msg, m)

					// m, ok = m2.(*windows.Model)
					// if !ok {
					// 	// Handle the case where m2 is not *windows.Model
					// 	log.Fatal("m2 is not a *windows.Model")
					// }

					// cmds = append(cmds, cmd)
					// return m, tea.Batch(cmds...)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (k *KafkaReadWriteTabsComponent) Init() tea.Cmd {
	cmd := k.kafkaSendMessageTextAreaComponent.Init()
	return cmd
}

var (
	// Активный стиль (выбранный, в фокусе)
	ActiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")). // Фиолетовый фон
			Foreground(lipgloss.Color("#FFFFFF")). // Белый текст
			Padding(0, 2).                         // Отступы слева/справа
			MarginRight(1).                        // Отступ между элементами
			Bold(true)                             // Жирный шрифт

	// Неактивный стиль (для сравнения)
	InactiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3C3C3C")).
			Foreground(lipgloss.Color("#9B9B9B")).
			Padding(0, 2).
			MarginRight(1)
)

func (k *KafkaReadWriteTabsComponent) View() string {
	if k.styles == nil {
		return ""
	}

	doc := strings.Builder{}
	s := k.styles

	var renderedTabs []string

	for i, t := range k.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(k.Tabs)-1, i == int(k.activeTab)
		if isActive {
			style = s.activeTab
		} else {
			style = s.inactiveTab
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	// doc.WriteString(s.window.Width((lipgloss.Width(row))).Render(k.TabContent[k.activeTab]))

	var l string
	switch k.activeTab {
	case WRITE:
		taText := k.kafkaSendMessageTextAreaComponent.View()
		tableText := k.kafkaSendMessageTableComponent.View()

		switch k.activeComponentWriteTab {
		case TABS:
			taText = InactiveStyle.Render(taText)
			tableText = InactiveStyle.Render(tableText)

		case TEXT_AREA:
			taText = ActiveStyle.Render(taText)
			tableText = InactiveStyle.Render(tableText)
		case TABLE:
			taText = InactiveStyle.Render(taText)
			tableText = ActiveStyle.Render(tableText)
		}

		l = lipgloss.JoinVertical(lipgloss.Top, taText, tableText)
	case READ:
		l = lipgloss.JoinVertical(lipgloss.Top, k.kafkaReadMessageTable.View(), "")
	}

	doc.WriteString(s.window.Width((lipgloss.Width(row) * 3)).Render(l))

	return s.doc.Render(doc.String())
}

func CreateKafkaReadWriteTabsComponent(m *windows.Model, ks KafkaSendMessageTextArea, kt KafkaSendMessageTable,
	kafkaSender kafkaConsumerProducer, historySentMessages historySentMessages, kr KafkaReadMessageTable) *KafkaReadWriteTabsComponent {
	tabs := []string{"Read", "Write"}
	tabContent := []string{"Lip Gloss Tab", "Blush Tab", "Eye Shadow Tab", "Mascara Tab", "Foundation Tab"}
	k := &KafkaReadWriteTabsComponent{Tabs: tabs, TabContent: tabContent, styles: newStyles(true), model: m, kafkaSendMessageTextAreaComponent: ks,
		kafkaSendMessageTableComponent: kt, kafkaProducerConsumer: kafkaSender, historySentMessages: historySentMessages, kafkaReadMessageTable: kr,
		isLoadHistorySentMessages: atomic.Bool{}}
	return k
}
