package readwritetabs

import (
	"log"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"reflect"
	"strings"
	"sync/atomic"

	"charm.land/bubbles/v2/spinner"
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
	TABS_WRITE activeComponentWriteTab = iota
	TEXT_AREA_WRITE
	TABLE_WRITE
)

type activeComponentReadTab int

const (
	TABS_READ activeComponentReadTab = iota
	TABLE_READ
)

type KafkaReadWriteTabsComponent struct {
	styles                            *KafkaReadWriteTabsstyles
	Tabs                              []string
	TabContent                        []string
	activeTab                         activeTab
	activeComponentWriteTab           activeComponentWriteTab
	activeComponentReadTab            activeComponentReadTab
	model                             *windows.Model
	kafkaSendMessageTextAreaComponent KafkaSendMessageTextArea
	kafkaSendMessageTableComponent    KafkaSendMessageTable
	kafkaReadMessageTable             KafkaReadMessageTable
	isLoadHistorySentMessages         atomic.Bool
	spinnerWhenLoadMessages           spinnerWhenLoadingKafkaMessages
	isSpinnerTick                     bool
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

type spinnerWhenLoadingKafkaMessages interface {
	startTick() tea.Msg
	update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd)
	View() string
}

func (k *KafkaReadWriteTabsComponent) IsFocusOnTabs() bool {
	return k.activeComponentReadTab == TABS_READ && k.activeComponentWriteTab == TABS_WRITE
}

func (k *KafkaReadWriteTabsComponent) clearReadTableKafkaMessages() {
	k.kafkaReadMessageTable.SetValues(make([]*contracts.ReadMessagesRow, 0))
}

