package components

import (
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

type ModelChangerKafkaCluster interface {
	SetActivePane(activePane int)
}

type ufKafkaCluster struct {
	model  ModelChangerKafkaCluster
	keys   *DelegateKeyMapKafkaCluster
	styles *StylesKafkaCluster
}

func (u *ufKafkaCluster) updateFunc(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string

	// if i, ok := m.SelectedItem().(Item); ok {
	// 	title = i.Title()
	// } else {
	// 	return nil
	// }

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, u.keys.Choose):
			u.model.SetActivePane(1)
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
