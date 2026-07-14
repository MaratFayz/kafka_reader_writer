package windows

import (
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"sync"
	"sync/atomic"

	list "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	kafkaClusters   map[string]*contracts.KafkaCluster
	muKafkaClusters sync.Mutex

	SelectedKafkaCluster string

	kafkaTopics   map[string][]string
	muKafkaTopics sync.Mutex

	SelectedKafkaTopic string
	IsLoadTopics       atomic.Bool

	kafkaPartitions   map[string]map[string][]int
	muKafkaPartitions sync.Mutex

	SelectedKafkaPartition string
	IsLoadPartitions       bool
	ActivePane             int //0-cluster,1-topic,2-partitions,3-tabs

	kafkaClusterList   KafkaClusterList
	kafkaTopicList     KafkaTopicList
	kafkaPartitionList KafkaPartitionList
	kafkaReadWriteTabs KafkaReadWriteTabsComponent
}

func (m *Model) GetClusterByTitle(title string) *contracts.KafkaCluster {
	m.muKafkaClusters.Lock()
	defer m.muKafkaClusters.Unlock()

	return m.kafkaClusters[title]
}

func (m *Model) SetClusters(clusters map[string]*contracts.KafkaCluster) {
	m.muKafkaClusters.Lock()
	defer m.muKafkaClusters.Unlock()

	m.kafkaClusters = clusters
}

func (m *Model) GetKafkaTopics(title string) []string {
	m.muKafkaTopics.Lock()
	defer m.muKafkaTopics.Unlock()

	return m.kafkaTopics[title]
}

func (m *Model) SetKafkaTopics(title string, topics []string) {
	m.muKafkaTopics.Lock()
	defer m.muKafkaTopics.Unlock()

	m.kafkaTopics[title] = topics
}

func (m *Model) GetKafkaPartitions(title string, topic string) []int {
	m.muKafkaPartitions.Lock()
	defer m.muKafkaPartitions.Unlock()

	return m.kafkaPartitions[title][topic]
}

type KafkaReadWriteTabsComponent interface {
	Update(msg tea.Msg, model *Model) (tea.Model, tea.Cmd)
	View() string
	Init() tea.Cmd
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

type KafkaPartitionList interface {
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
	m.SelectedKafkaCluster = selectedKafkaCluster
}

func (m *Model) SetActivePaneAfterKafkaTopicChosen(activePane int) {
	m.ActivePane = activePane
}

func (m *Model) SetChosenKafkaTopic(selectedKafkaTopic string) {
	m.SelectedKafkaTopic = selectedKafkaTopic
}

func (m *Model) SetTopicsForCluster(kafkaCluster string, topics []string) {
	m.SetKafkaTopics(kafkaCluster, topics)
}

func (m *Model) SetActivePaneAfterKafkaPartitionChosen(activePane int) {
	m.ActivePane = activePane
}

func (m *Model) SetChosenKafkaPartition(selectedKafkaPartition string) {
	m.SelectedKafkaPartition = selectedKafkaPartition
}

func (m *Model) SetPartitionsForClusterAndTopic(clusterName string, topicName string, partitions []int) {
	m.muKafkaClusters.Lock()
	m.muKafkaPartitions.Lock()
	m.muKafkaTopics.Lock()
	defer m.muKafkaClusters.Unlock()
	defer m.muKafkaPartitions.Unlock()
	defer m.muKafkaTopics.Unlock()

	if m.kafkaPartitions == nil {
		m.kafkaPartitions = make(map[string]map[string][]int)
	}

	c, ok := m.kafkaPartitions[clusterName]
	if !ok {
		m.kafkaPartitions[clusterName] = make(map[string][]int)
	}
	c = m.kafkaPartitions[clusterName]

	t, ok := c[topicName]
	if !ok {
		c[topicName] = make([]int, 0)
	}
	t = c[topicName]

	c[topicName] = append(t, partitions...)

	// slog.Info("A", "a", m.ActivePane)
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

type LocalStorage interface {
	GetKafkaClusters() []*contracts.KafkaCluster
}

func InitialModel(ls LocalStorage) *Model {
	model := &Model{
		ActivePane:       0,
		kafkaClusters:    make(map[string]*contracts.KafkaCluster),
		kafkaTopics:      make(map[string][]string),
		kafkaPartitions:  make(map[string]map[string][]int),
		IsLoadTopics:     atomic.Bool{},
		IsLoadPartitions: false,
	}

	return model
}

func PostInitModel(model *Model, kafkaClusterList KafkaClusterList, kafkaTopicListComponent KafkaTopicList,
	kafkaPartitionListComponent KafkaPartitionList, kafkaReadWriteTabsComponent KafkaReadWriteTabsComponent) *Model {
	model.kafkaClusterList = kafkaClusterList
	model.kafkaTopicList = kafkaTopicListComponent
	model.kafkaPartitionList = kafkaPartitionListComponent
	model.kafkaReadWriteTabs = kafkaReadWriteTabsComponent

	return model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd4 := m.kafkaReadWriteTabs.Update(msg, m)
	_, cmd3 := m.kafkaPartitionList.Update(msg, m)
	_, cmd2 := m.kafkaTopicList.Update(msg, m)
	_, cmd := m.kafkaClusterList.Update(msg, m)

	return m, tea.Batch(cmd, cmd2, cmd3, cmd4)
}

func (m *Model) View() tea.View {
	kcl := m.kafkaClusterList.GetList()
	styles := m.kafkaClusterList.GetStyles()

	// Update list size.
	// h, _ := styles.GetApp().GetFrameSize()
	// kcl.SetSize(100-h, 30)
	kcl.SetSize(50, 30)
	// h, v := styles.GetApp().GetFrameSize()
	// kcl.SetSize(100-h, 100-v)

	// Update the model and list styles.
	kcl.Styles.Title = styles.GetTitle()

	ktl := m.kafkaTopicList.GetList()
	stylesKtl := m.kafkaTopicList.GetStyles()

	// Update list size.
	// h2, v2 := stylesKtl.GetApp().GetFrameSize()
	// ktl.SetSize(100-h2, 100-v2)

	// h2, _ := stylesKtl.GetApp().GetFrameSize()
	// slog.Info("V", "Kafkfa Topic", v2, "v1", v)
	// ktl.SetSize(100-h2, 30)
	ktl.SetSize(50, 30)

	// Update the model and list styles.
	ktl.Styles.Title = stylesKtl.GetTitle()

	kpl := m.kafkaPartitionList.GetList()
	stylesKpl := m.kafkaPartitionList.GetStyles()

	// Update list size.
	// h2, v2 := stylesKtl.GetApp().GetFrameSize()
	// ktl.SetSize(100-h2, 100-v2)

	// h3, _ := stylesKpl.GetApp().GetFrameSize()
	// slog.Info("V", "Kafkfa Topic", v2, "v1", v)
	// kpl.SetSize(100-h3, 30)
	kpl.SetSize(50, 30)

	// Update the model and list styles.
	kpl.Styles.Title = stylesKpl.GetTitle()

	v3 := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, kcl.View(), ktl.View(), kpl.View(), m.kafkaReadWriteTabs.View()))
	v3.AltScreen = true

	return v3
}
