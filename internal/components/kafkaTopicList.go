package components

import (
	"fmt"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaTopicList struct {
	List                   *list.Model
	DelegateKeys           *DelegateKeyMapKafkaTopic
	ListKeys               *ListKeyMapKafkaTopic
	Styles                 *StylesKafkaTopic
	model                  ModelChangerKafkaTopic
	kafkaConnectorProvider KafkaTopicsProvider
}

func (k *KafkaTopicList) GetList() *list.Model {
	return k.List
}

func (k *KafkaTopicList) GetStyles() windows.StylesKafkaCluster {
	return k.Styles
}

type ModelChangerKafkaTopic interface {
	KafkaTopicChosenNextActivePane() tea.Cmd
	SetChosenKafkaTopic(name string)
	SetTopicsForCluster(clusterName string, topics []string)
}

type KafkaTopicsProvider interface {
	GetTopicsByClusterName(clusterName *contracts.KafkaCluster) ([]string, error)
}

type ufKafkaTopic struct {
	model  ModelChangerKafkaTopic
	keys   *DelegateKeyMapKafkaTopic
	styles *StylesKafkaTopic
}

func (u *ufKafkaTopic) updateFuncKafkaTopic(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, u.keys.Choose):
			cmd := u.model.KafkaTopicChosenNextActivePane()

			if i, ok := m.SelectedItem().(ItemKafkaTopic); ok {
				// title = i.Description()
				title = i.Title()

				u.model.SetChosenKafkaTopic(title)
			} else {
				return nil
			}

			return tea.Batch(m.StartSpinner(), m.NewStatusMessage(u.styles.StatusMessage.Render("You chose "+title+"; pane 1")), cmd)

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

func newItemDelegateKafkaTopic(keys *DelegateKeyMapKafkaTopic, styles *StylesKafkaTopic, model ModelChangerKafkaTopic) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	uf := ufKafkaTopic{model: model, keys: keys, styles: styles}

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

func CreateKafkaTopicsList(model ModelChangerKafkaTopic, kafkaConnectorProvider KafkaTopicsProvider) *KafkaTopicList {
	styles := newStylesKafkaTopic(false) // default to dark background styles

	delegateKeys := NewDelegateKeyMapKafkaTopic()
	listKeys := newListKeyMapKafkaTopic()

	items := make([]list.Item, 0)

	delegate := newItemDelegateKafkaTopic(delegateKeys, &styles, model)
	list := list.New(items, delegate, 0, 0)
	list.Title = "Kafka Topics"
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

	return &KafkaTopicList{&list, delegateKeys, listKeys, &styles, model, kafkaConnectorProvider}
}

type initList interface {
	IsInitList() bool
}

func (kt *KafkaTopicList) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// if m.ActivePane == 1 {
	kcl := kt.List
	keys := kt.ListKeys
	delegateKeys := kt.DelegateKeys

	switch msg := msg.(type) {
	case initList:
		v, ok := m.IsLoadTopics[m.SelectedKafkaCluster]

		if !ok {
			m.IsLoadTopics[m.SelectedKafkaCluster] = false
			v = m.IsLoadTopics[m.SelectedKafkaCluster]
		}

		if v == false {
			cmd := kt.loadTopics(m.GetClusterByTitle(m.SelectedKafkaCluster))
			cmds = append(cmds, cmd)
			m.IsLoadTopics[m.SelectedKafkaCluster] = true
		} else {
			cmd := kt.showTopics()
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case loadedKafkaTopicsMsg:
		kt.model.SetTopicsForCluster(msg.cluster, msg.topics)
		cmd := func() tea.Msg {
			return showKafkaTopicsMsg{}
		}
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	case showKafkaTopicsMsg:
		delegateKeys.Remove.SetEnabled(true)
		kcl.SetItems(make([]list.Item, 0))
		for i, sm := range m.GetKafkaTopics(m.SelectedKafkaCluster) {
			newItem := NewItemKafkaTopic(sm, "")

			insCmd := kcl.InsertItem(i, newItem)
			cmds = append(cmds, insCmd)
		}

		statusCmd := kcl.NewStatusMessage(fmt.Sprintf("Added %d items", len(cmds)))
		cmds = append(cmds, statusCmd)
		return m, tea.Batch(cmds...)
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
			newItem := NewItemKafkaTopic("aaa", "bbb")
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
	// } else if m.ActivePane != 1 {
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

type loadedKafkaTopicsMsg struct {
	cluster string
	topics  []string
}

type showKafkaTopicsMsg struct {
}

type loadingKafkaTopicErrMsg struct {
	err error
}

func (e loadingKafkaTopicErrMsg) Error() string { return e.err.Error() }

func (ktl *KafkaTopicList) loadTopics(cluster *contracts.KafkaCluster) tea.Cmd {
	return func() tea.Msg {
		topicNames, err := ktl.kafkaConnectorProvider.GetTopicsByClusterName(cluster)
		if err != nil {
			slog.Error("Error during getting topic names", "topicNames", err)
			return loadingKafkaTopicErrMsg{err: err}
		}

		return loadedKafkaTopicsMsg{cluster: cluster.Title, topics: topicNames}
	}
}

func (ktl *KafkaTopicList) showTopics() tea.Cmd {
	return func() tea.Msg {
		return showKafkaTopicsMsg{}
	}
}
