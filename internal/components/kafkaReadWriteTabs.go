package components

import (
	"log"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"strings"

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

type activeComponent int

const (
	TABS activeComponent = iota
	TEXT_AREA
	TABLE
)

type KafkaReadWriteTabsComponent struct {
	styles                            *KafkaReadWriteTabsstyles
	Tabs                              []string
	TabContent                        []string
	activeTab                         activeTab
	activeComponent                   activeComponent
	model                             *windows.Model
	kafkaSendMessageTextAreaComponent KafkaSendMessageTextArea
	kafkaSendMessageTableComponent    KafkaSendMessageTable
}

type KafkaSendMessageTextArea interface {
	Update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd)
	View() string
	Init() tea.Cmd
	SetText(text string)
}

type KafkaSendMessageTable interface {
	Update(msg tea.Msg, model *windows.Model) (tea.Model, tea.Cmd)
	View() string
	Init() tea.Cmd
	GetActiveRow() []string
}

func (k *KafkaReadWriteTabsComponent) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	//добавить проверку, какая вкладка видима и направлять туда события
	if m.ActivePane == 3 {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch k.activeComponent {
			case TABS:
				switch keypress := msg.String(); keypress {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "right", "l", "n", "tab":
					k.activeTab = min(k.activeTab+1, activeTab(len(k.Tabs)-1))
					return m, nil
				case "left", "h", "p", "shift+tab":
					k.activeTab = max(k.activeTab-1, 0)
					return m, nil
				case "down":
					k.activeComponent = min(k.activeComponent+1, 3)
					return m, nil
				}
			}

			switch k.activeTab {
			case WRITE:
				switch k.activeComponent {
				case TEXT_AREA:
					switch keypress := msg.String(); keypress {
					case "down":
						k.activeComponent = min(k.activeComponent+1, 3)
						return m, nil
					case "ctrl+s":
						k.kafkaSendMessageTextAreaComponent.SetText("")
						return m, nil
					default:
						var ok bool
						m2, cmd := k.kafkaSendMessageTextAreaComponent.Update(msg, m)

						m, ok = m2.(*windows.Model)
						if !ok {
							// Handle the case where m2 is not *windows.Model
							log.Fatal("m2 is not a *windows.Model")
						}

						return m, tea.Batch(cmd)
					}
				case TABLE:
					switch keypress := msg.String(); keypress {
					case "enter":
						k.kafkaSendMessageTextAreaComponent.SetText(k.kafkaSendMessageTableComponent.GetActiveRow()[1])
						k.activeComponent = TEXT_AREA
					default:
						var ok bool
						m2, cmd := k.kafkaSendMessageTableComponent.Update(msg, m)

						m, ok = m2.(*windows.Model)
						if !ok {
							// Handle the case where m2 is not *windows.Model
							log.Fatal("m2 is not a *windows.Model")
						}

						return m, tea.Batch(cmd)
					}
				}
			case READ:
			}
		}
	}

	return m, nil
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

		switch k.activeComponent {
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
		l = lipgloss.JoinVertical(lipgloss.Top, "", "")
	}

	doc.WriteString(s.window.Width((lipgloss.Width(row) * 3)).Render(l))

	return s.doc.Render(doc.String())
}

func CreateKafkaReadWriteTabsComponent(m *windows.Model, ks KafkaSendMessageTextArea, kt KafkaSendMessageTable) *KafkaReadWriteTabsComponent {
	tabs := []string{"Read", "Write"}
	tabContent := []string{"Lip Gloss Tab", "Blush Tab", "Eye Shadow Tab", "Mascara Tab", "Foundation Tab"}
	k := &KafkaReadWriteTabsComponent{Tabs: tabs, TabContent: tabContent, styles: newStyles(true), model: m, kafkaSendMessageTextAreaComponent: ks,
		kafkaSendMessageTableComponent: kt}
	return k
}
