package components

import (
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

type ModelChangerKafkaTopic interface {
	SetActivePane(activePane int)
}

type ufKafkaTopic struct {
	model  ModelChangerKafkaTopic
	keys   *DelegateKeyMapKafkaTopic
	styles *StylesKafkaTopic
}

func (u *ufKafkaTopic) updateFuncKafkaTopic(msg tea.Msg, m *list.Model) tea.Cmd {
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
			u.model.SetActivePane(2)
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

func CreateKafkaTopicsList(model ModelChangerKafkaTopic) *KafkaTopicList {
	styles := newStylesKafkaTopic(false) // default to dark background styles

	delegateKeys := NewDelegateKeyMapKafkaTopic()
	listKeys := newListKeyMapKafkaTopic()

	items := make([]list.Item, 0)

	delegate := newItemDelegateKafkaTopic(delegateKeys, &styles, model)
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

	return &KafkaTopicList{&list, delegateKeys, listKeys, &styles}
}
