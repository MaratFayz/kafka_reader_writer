package components

import (
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/debug"
	"marat/fayz/kafka_reader_writer/internal/windows"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaClusterList struct {
	List         *list.Model
	DelegateKeys *DelegateKeyMapKafkaCluster
	ListKeys     *ListKeyMapKafkaCluster
	Styles       *StylesKafkaCluster
}

func (k *KafkaClusterList) GetList() *list.Model {
	return k.List
}

func (k *KafkaClusterList) GetStyles() windows.StylesKafkaCluster {
	return k.Styles
}

type ModelChangerKafkaCluster interface {
	KafkaClusterChosenNextActivePane() tea.Cmd
	SetChosenKafkaCluster(name string)
}

type ufKafkaCluster struct {
	model  ModelChangerKafkaCluster
	keys   *DelegateKeyMapKafkaCluster
	styles *StylesKafkaCluster
}

func (u *ufKafkaCluster) updateFunc(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, u.keys.Choose):
			if i, ok := m.SelectedItem().(ItemKafkaCluster); ok {
				title = i.Title()
				u.model.SetChosenKafkaCluster(title)
				cmd := u.model.KafkaClusterChosenNextActivePane()
				cmds = append(cmds, cmd)
			} else {
				return nil
			}

			return tea.Batch(cmds...)

		case key.Matches(msg, u.keys.Remove):
			//TODO add вызов окна удаления
		}
	}

	return nil
}

func newItemDelegateKafkaCluster(keys *DelegateKeyMapKafkaCluster, styles *StylesKafkaCluster, model ModelChangerKafkaCluster) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	uf := ufKafkaCluster{model: model, keys: keys, styles: styles}

	d.UpdateFunc = uf.updateFunc
	return d
}

type DelegateKeyMapKafkaCluster struct {
	Choose key.Binding
	Remove key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d DelegateKeyMapKafkaCluster) ShortHelp() []key.Binding {
	return []key.Binding{
		d.Choose,
		d.Remove,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d DelegateKeyMapKafkaCluster) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.Choose,
			d.Remove,
		},
	}
}

func NewDelegateKeyMapKafkaCluster() *DelegateKeyMapKafkaCluster {
	return &DelegateKeyMapKafkaCluster{
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

func newListKeyMapKafkaCluster() *ListKeyMapKafkaCluster {
	return &ListKeyMapKafkaCluster{
		InsertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
	}
}

type ListKeyMapKafkaCluster struct {
	InsertItem key.Binding
}

type StylesKafkaCluster struct {
	App           lipgloss.Style
	Title         lipgloss.Style
	StatusMessage lipgloss.Style
}

func (s *StylesKafkaCluster) GetApp() lipgloss.Style {
	return s.App
}
func (s *StylesKafkaCluster) GetTitle() lipgloss.Style         { return s.Title }
func (s *StylesKafkaCluster) GetStatusMessage() lipgloss.Style { return s.StatusMessage }

func newStylesKafkaCluster(darkBG bool) StylesKafkaCluster {
	lightDark := lipgloss.LightDark(darkBG)

	return StylesKafkaCluster{
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

type ItemKafkaCluster struct {
	title       string
	description string
}

func (i ItemKafkaCluster) Title() string       { return i.title }
func (i ItemKafkaCluster) Description() string { return i.description }
func (i ItemKafkaCluster) FilterValue() string { return i.title }

func NewItemKafkaCluster(title string, description string) ItemKafkaCluster {
	return ItemKafkaCluster{title: title, description: description}
}

func CreateKafkaClustersList(model ModelChangerKafkaCluster) *KafkaClusterList {
	styles := newStylesKafkaCluster(false) // default to dark background styles

	delegateKeys := NewDelegateKeyMapKafkaCluster()
	listKeys := newListKeyMapKafkaCluster()

	items := make([]list.Item, 0)

	delegate := newItemDelegateKafkaCluster(delegateKeys, &styles, model)
	list := list.New(items, delegate, 0, 0)
	list.Title = "Kafka Clusters"
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.InsertItem,
		}
	}
	list.SetShowHelp(false)

	return &KafkaClusterList{&list, delegateKeys, listKeys, &styles}
}

type LocalStorage interface {
	GetKafkaClusters() []*contracts.KafkaCluster
}

func CreateKafkaClustersListAddValues(ls LocalStorage, model *windows.Model) *KafkaClusterList {
	kafkaClusterList := CreateKafkaClustersList(model)

	kc := ls.GetKafkaClusters()

	clusterList := make([]*contracts.KafkaCluster, len(kc))
	clusterMap := make(map[string]*contracts.KafkaCluster, len(kc))
	for i, cluster := range kc {
		clusterList[i] = cluster // Каждый элемент преобразуется отдельно
		clusterMap[cluster.Title] = cluster
	}
	model.SetClusters(clusterMap)
	for i, v := range clusterList {
		kafkaClusterList.List.InsertItem(i, NewItemKafkaCluster(v.Title, v.Url))
	}

	return kafkaClusterList
}

func (kc *KafkaClusterList) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	kcl := kc.List
	// keys := kc.ListKeys

	switch msg.(type) {

	case contracts.AskForHelp:
		debug.WriteDebugFile("in kafkaClusterList: event contracts.AskForHelp")
		values := make([]contracts.KeyValue, 0)
		values = append(values, contracts.KeyValue{Key: "?", Value: "Help window"})
		values = append(values, contracts.KeyValue{Key: "a", Value: "Append new cluster"})
		values = append(values, contracts.KeyValue{Key: "x", Value: "Remove selected cluster"})

		return m, contracts.CreateHelpMessageEvent(values)

		// case contracts.UserWantsClusterFromList:
		// 				index := m.Index()
		// 		m.RemoveItem(index)
		// 		if len(m.Items()) == 0 {
		// 			u.keys.Remove.SetEnabled(false)
		// 		}
		// 		return m.NewStatusMessage(u.styles.StatusMessage.Render("Deleted " + title))

		// case tea.KeyPressMsg:
		// 	if kcl.FilterState() == list.Filtering {
		// 		break
		// 	}

		// 	switch {
		// 	case key.Matches(msg, keys.InsertItem): //TODO add открыть окно добавления кластера
		// 		newItem := NewItemKafkaCluster("aaa", "bbb")
		// 		insCmd := kcl.InsertItem(0, newItem)
		// 		return m, insCmd
		// 	}
	}

	newListModel, cmd := kcl.Update(msg)
	kc.List = &newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
