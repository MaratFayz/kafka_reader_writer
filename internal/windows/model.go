package windows

import (
	"bytes"
	"fmt"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"marat/fayz/kafka_reader_writer/internal/debug"
	"strconv"
	"sync"

	list "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	overlay "github.com/madicen/bubble-overlay"
	overlayv2 "github.com/madicen/bubble-overlay/v2"
)

type Model struct {
	kafkaClusters   map[string]*contracts.KafkaCluster
	muKafkaClusters sync.Mutex

	SelectedKafkaCluster string

	kafkaTopics   map[string][]string
	muKafkaTopics sync.Mutex

	SelectedKafkaTopic string
	IsLoadTopics       map[string]bool

	kafkaPartitions   map[string]map[string][]int
	muKafkaPartitions sync.Mutex

	SelectedKafkaPartition string
	IsLoadPartitions       map[string]map[string]bool
	activePane             activePane

	kafkaClusterList   KafkaClusterList
	kafkaTopicList     KafkaTopicList
	kafkaPartitionList KafkaPartitionList
	kafkaReadWriteTabs KafkaReadWriteTabsComponent

	stack               overlayv2.Stack
	isOverlayWindowShow bool

	overlayWindowProvider OverlayWindowProvider
}

type activePane int

const (
	CLUSTERS activePane = iota
	TOPICS
	PARTITIONS
	TABS_READ_WRITE
	TABS_READ_WRITE_OVERLAY_WINDOW
)

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
	IsFocusOnTabs() bool
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

type OverlayWindowProvider interface {
	ProvideNew(body string, header string, offset string, datetime string) tea.Model
}

func (m *Model) KafkaClusterChosenNextActivePane() tea.Cmd {
	m.activePane = TOPICS
	return func() tea.Msg {
		return initList{}
	}
}

func (m *Model) SetChosenKafkaCluster(selectedKafkaCluster string) {
	m.SelectedKafkaCluster = selectedKafkaCluster
}

func (m *Model) KafkaTopicChosenNextActivePane() tea.Cmd {
	m.activePane = PARTITIONS
	return func() tea.Msg {
		return initList{}
	}
}

type initList struct {
}

func (i initList) IsInitList() bool {
	return true
}

func (m *Model) SetChosenKafkaTopic(selectedKafkaTopic string) {
	m.SelectedKafkaTopic = selectedKafkaTopic
}

func (m *Model) SetTopicsForCluster(kafkaCluster string, topics []string) {
	m.SetKafkaTopics(kafkaCluster, topics)
}

func (m *Model) KafkaPartitionChosenNextActivePane() tea.Cmd {
	m.activePane = TABS_READ_WRITE
	return func() tea.Msg {
		return initList{}
	}
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
}

type LocalStorage interface {
	GetKafkaClusters() []*contracts.KafkaCluster
}

func InitialModel(ls LocalStorage) *Model {
	model := &Model{
		activePane:       CLUSTERS,
		kafkaClusters:    make(map[string]*contracts.KafkaCluster),
		kafkaTopics:      make(map[string][]string),
		kafkaPartitions:  make(map[string]map[string][]int),
		IsLoadTopics:     make(map[string]bool),
		IsLoadPartitions: make(map[string]map[string]bool),
	}

	return model
}

func PostInitModel(model *Model, kafkaClusterList KafkaClusterList, kafkaTopicListComponent KafkaTopicList,
	kafkaPartitionListComponent KafkaPartitionList, kafkaReadWriteTabsComponent KafkaReadWriteTabsComponent, overlayWindowProvider OverlayWindowProvider) *Model {
	model.kafkaClusterList = kafkaClusterList
	model.kafkaTopicList = kafkaTopicListComponent
	model.kafkaPartitionList = kafkaPartitionListComponent
	model.kafkaReadWriteTabs = kafkaReadWriteTabsComponent
	model.overlayWindowProvider = overlayWindowProvider

	return model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case contracts.CloseOverlayWindow:
		slog.Error("B", "m", "CloseOverlayWindow", "depth", m.stack.Depth())
		debug.WriteDebugFile("in main: event contracts.CloseOverlayWindow, depth before = " + strconv.Itoa(m.stack.Depth()))

		var c tea.Cmd

		if m.stack.Depth() > 0 {
			_, c = m.stack.Pop()
		}

		if m.stack.Depth() == 0 {
			debug.WriteDebugFile("in main: event contracts.CloseOverlayWindow, depth ZERO = " + strconv.Itoa(m.stack.Depth()))
			m.activePane = TABS_READ_WRITE
		}

		debug.WriteDebugFile("in main: event contracts.CloseOverlayWindow, depth after = " + strconv.Itoa(m.stack.Depth()) + ", m.activePane = " + strconv.Itoa(int(m.activePane)))

		return m, c
	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "esc":
			switch m.activePane {
			case TOPICS:
				m.activePane = CLUSTERS
				return m, nil
			case PARTITIONS:
				m.activePane = TOPICS
				return m, nil
			case TABS_READ_WRITE:
				if m.kafkaReadWriteTabs.IsFocusOnTabs() {
					m.activePane = PARTITIONS
					return m, nil
				}
			case TABS_READ_WRITE_OVERLAY_WINDOW:
			}
		}

	case contracts.ReadMessagesTableChosenRow:
		debug.WriteDebugFile("in main: event contracts.ReadMessagesTableChosenRow")
		m.activePane = TABS_READ_WRITE_OVERLAY_WINDOW
		return m, m.stack.Push(m.overlayWindowProvider.ProvideNew(msg.Body, msg.Header, msg.Offset, msg.Datetime), overlay.OverlayConfig{
			Placement:           overlay.Center(),
			DimOpacity:          0.35,
			CloseOnEscape:       false,
			CloseOnClickOutside: false,
		})
	}

	switch m.activePane {
	case CLUSTERS:
		_, cmd := m.kafkaClusterList.Update(msg, m)
		return m, tea.Batch(cmd)
	case TOPICS:
		_, cmd := m.kafkaTopicList.Update(msg, m)
		return m, tea.Batch(cmd)
	case PARTITIONS:
		_, cmd := m.kafkaPartitionList.Update(msg, m)
		return m, tea.Batch(cmd)
	case TABS_READ_WRITE:
		debug.WriteDebugFile("in main: TABS_READ_WRITE")
		_, cmd := m.kafkaReadWriteTabs.Update(msg, m)
		return m, tea.Batch(cmd)
	case TABS_READ_WRITE_OVERLAY_WINDOW:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "in main: TABS_READ_WRITE_OVERLAY_WINDOW: event %v", msg)
		debug.WriteDebugFile(buf.String())

		cmd := m.stack.Update(msg)
		return m, tea.Batch(cmd)
	}

	return m, nil
}

func (m *Model) View() tea.View {
	kcl := m.kafkaClusterList.GetList()
	styles := m.kafkaClusterList.GetStyles()

	// Update list size.
	// h, _ := styles.GetApp().GetFrameSize()
	// kcl.SetSize(100-h, 30)
	kcl.SetSize(30, 30)
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

	v3 := lipgloss.JoinHorizontal(lipgloss.Left, kcl.View(), ktl.View(), kpl.View(), m.kafkaReadWriteTabs.View())

	if m.stack.Depth() > 0 {
		v4 := m.stack.CompositeView(v3, 120, 30)
		v4.AltScreen = true
		return v4
	} else {
		vv := tea.NewView(v3)
		vv.AltScreen = true

		return vv
	}
}
