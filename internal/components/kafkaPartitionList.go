package components

import (
	"fmt"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"strconv"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaPartitionList struct {
	List                    *list.Model
	DelegateKeys            *DelegateKeyMapKafkaPartition
	ListKeys                *ListKeyMapKafkaPartition
	Styles                  *StylesKafkaPartition
	model                   ModelChangerKafkaPartition
	kafkaPartitionsProvider KafkaPartitionsProvider
}

func (k *KafkaPartitionList) GetList() *list.Model {
	return k.List
}

func (k *KafkaPartitionList) GetStyles() windows.StylesKafkaCluster {
	return k.Styles
}

type KafkaPartitionsProvider interface {
	GetPartitionsByClusterNameAndTopic(topicName string, cluster *contracts.KafkaCluster) []int
}

type ModelChangerKafkaPartition interface {
	KafkaPartitionChosenNextActivePane()
	SetChosenKafkaPartition(name string)
	SetPartitionsForClusterAndTopic(clusterName string, topicName string, partitions []int)
}

type ufKafkaPartition struct {
	model  ModelChangerKafkaPartition
	keys   *DelegateKeyMapKafkaPartition
	styles *StylesKafkaPartition
}

func (u *ufKafkaPartition) updateFuncKafkaPartition(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, u.keys.Choose):
			u.model.KafkaPartitionChosenNextActivePane()

			if i, ok := m.SelectedItem().(ItemKafkaPartition); ok {
				title = i.Title()
				u.model.SetChosenKafkaPartition(title)
			} else {
				return nil
			}

			return tea.Batch(m.StartSpinner(), m.NewStatusMessage(u.styles.StatusMessage.Render("You chose "+title+"; pane 1")))

		case key.Matches(msg, u.keys.Remove):
			index := m.Index()
			m.RemoveItem(index)
			if len(m.Items()) == 0 {
				u.keys.Remove.SetEnabled(false)
			}
			return m.NewStatusMessage(u.styles.StatusMessage.Render("Deleted " + title))
		}
	}

	return nil
}

func newItemDelegateKafkaPartition(keys *DelegateKeyMapKafkaPartition, styles *StylesKafkaPartition, model ModelChangerKafkaPartition) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	uf := ufKafkaPartition{model: model, keys: keys, styles: styles}

	d.UpdateFunc = uf.updateFuncKafkaPartition

	help := []key.Binding{keys.Choose, keys.Remove}

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
		d.Remove,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d DelegateKeyMapKafkaPartition) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.Choose,
			d.Remove,
		},
	}
}

