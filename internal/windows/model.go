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
	IsLoadTopics       bool

	kafkaPartitions        map[int]int
	selectedKafkaPartition int

	ActivePane int //0-cluster,1-topic,2-partitions,3-readMessages,4-writeMessages

	kafkaClusterList    KafkaClusterList
	kafkaTopicList      KafkaTopicList
	kafkaPartitionList  list.Model
	sendMessageTextArea textarea.Model
	sentMessagesList    list.Model
}

type KafkaClusterList interface {
	Update(msg tea.Msg, model *Model) (tea.Model, tea.Cmd)
	GetList() *list.Model
	GetStyles() StylesKafkaCluster
}

type KafkaTopicList interface {
	Update(msg tea.Msg, model *Model) (tea.Model, tea.Cmd)
	GetList() *list.Model
	GetStyles() StylesKafkaCluster
}

type StylesKafkaCluster interface {
	GetApp() lipgloss.Style
	GetTitle() lipgloss.Style
	GetStatusMessage() lipgloss.Style
}

func (m *Model) SetActivePaneAfterKafkaClusterChosen(activePane int) {
	m.ActivePane = activePane
}

func (m *Model) SetChosenKafkaCluster(selectedKafkaCluster string) {
	m.selectedKafkaCluster = selectedKafkaCluster
}

func (m *Model) SetActivePaneAfterKafkaTopicChosen(activePane int) {
	m.ActivePane = activePane
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
		ActivePane: 0,
	}

	return model
}

func PostInitModel(model *Model, kafkaClusterList KafkaClusterList, kafkaTopicListComponent KafkaTopicList) *Model {
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

	// slog.Info("Active pane", "pane", m.ActivePane)

	_, cmd2 := m.kafkaTopicList.Update(msg, m)
	_, cmd := m.kafkaClusterList.Update(msg, m)

	return m, tea.Batch(cmd, cmd2)
}

func (m *Model) View() tea.View {
	kcl := m.kafkaClusterList.GetList()
	styles := m.kafkaClusterList.GetStyles()

	// Update list size.
	h, v := styles.GetApp().GetFrameSize()
	kcl.SetSize(100-h, 100-v)

	// Update the model and list styles.
	kcl.Styles.Title = styles.GetTitle()

	ktl := m.kafkaTopicList.GetList()
	stylesKtl := m.kafkaTopicList.GetStyles()

	// Update list size.
	h2, v2 := stylesKtl.GetApp().GetFrameSize()
	ktl.SetSize(100-h2, 100-v2)

	// Update the model and list styles.
	ktl.Styles.Title = stylesKtl.GetTitle()

	v3 := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, kcl.View(), ktl.View()))
	v3.AltScreen = true

	return v3
}
