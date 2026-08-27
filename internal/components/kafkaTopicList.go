package components

import (
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaTopicList struct {
	List         *list.Model
	DelegateKeys *DelegateKeyMapKafkaTopic
	ListKeys     *ListKeyMapKafkaTopic
	Styles       *StylesKafkaTopic
}

func (k *KafkaTopicList) GetList() *list.Model {
	return k.List
}

func (k *KafkaTopicList) GetStyles() windows.StylesKafkaCluster {
	return k.Styles
}

type ufKafkaTopic struct {
	keys   *DelegateKeyMapKafkaTopic
	styles *StylesKafkaTopic
}

func (u *ufKafkaTopic) updateFuncKafkaTopic(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, u.keys.Choose):
			slog.Info("KafkaTopicList, user chose topic", "topic", msg.Keystroke())

			if i, ok := m.SelectedItem().(ItemKafkaTopic); ok {
				title = i.Title()

				return func() tea.Msg {
					return contracts.UserChosenKafkaTopic{Topic: title}
				}
			} else {
				return nil
			}
		}
	}

	return nil
}

func newItemDelegateKafkaTopic(keys *DelegateKeyMapKafkaTopic, styles *StylesKafkaTopic) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	uf := ufKafkaTopic{keys: keys, styles: styles}

	d.UpdateFunc = uf.updateFuncKafkaTopic

	help := []key.Binding{keys.Choose, keys.Remove}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type DelegateKeyMapKafkaTopic struct {
	Choose key.Binding
	Remove key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d DelegateKeyMapKafkaTopic) ShortHelp() []key.Binding {
	return []key.Binding{
		d.Choose,
		d.Remove,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d DelegateKeyMapKafkaTopic) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.Choose,
			d.Remove,
		},
	}
}

func NewDelegateKeyMapKafkaTopic() *DelegateKeyMapKafkaTopic {
	return &DelegateKeyMapKafkaTopic{
		Choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		Remove: key.NewBinding(
			key.WithKeys("x", "backspace"),
			key.WithHelp("x", "delete"),
		),
	}
}

func newListKeyMapKafkaTopic() *ListKeyMapKafkaTopic {
	return &ListKeyMapKafkaTopic{
		InsertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		ToggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		ToggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		ToggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		TogglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		ToggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type ListKeyMapKafkaTopic struct {
	ToggleSpinner    key.Binding
	ToggleTitleBar   key.Binding
	ToggleStatusBar  key.Binding
	TogglePagination key.Binding
	ToggleHelpMenu   key.Binding
	InsertItem       key.Binding
}

type StylesKafkaTopic struct {
	App           lipgloss.Style
	Title         lipgloss.Style
	StatusMessage lipgloss.Style
}

func (s *StylesKafkaTopic) GetApp() lipgloss.Style {
	return s.App
}
func (s *StylesKafkaTopic) GetTitle() lipgloss.Style         { return s.Title }
func (s *StylesKafkaTopic) GetStatusMessage() lipgloss.Style { return s.StatusMessage }

func newStylesKafkaTopic(darkBG bool) StylesKafkaTopic {
	lightDark := lipgloss.LightDark(darkBG)

	return StylesKafkaTopic{
		App: lipgloss.NewStyle().
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1),
		StatusMessage: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#04B575"), lipgloss.Color("#04B575"))),
	}
}

type ItemKafkaTopic struct {
	title       string
	description string
}

func (i ItemKafkaTopic) Title() string       { return i.title }
func (i ItemKafkaTopic) Description() string { return i.description }
func (i ItemKafkaTopic) FilterValue() string { return i.title }

func NewItemKafkaTopic(title string, description string) ItemKafkaTopic {
	return ItemKafkaTopic{title: title, description: description}
}

func CreateKafkaTopicsList() *KafkaTopicList {
	styles := newStylesKafkaTopic(false) // default to dark background styles

	delegateKeys := NewDelegateKeyMapKafkaTopic()
	listKeys := newListKeyMapKafkaTopic()

	items := make([]list.Item, 0)

	delegate := newItemDelegateKafkaTopic(delegateKeys, &styles)
	list := list.New(items, delegate, 0, 0)
	list.Title = "Kafka Topics"
	list.SetShowHelp(false)

	return &KafkaTopicList{&list, delegateKeys, listKeys, &styles}
}

func (kt *KafkaTopicList) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	kcl := kt.List
	delegateKeys := kt.DelegateKeys

	switch msg := msg.(type) {
	case contracts.AskForHelp:
		slog.Info("in kafkaTopicList: event contracts.AskForHelp")
		values := make([]contracts.KeyValue, 0)
		values = append(values, contracts.KeyValue{Key: "?", Value: "Help window"})
		// values = append(values, contracts.KeyValue{Key: "a", Value: "Append new cluster"})
		// values = append(values, contracts.KeyValue{Key: "x", Value: "Remove selected cluster"})

		return m, contracts.CreateHelpMessageEvent(values)
	case contracts.UpdateListOfTopicsEvent:
		delegateKeys.Remove.SetEnabled(true)
		kcl.SetItems(make([]list.Item, 0))
		for i, sm := range msg.Topics {
			newItem := NewItemKafkaTopic(sm, "")

			insCmd := kcl.InsertItem(i, newItem)
			cmds = append(cmds, insCmd)
		}

		return m, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		// Don't match any of the keys below if we're actively filtering.
		if kcl.FilterState() == list.Filtering {
			break
		}
	}

	newListModel, cmd := kcl.Update(msg)
	kt.List = &newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
