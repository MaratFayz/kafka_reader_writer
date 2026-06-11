package windows

import (
	"marat/fayz/kafka_reader_writer/internal/localstorage"

	list "charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	kafkaClusters        []KafkaCluster
	selectedKafkaCluster string

	kafkaTopics        map[string]int
	selectedKafkaTopic string
	isLoadTopics       bool

	kafkaPartitions        map[int]int
	selectedKafkaPartition int

	activePane int //0-cluster,1-topic,2-partitions,3-readMessages,4-writeMessages

	kafkaClusterList    kafkaClusterList
	kafkaTopicList      kafkaTopicList
	kafkaPartitionList  list.Model
	sendMessageTextArea textarea.Model
	sentMessagesList    list.Model
}

type kafkaClusterList interface {
	update(msg tea.Msg, model *Model) (tea.Model, tea.Cmd)
}

type kafkaTopicList interface {
	update(msg tea.Msg, model *Model) (tea.Model, tea.Cmd)
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

type KafkaCluster interface {
	Title() string
	Url() string
	TrustStorePath() string
	TrustStorePassword() string
	Username() string
	Password() string
	SaslMechanism() string
}

func InitialModel(ls localstorage.LocalStorage) *Model {
	// ta := components.CreateTextArea()

	model := &Model{
		// sendMessageTextArea: ta,
		activePane: 0,
	}

	return model
}

func PostInitModel(model *Model, kafkaClusterList kafkaClusterList, kafkaTopicListComponent kafkaTopicList) *Model {
	model.kafkaClusterList = kafkaClusterList
	model.kafkaTopicList = kafkaTopicListComponent

	return model
}

func (m *Model) Init() tea.Cmd {
	//TODO add query to database
	// newItem := components.NewItem("sfa", "172.16.15.171:9093")
	// insCmd := kcl.InsertItem(0, newItem)
	// return insCmd
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmds []tea.Cmd

	_, cmd := m.kafkaClusterList.update(msg, m)
	_, cmd2 := m.kafkaTopicList.update(msg, m)

	return m, tea.Batch(cmd, cmd2)
}

func (m *Model) View() tea.View {
	kcl := m.kafkaClusterList.List()
	styles := m.kafkaClusterList.Styles()

	// Update list size.
	h, v := styles.App.GetFrameSize()
	kcl.SetSize(100-h, 100-v)

	// Update the model and list styles.
	kcl.Styles.Title = styles.Title

	ktl := m.kafkaTopicList.List()
	stylesKtl := m.kafkaTopicList.Styles()

	// Update list size.
	h2, v2 := stylesKtl.App.GetFrameSize()
	ktl.SetSize(100-h2, 100-v2)

	// Update the model and list styles.
	ktl.Styles.Title = stylesKtl.Title

	v3 := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, kcl.View(), ktl.View()))
	v3.AltScreen = true

	return v3
}
