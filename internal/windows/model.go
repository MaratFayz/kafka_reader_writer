package windows

import (
	"fmt"
	"log/slog"
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
	selectedKafkaCluster int

	kafkaTopics        map[int]int
	selectedKafkaTopic int

	kafkaPartitions        map[int]int
	selectedKafkaPartition int

	activePane int //0-cluster,1-topic,2-partitions,3-readMessages,4-writeMessages

	kafkaClusterList    *components.KafkaClusterList
	kafkaTopicList      list.Model
	kafkaPartitionList  list.Model
	sendMessageTextArea textarea.Model
	sentMessagesList    list.Model
}

func (m *Model) SetActivePane(activePane int) {
	slog.Info("asd1", "as", fmt.Sprintf("%s %d %d", "SetActivePane", activePane, m.activePane))
	m.activePane = 1
	slog.Info("asd2", "as", fmt.Sprintf("%s %d %d", "SetActivePane", activePane, m.activePane))
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

	v2 := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, kcl.View(), kcl.View()))
	v2.AltScreen = true

	return v2
}
