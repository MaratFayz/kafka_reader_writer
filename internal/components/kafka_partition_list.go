package components

import (
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"strconv"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaPartitionList struct {
	List         *list.Model
	DelegateKeys *DelegateKeyMapKafkaPartition
	ListKeys     *ListKeyMapKafkaPartition
	Styles       *StylesKafkaPartition
}

func (k *KafkaPartitionList) GetList() *list.Model {
	return k.List
}

func (k *KafkaPartitionList) GetStyles() windows.StylesKafkaCluster {
	return k.Styles
}

type ufKafkaPartition struct {
	keys   *DelegateKeyMapKafkaPartition
	styles *StylesKafkaPartition
}

func (u *ufKafkaPartition) updateFuncKafkaPartition(msg tea.Msg, m *list.Model) tea.Cmd {
	var partition string

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, u.keys.Choose):
			slog.Info("KafkaPartitionList, user chose partition", "partition", msg.Keystroke())

			if i, ok := m.SelectedItem().(ItemKafkaPartition); ok {
				partition = i.Title()

				return func() tea.Msg {
					return contracts.UserChosenPartition{Partition: partition}
				}
			} else {
				return nil
			}
		}
	}

	return nil
}

func newItemDelegateKafkaPartition(keys *DelegateKeyMapKafkaPartition, styles *StylesKafkaPartition) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	uf := ufKafkaPartition{keys: keys, styles: styles}

	d.UpdateFunc = uf.updateFuncKafkaPartition

	help := []key.Binding{keys.Choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type DelegateKeyMapKafkaPartition struct {
	Choose key.Binding
	Remove key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d DelegateKeyMapKafkaPartition) ShortHelp() []key.Binding {
	return []key.Binding{
		d.Choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d DelegateKeyMapKafkaPartition) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.Choose,
		},
	}
}

func NewDelegateKeyMapKafkaPartition() *DelegateKeyMapKafkaPartition {
	return &DelegateKeyMapKafkaPartition{
		Choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}

func newListKeyMapKafkaPartition() *ListKeyMapKafkaPartition {
	return &ListKeyMapKafkaPartition{}
}

type ListKeyMapKafkaPartition struct {
}

type StylesKafkaPartition struct {
	App           lipgloss.Style
	Title         lipgloss.Style
	StatusMessage lipgloss.Style
}

func (s *StylesKafkaPartition) GetApp() lipgloss.Style {
	return s.App
}
func (s *StylesKafkaPartition) GetTitle() lipgloss.Style         { return s.Title }
func (s *StylesKafkaPartition) GetStatusMessage() lipgloss.Style { return s.StatusMessage }

func newStylesKafkaPartition(darkBG bool) StylesKafkaPartition {
	lightDark := lipgloss.LightDark(darkBG)

	return StylesKafkaPartition{
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

type ItemKafkaPartition struct {
	title       string
	description string
}

func (i ItemKafkaPartition) Title() string       { return i.title }
func (i ItemKafkaPartition) Description() string { return i.description }
func (i ItemKafkaPartition) FilterValue() string { return i.title }

func NewItemKafkaPartition(title string, description string) ItemKafkaPartition {
	return ItemKafkaPartition{title: title, description: description}
}

func CreateKafkaPartitionsList() *KafkaPartitionList {
	styles := newStylesKafkaPartition(false) // default to dark background styles

	delegateKeys := NewDelegateKeyMapKafkaPartition()
	listKeys := newListKeyMapKafkaPartition()

	items := make([]list.Item, 0)

	delegate := newItemDelegateKafkaPartition(delegateKeys, &styles)
	list := list.New(items, delegate, 0, 0)
	list.Title = "Kafka Partitions"
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{}
	}
	list.SetShowHelp(false)

	return &KafkaPartitionList{&list, delegateKeys, listKeys, &styles}
}

func (kt *KafkaPartitionList) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	kcl := kt.List
	delegateKeys := kt.DelegateKeys

	switch msg := msg.(type) {
	case contracts.AskForHelp:
		slog.Info("in kafkaPartitionList: event contracts.AskForHelp")
		values := make([]contracts.KeyValue, 0)
		values = append(values, contracts.KeyValue{Key: "?", Value: "Help window"})
		// values = append(values, contracts.KeyValue{Key: "a", Value: "Append new cluster"})
		// values = append(values, contracts.KeyValue{Key: "x", Value: "Remove selected cluster"})

		return m, contracts.CreateHelpMessageEvent(values)
	case contracts.UpdateListOfPartitionsEvent:
		delegateKeys.Remove.SetEnabled(true)
		kcl.SetItems(make([]list.Item, 0))
		for i, sm := range msg.Partitions {
			newItem := NewItemKafkaPartition(strconv.Itoa(sm), "")

			insCmd := kcl.InsertItem(i, newItem)
			cmds = append(cmds, insCmd)
		}

		return m, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		if kcl.FilterState() == list.Filtering {
			break
		}
	}

	newListModel, cmd := kcl.Update(msg)
	kt.List = &newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
