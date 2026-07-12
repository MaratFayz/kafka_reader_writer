package components

import (
	"fmt"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/windows"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
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
	SetActivePaneAfterKafkaClusterChosen(activePane int)
	SetChosenKafkaCluster(name string)
}

type ufKafkaCluster struct {
	model  ModelChangerKafkaCluster
	keys   *DelegateKeyMapKafkaCluster
	styles *StylesKafkaCluster
}

func (u *ufKafkaCluster) updateFunc(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, u.keys.Choose):
			u.model.SetActivePaneAfterKafkaClusterChosen(1)
			if i, ok := m.SelectedItem().(ItemKafkaCluster); ok {
				title = i.Title()
				u.model.SetChosenKafkaCluster(title)
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

func newItemDelegateKafkaCluster(keys *DelegateKeyMapKafkaCluster, styles *StylesKafkaCluster, model ModelChangerKafkaCluster) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	uf := ufKafkaCluster{model: model, keys: keys, styles: styles}

	d.UpdateFunc = uf.updateFunc

	help := []key.Binding{keys.Choose, keys.Remove}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

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

type ListKeyMapKafkaCluster struct {
	ToggleSpinner    key.Binding
	ToggleTitleBar   key.Binding
	ToggleStatusBar  key.Binding
	TogglePagination key.Binding
	ToggleHelpMenu   key.Binding
	InsertItem       key.Binding
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

	return &KafkaClusterList{&list, delegateKeys, listKeys, &styles}
}

type LocalStorage interface {
	GetKafkaClusters() []*contracts.KafkaCluster
}

// type KafkaCluster interface {
// 	Title() string
// 	Url() string
// 	TrustStorePath() string
// 	TrustStorePassword() string
// 	Username() string
// 	Password() string
// 	SaslMechanism() string
// }

func CreateKafkaClustersListAddValues(ls LocalStorage, model *windows.Model) *KafkaClusterList {
	kafkaClusterList := CreateKafkaClustersList(model)

	kc := ls.GetKafkaClusters()

	clusterList := make([]*contracts.KafkaCluster, len(kc))
	clusterMap := make(map[string]*contracts.KafkaCluster, len(kc))
	for i, cluster := range kc {
		clusterList[i] = cluster // Каждый элемент преобразуется отдельно
		clusterMap[cluster.Title] = cluster
	}
	model.KafkaClusters = clusterMap
	// fmt.Println(clusterList)
	for i, v := range clusterList {
		kafkaClusterList.List.InsertItem(i, NewItemKafkaCluster(v.Title, v.Url))
	}

	return kafkaClusterList
}

func (kc *KafkaClusterList) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.ActivePane == 0 {
		kcl := kc.List
		keys := kc.ListKeys
		delegateKeys := kc.DelegateKeys

		switch msg := msg.(type) {
		case tea.BackgroundColorMsg:
			// m.darkBG = msg.IsDark()
			// m.updateListProperties()
			fmt.Printf("%s", msg)
			return m, nil

		case tea.WindowSizeMsg:
			// m.width, m.height = msg.Width, msg.Height
			// m.updateListProperties()
			return m, nil
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
				statusCmd := kcl.NewStatusMessage("Pane " + fmt.Sprint(m.ActivePane))
				return m, tea.Batch(cmd, statusCmd)

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
				newItem := NewItemKafkaCluster("aaa", "bbb")
				insCmd := kcl.InsertItem(0, newItem)
				statusCmd := kcl.NewStatusMessage("Added " + newItem.Title() + ", pane " + fmt.Sprint(m.ActivePane))
				return m, tea.Batch(insCmd, statusCmd)
			}
		}

		// This will also call our delegate's update function.
		newListModel, cmd := kcl.Update(msg)
		kc.List = &newListModel
		cmds = append(cmds, cmd)
	} else if m.ActivePane != 0 {
		// slog.Info("XXXXXX", "pane", m.ActivePane)
		switch msg := msg.(type) {
		case spinner.TickMsg:
			newListModel, cmd := kc.List.Update(msg)
			kc.List = &newListModel
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}
	}

	return m, tea.Batch(cmds...)
}