func (k *KafkaReadWriteTabsComponent) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	//добавить проверку, какая вкладка видима и направлять туда события
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case contracts.StartSpinningWhenReadMessagesFromKafka:
		slog.Info("In KafkaReadWriteTabsComponent: contracts.StartSpinningWhenReadMessagesFromKafka")
		k.isSpinnerTick = true
		return m, k.spinnerWhenLoadMessages.startTick
	case contracts.StopSpinningWhenReadMessagesFromKafka:
		slog.Info("In KafkaReadWriteTabsComponent: contracts.StopSpinningWhenReadMessagesFromKafka")
		k.isSpinnerTick = false
		return m, nil
	case contracts.ClearHistoryMessagesTableEvent:
		slog.Info("In KafkaReadWriteTabsComponent: contracts.ClearHistoryMessagesTableEvent")
		k.clearReadTableKafkaMessages()
		return m, nil
	case contracts.UpdateListOfMessagesFromHistoryEvent:
		slog.Info("In KafkaReadWriteTabsComponent: contracts.UpdateListOfMessagesFromHistoryEvent")
		k.kafkaSendMessageTableComponent.SetValues(msg.Messages)
		return m, nil
	case contracts.UpdateListOfReadMessagesFromKafkaEvent:
		slog.Info("In KafkaReadWriteTabsComponent: contracts.UpdateListOfReadMessagesFromKafkaEvent")
		k.kafkaReadMessageTable.SetValues(msg.Messages)
		return m, nil
	case tea.KeyPressMsg:
		slog.Info("In KafkaReadWriteTabsComponent: tea.KeyPressMsg", "msg", msg,
			"k.activeTab", k.activeTab, "k.activeComponentWriteTab", k.activeComponentWriteTab)

		switch k.activeTab {
		case WRITE:
			switch k.activeComponentWriteTab {
			case TEXT_AREA_WRITE:
				switch keypress := msg.String(); keypress {
				case "down":
					k.activeComponentWriteTab = min(k.activeComponentWriteTab+1, 3)
					return m, tea.Batch(cmds...)
				case "esc":
					k.activeComponentWriteTab = max(k.activeComponentWriteTab-1, 0)
					return m, tea.Batch(cmds...)
				case "ctrl+s":
					textToSend := k.kafkaSendMessageTextAreaComponent.GetText()

					return m, func() tea.Msg {
						return contracts.SendMessageToKafka{Message: textToSend}
					}
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
			case TABLE_WRITE:
				switch keypress := msg.String(); keypress {
				case "enter":
					columns := k.kafkaSendMessageTableComponent.GetActiveRow()
					if len(columns) > 0 {
						k.kafkaSendMessageTextAreaComponent.SetText(k.kafkaSendMessageTableComponent.GetActiveRow()[0])
					}
					k.activeComponentWriteTab = TEXT_AREA_WRITE
					return m, tea.Batch(cmds...)
				case "esc":
					k.activeComponentWriteTab = max(k.activeComponentWriteTab-1, 0)
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
			case TABS_WRITE:
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
				case "esc":
				}
			}
		case READ:
			switch k.activeComponentReadTab {
			case TABS_READ:
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
					k.activeComponentReadTab = min(k.activeComponentReadTab+1, 1)
					return m, tea.Batch(cmds...)
				default:
					slog.Error("TABS_READ", "k", reflect.TypeOf(msg), "m", msg, "f", k.activeTab, "f", k.activeComponentReadTab)
				}
			case TABLE_READ:
				switch keypress := msg.String(); keypress {
				case "esc":
					k.activeComponentReadTab = max(k.activeComponentReadTab-1, 0)
					return m, tea.Batch(cmds...)
				case "ctrl+s":
					slog.Info("In KafkaReadWriteTabsComponent: send message to kafka")

					return m, func() tea.Msg {
						slog.Info("In KafkaReadWriteTabsComponent: emit event contracts.UserWantsReadMessagesFromKafka")

						return contracts.UserWantsReadMessagesFromKafka{}
					}
				default:
					var ok bool
					m2, cmd := k.kafkaReadMessageTable.Update(msg, m)

					m, ok = m2.(*windows.Model)
					if !ok {
						// Handle the case where m2 is not *windows.Model
						log.Fatal("m2 is not a *windows.Model")
					}

					cmds = append(cmds, cmd)
					return m, tea.Batch(cmds...)
				}
			}
		}
	case contracts.CloseWindowForDetailedReadMessage:
		var ok bool
		m2, cmd := k.kafkaReadMessageTable.Update(msg, m)

		m, ok = m2.(*windows.Model)
		if !ok {
			// Handle the case where m2 is not *windows.Model
			log.Fatal("m2 is not a *windows.Model")
		}

		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case spinner.TickMsg:
		if k.isSpinnerTick == false {
			return m, tea.Batch(cmds...)
		}

		m, cmd := k.spinnerWhenLoadMessages.update(msg, m)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
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
		case TABS_WRITE:
			taText = InactiveStyle.Render(taText)
			tableText = InactiveStyle.Render(tableText)

		case TEXT_AREA_WRITE:
			taText = ActiveStyle.Render(taText)
			tableText = InactiveStyle.Render(tableText)
		case TABLE_WRITE:
			taText = InactiveStyle.Render(taText)
			tableText = ActiveStyle.Render(tableText)
		}

		l = lipgloss.JoinVertical(lipgloss.Top, taText, tableText)
	case READ:
		var spinnerString string
		if k.isSpinnerTick == true {
			spinnerString = k.spinnerWhenLoadMessages.View()
		} else {
			spinnerString = ""
		}

		l = lipgloss.JoinVertical(lipgloss.Top, spinnerString, k.kafkaReadMessageTable.View())
	}

	doc.WriteString(s.window.Width((lipgloss.Width(row) * 4)).Render(l))

	return s.doc.Render(doc.String())
}

func CreateKafkaReadWriteTabsComponent(m *windows.Model, ks KafkaSendMessageTextArea, kt KafkaSendMessageTable, kr KafkaReadMessageTable,
	spinnerWhenLoadMessages spinnerWhenLoadingKafkaMessages) *KafkaReadWriteTabsComponent {
	tabs := []string{"Read", "Write"}
	tabContent := []string{"Lip Gloss Tab", "Blush Tab", "Eye Shadow Tab", "Mascara Tab", "Foundation Tab"}
	k := &KafkaReadWriteTabsComponent{Tabs: tabs, TabContent: tabContent, styles: newStyles(true), model: m, kafkaSendMessageTextAreaComponent: ks,
		kafkaSendMessageTableComponent: kt, kafkaReadMessageTable: kr,
		isLoadHistorySentMessages: atomic.Bool{}, spinnerWhenLoadMessages: spinnerWhenLoadMessages}
	return k
}