func NewDelegateKeyMapKafkaPartition() *DelegateKeyMapKafkaPartition {
	return &DelegateKeyMapKafkaPartition{
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

func newListKeyMapKafkaPartition() *ListKeyMapKafkaPartition {
	return &ListKeyMapKafkaPartition{
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

type ListKeyMapKafkaPartition struct {
	ToggleSpinner    key.Binding
	ToggleTitleBar   key.Binding
	ToggleStatusBar  key.Binding
	TogglePagination key.Binding
	ToggleHelpMenu   key.Binding
	InsertItem       key.Binding
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

func CreateKafkaPartitionsList(model ModelChangerKafkaPartition, kafkaPartitionsProvider KafkaPartitionsProvider) *KafkaPartitionList {
	styles := newStylesKafkaPartition(false) // default to dark background styles

	delegateKeys := NewDelegateKeyMapKafkaPartition()
	listKeys := newListKeyMapKafkaPartition()

	items := make([]list.Item, 0)

	delegate := newItemDelegateKafkaPartition(delegateKeys, &styles, model)
	list := list.New(items, delegate, 0, 0)
	list.Title = "Kafka Partitions"
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

	return &KafkaPartitionList{&list, delegateKeys, listKeys, &styles, model, kafkaPartitionsProvider}
}

func (kt *KafkaPartitionList) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// if m.ActivePane == 2 {
	kcl := kt.List
	keys := kt.ListKeys
	delegateKeys := kt.DelegateKeys

	switch msg := msg.(type) {
	case initList:
		mapTopics, ok := m.IsLoadPartitions[m.SelectedKafkaCluster]

		if !ok {
			m.IsLoadPartitions[m.SelectedKafkaCluster] = make(map[string]bool)
			mapTopics = m.IsLoadPartitions[m.SelectedKafkaCluster]
		}

		arePartitionsLoadedForClusterTopic, ok := mapTopics[m.SelectedKafkaTopic]

		if !ok {
			mapTopics[m.SelectedKafkaTopic] = false
			arePartitionsLoadedForClusterTopic = mapTopics[m.SelectedKafkaTopic]
		}

		if arePartitionsLoadedForClusterTopic == false {
			cmd := kt.loadPartitions(m.GetClusterByTitle(m.SelectedKafkaCluster), m.SelectedKafkaTopic)
			cmds = append(cmds, cmd)
			m.IsLoadPartitions[m.SelectedKafkaCluster][m.SelectedKafkaTopic] = true
		} else {
			cmd := kt.showPartitions()
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case loadedKafkaPartitionsMsg:
		kt.model.SetPartitionsForClusterAndTopic(msg.cluster, msg.topic, msg.partitions)
		cmd := func() tea.Msg {
			return showKafkaPartitionsMsg{}
		}
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	case showKafkaPartitionsMsg:
		delegateKeys.Remove.SetEnabled(true)
		kcl.SetItems(make([]list.Item, 0))
		for i, sm := range m.GetKafkaPartitions(m.SelectedKafkaCluster, m.SelectedKafkaTopic) {
			newItem := NewItemKafkaPartition(strconv.Itoa(sm), "")

			insCmd := kcl.InsertItem(i, newItem)
			cmds = append(cmds, insCmd)
		}

		statusCmd := kcl.NewStatusMessage(fmt.Sprintf("Added %d items", len(cmds)))
		cmds = append(cmds, statusCmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Don't match any of the keys below if we're actively filtering.
		if kcl.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, keys.ToggleSpinner):
			cmd := kcl.ToggleSpinner()
			// statusCmd := kcl.NewStatusMessage("Pane " + fmt.Sprint(m.ActivePane))
			// return m, tea.Batch(cmd, statusCmd)
			return m, cmd
		case key.Matches(msg, keys.ToggleTitleBar):
			v := !kcl.ShowTitle()
			kcl.SetShowTitle(v)
			kcl.SetShowFilter(v)
			kcl.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, keys.ToggleStatusBar):
			kcl.SetShowStatusBar(!kcl.ShowStatusBar())
			return m, nil

		case key.Matches(msg, keys.TogglePagination):
			kcl.SetShowPagination(!kcl.ShowPagination())
			return m, nil

		case key.Matches(msg, keys.ToggleHelpMenu):
			kcl.SetShowHelp(!kcl.ShowHelp())
			return m, nil

		case key.Matches(msg, keys.InsertItem):
			delegateKeys.Remove.SetEnabled(true)
			// newItem := m.itemGenerator.next()
			newItem := NewItemKafkaPartition("aaa", "bbb")
			insCmd := kcl.InsertItem(0, newItem)
			// statusCmd := kcl.NewStatusMessage("Added " + newItem.Title() + ", pane " + fmt.Sprint(m.ActivePane))
			// return m, tea.Batch(insCmd, statusCmd)
			return m, insCmd
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := kcl.Update(msg)
	kt.List = &newListModel
	cmds = append(cmds, cmd)
	// } else if m.ActivePane != 2 {
	// 	// slog.Info("XXXXXX", "pane", m.ActivePane)
	// 	switch msg := msg.(type) {
	// 	case spinner.TickMsg:
	// 		newListModel, cmd := kt.List.Update(msg)
	// 		kt.List = &newListModel
	// 		cmds = append(cmds, cmd)
	// 		return m, tea.Batch(cmds...)
	// 	}
	// }

	return m, tea.Batch(cmds...)
}

type loadedKafkaPartitionsMsg struct {
	cluster    string
	topic      string
	partitions []int
}

type showKafkaPartitionsMsg struct{}
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (ktl *KafkaPartitionList) loadPartitions(cluster *contracts.KafkaCluster, topicName string) tea.Cmd {
	return func() tea.Msg {
		partitions := ktl.kafkaPartitionsProvider.GetPartitionsByClusterNameAndTopic(topicName, cluster)

		return loadedKafkaPartitionsMsg{cluster: cluster.Title, topic: topicName, partitions: partitions}
	}
}

func (ktl *KafkaPartitionList) showPartitions() tea.Cmd {
	return func() tea.Msg {
		return showKafkaPartitionsMsg{}
	}
}
