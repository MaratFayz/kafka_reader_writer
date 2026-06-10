package windows

import (
	"fmt"
	"marat/fayz/kafka_reader_writer/internal/components"
	"marat/fayz/kafka_reader_writer/internal/localstorage"

	"charm.land/bubbles/v2/key"
	list "charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	kafkaClusters        []kafkaCluster
	selectedKafkaCluster string

	kafkaTopics        map[string]int
	selectedKafkaTopic string

	kafkaPartitions        map[int]int
	selectedKafkaPartition int

	activePane int //0-cluster,1-topic,2-partitions,3-readMessages,4-writeMessages

	kafkaClusterList    *components.KafkaClusterList
	kafkaTopicList      *components.KafkaTopicList
	kafkaPartitionList  list.Model
	sendMessageTextArea textarea.Model
	sentMessagesList    list.Model
}

func (m *Model) SetActivePaneAfterKafkaClusterChosen(activePane int) {
	m.activePane = activePane
}

func (m *Model) SetChosenKafkaCluster(selectedKafkaCluster string) {
	m.selectedKafkaCluster = selectedKafkaCluster
}

func (m *Model) SetActivePaneAfterKafkaTopicChosen(activePane int) {
	m.activePane = activePane
}

func (m *Model) SetChosenKafkaTopic(selectedKafkaTopic string) {
	m.selectedKafkaTopic = selectedKafkaTopic
}

type kafkaCluster interface {
	Title() string
	Url() string
	TrustStorePath() string
	TrustStorePassword() string
	Username() string
	Password() string
	SaslMechanism() string
}

func InitialModel(ls localstorage.LocalStorage) *Model {
	ta := components.CreateTextArea()

	model := &Model{
		sendMessageTextArea: ta,
		activePane:          0,
	}

	model.kafkaClusterList = createKafkaClustersListAddValues(ls, model)
	model.kafkaTopicList = createKafkaTopicListAddValues(ls, model)

	return model
}

func createKafkaClustersListAddValues(ls localstorage.LocalStorage, model *Model) *components.KafkaClusterList {
	kafkaClusterList := components.CreateKafkaClustersList(model)

	kc := ls.GetKafkaClusters()

	var namedUsers []kafkaCluster = make([]kafkaCluster, len(kc))
	for i, user := range kc {
		namedUsers[i] = user // Каждый элемент преобразуется отдельно
	}
	fmt.Println(namedUsers)
	for i, v := range namedUsers {
		kafkaClusterList.List.InsertItem(i, components.NewItemKafkaCluster(v.Title(), v.Url()))
	}

	return kafkaClusterList
}

func createKafkaTopicListAddValues(ls localstorage.LocalStorage, model *Model) *components.KafkaTopicList {
	kafkaTopicList := components.CreateKafkaTopicsList(model)

	// kc := ls.GetKafkaClusters()

	// var namedUsers []kafkaCluster = make([]kafkaCluster, len(kc))
	// for i, user := range kc {
	// namedUsers[i] = user // Каждый элемент преобразуется отдельно
	// }
	// fmt.Println(namedUsers)
	// for i, v := range namedUsers {
	// 	kafkaTopicList.List.InsertItem(i, components.NewItemKafkaCluster(v.Title(), v.Url()))
	// }

	return kafkaTopicList
}

func (m *Model) Init() tea.Cmd {
	//TODO add query to database
	// newItem := components.NewItem("sfa", "172.16.15.171:9093")
	// insCmd := kcl.InsertItem(0, newItem)
	// return insCmd
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.activePane == 0 {
		kcl := m.kafkaClusterList.List
		keys := m.kafkaClusterList.ListKeys
		delegateKeys := m.kafkaClusterList.DelegateKeys

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
				statusCmd := kcl.NewStatusMessage("Pane " + fmt.Sprint(m.activePane))
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
				newItem := components.NewItemKafkaCluster("aaa", "bbb")
				insCmd := kcl.InsertItem(0, newItem)
				statusCmd := kcl.NewStatusMessage("Added " + newItem.Title() + ", pane " + fmt.Sprint(m.activePane))
				return m, tea.Batch(insCmd, statusCmd)
			}
		}

		// This will also call our delegate's update function.
		newListModel, cmd := kcl.Update(msg)
		m.kafkaClusterList.List = &newListModel
		cmds = append(cmds, cmd)

		// m.activePane = 1
	} else if m.activePane == 1 {
		kcl := m.kafkaTopicList.List
		keys := m.kafkaTopicList.ListKeys
		delegateKeys := m.kafkaTopicList.DelegateKeys

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
				statusCmd := kcl.NewStatusMessage("Pane " + fmt.Sprint(m.activePane))
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
				newItem := components.NewItemKafkaTopic("aaa", "bbb")
				insCmd := kcl.InsertItem(0, newItem)
				statusCmd := kcl.NewStatusMessage("Added " + newItem.Title() + ", pane " + fmt.Sprint(m.activePane))
				return m, tea.Batch(insCmd, statusCmd)
			}
		}

		// This will also call our delegate's update function.
		newListModel, cmd := kcl.Update(msg)
		m.kafkaTopicList.List = &newListModel
		cmds = append(cmds, cmd)

		// m.activePane = 1
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() tea.View {
	kcl := m.kafkaClusterList.List
	styles := m.kafkaClusterList.Styles

	// Update list size.
	h, v := styles.App.GetFrameSize()
	kcl.SetSize(100-h, 100-v)

	// Update the model and list styles.
	kcl.Styles.Title = styles.Title

	ktl := m.kafkaTopicList.List
	stylesKtl := m.kafkaTopicList.Styles

	// Update list size.
	h2, v2 := stylesKtl.App.GetFrameSize()
	ktl.SetSize(100-h2, 100-v2)

	// Update the model and list styles.
	ktl.Styles.Title = stylesKtl.Title

	v3 := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, kcl.View(), ktl.View()))
	v3.AltScreen = true

	return v3
}
