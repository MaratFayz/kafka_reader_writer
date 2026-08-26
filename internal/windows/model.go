package windows

import (
	"bytes"
	"fmt"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
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

	stackOfOverlayWindows                    overlayv2.Stack
	isWindowForDetailedReadKafkaMessageShown bool

	overlayWindowProvider OverlayWindowForKafkaReadDetailedMessageProvider

	isHelpWindowNeedToOpen bool
	helpWindow             HelpWindowProvider

	backgroundColorMsg string
	height             int
	width              int

	errorWindow ErrorWindowProvider

	localStorage LocalStorage

	deleteClusterWindow    OverlayWindowForDeleteNewServerProvider
	appendNewClusterWindow OverlayWindowForAppendNewServerProvider
	editClusterWindow      OverlayWindowForEditClusterProvider
}

type activePane int

type LocalStorage interface {
	GetKafkaClusters() ([]*contracts.KafkaCluster, error)
}

const (
	CLUSTERS activePane = iota
	TOPICS
	PARTITIONS
	TABS_READ_WRITE
	TABS_READ_WRITE_WINDOW_DETAILED_MESSAGE
	CLUSTERS_HELP_WINDOW
	TOPICS_HELP_WINDOW
	PARTITIONS_HELP_WINDOW
	TABS_READ_WRITE_HELP_WINDOW
	CLUSTERS_ERROR_WINDOW
	TOPICS_ERROR_WINDOW
	PARTITIONS_ERROR_WINDOW
	TABS_READ_WRITE_ERROR_WINDOW
	CLUSTERS_APPEND_CLUSTER_WINDOW
	CLUSTERS_DELETE_CLUSTER_WINDOW
	CLUSTERS_EDIT_CLUSTER_WINDOW
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

type OverlayWindowForKafkaReadDetailedMessageProvider interface {
	ProvideNewWindowForKafkaReadDetailedMessage(body string, header string, offset string, datetime string) tea.Model
}

type HelpWindowProvider interface {
	ProvideNewOverlayWindowForHelp(values []contracts.KeyValue) tea.Model
}

type ErrorWindowProvider interface {
	ProvideNewOverlayWindowForErrorMessagesProvider(text string) tea.Model
}

type OverlayWindowForDeleteNewServerProvider interface {
	ProvideNewOverlayWindowForDeleteNewServer() tea.Model
}

type OverlayWindowForAppendNewServerProvider interface {
	ProvideNew(body string, key string, offset string, datetime string) tea.Model
}

type OverlayWindowForEditClusterProvider interface {
	ProvideNewOverlayWindowForEditCluster(body string, key string, offset string, datetime string) tea.Model
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

func InitialModel(ls LocalStorage) *Model {
	model := &Model{
		activePane:       CLUSTERS,
		kafkaClusters:    make(map[string]*contracts.KafkaCluster),
		kafkaTopics:      make(map[string][]string),
		kafkaPartitions:  make(map[string]map[string][]int),
		IsLoadTopics:     make(map[string]bool),
		IsLoadPartitions: make(map[string]map[string]bool),
		localStorage:     ls,
	}

	return model
}

func PostInitModel(model *Model, kafkaClusterList KafkaClusterList, kafkaTopicListComponent KafkaTopicList,
	kafkaPartitionListComponent KafkaPartitionList, kafkaReadWriteTabsComponent KafkaReadWriteTabsComponent,
	overlayWindowProvider OverlayWindowForKafkaReadDetailedMessageProvider, helpWindow HelpWindowProvider,
	errorWindow ErrorWindowProvider, deleteClusterWindow OverlayWindowForDeleteNewServerProvider,
	appendNewClusterWindow OverlayWindowForAppendNewServerProvider,
	editClusterWindow OverlayWindowForEditClusterProvider) *Model {
	model.kafkaClusterList = kafkaClusterList
	model.kafkaTopicList = kafkaTopicListComponent
	model.kafkaPartitionList = kafkaPartitionListComponent
	model.kafkaReadWriteTabs = kafkaReadWriteTabsComponent
	model.overlayWindowProvider = overlayWindowProvider
	model.helpWindow = helpWindow
	model.errorWindow = errorWindow
	model.deleteClusterWindow = deleteClusterWindow
	model.appendNewClusterWindow = appendNewClusterWindow
	model.editClusterWindow = editClusterWindow

	return model
}

func (m *Model) Init() tea.Cmd {
	slog.Info("Init in main model")
	return tea.Batch(tea.RequestBackgroundColor, func() tea.Msg {
		slog.Info("Init in main model: получение кластеров из локального хранилища")

		kc, err := m.localStorage.GetKafkaClusters()

		if err != nil {
			slog.Error("Init in main model: ошибка при получении списка кластеров из локального хранилища")
			return contracts.ErrorClusterLoadedEvent{Err: err}
		}

		clusterList := make([]*contracts.KafkaCluster, len(kc))
		clusterMap := make(map[string]*contracts.KafkaCluster, len(kc))
		for i, cluster := range kc {
			clusterList[i] = cluster
			clusterMap[cluster.Title] = cluster
		}

		return contracts.SuccessClusterLoadedEvent{KafkaClusters: clusterMap, KafkaClustersList: clusterList}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case contracts.SuccessClusterLoadedEvent:
		slog.Info("In main model: event contracts.ClusterLoadedEvent")
		m.kafkaClusters = msg.KafkaClusters
		return m, func() tea.Msg {
			return contracts.UpdateListOfClustersEvent{Clusters: msg.KafkaClustersList}
		}
	case contracts.ErrorClusterLoadedEvent:
		slog.Info("In main model: event contracts.ErrorClusterLoadedEvent")

		switch m.activePane {
		case CLUSTERS:
			m.activePane = CLUSTERS_ERROR_WINDOW
		case TOPICS:
			m.activePane = TOPICS_ERROR_WINDOW
		case PARTITIONS:
			m.activePane = PARTITIONS_ERROR_WINDOW
		case TABS_READ_WRITE:
			m.activePane = TABS_READ_WRITE_ERROR_WINDOW
		}

		return m, m.stackOfOverlayWindows.Push(m.errorWindow.ProvideNewOverlayWindowForErrorMessagesProvider(msg.Err.Error()),
			overlay.OverlayConfig{
				Placement:           overlay.Center(),
				DimOpacity:          0.35,
				CloseOnEscape:       false,
				CloseOnClickOutside: false,
			})
	case contracts.CloseWindowForErrorMessages:
		slog.Info("in main: event contracts.CloseWindowForErrorMessages, depth before = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

		var c tea.Cmd

		if m.stackOfOverlayWindows.Depth() > 0 {
			_, c = m.stackOfOverlayWindows.Pop()
		}

		if m.stackOfOverlayWindows.Depth() == 0 {
			slog.Info("in main: event contracts.CloseWindowForErrorMessages, depth ZERO = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

			switch m.activePane {
			case CLUSTERS_ERROR_WINDOW:
				m.activePane = CLUSTERS
			case TOPICS_ERROR_WINDOW:
				m.activePane = TOPICS
			case PARTITIONS_ERROR_WINDOW:
				m.activePane = PARTITIONS
			case TABS_READ_WRITE_ERROR_WINDOW:
				m.activePane = TABS_READ_WRITE
			}
		}

		slog.Info("in main: event contracts.CloseWindowForErrorMessages, depth after = " +
			strconv.Itoa(m.stackOfOverlayWindows.Depth()) + ", m.activePane = " + strconv.Itoa(int(m.activePane)))

		return m, c
	case tea.BackgroundColorMsg:
		m.backgroundColorMsg = msg.String()
		slog.Info("in main: event BackgroundColorMsg = " + msg.String())
		return m, nil
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		slog.Info("in main: event WindowSizeMsg; h = " + strconv.Itoa(int(msg.Height)) + ", w = " + strconv.Itoa(int(msg.Width)))
		return m, nil
	case contracts.UserChosenKafkaCluster:
		m.activePane = TOPICS
		m.SelectedKafkaCluster = msg.ChosenCluster
		slog.Info("in main: event contracts.UserChosenKafkaCluster: activePane = " + strconv.Itoa(int(m.activePane)) +
			", m.SelectedKafkaCluster = " + m.SelectedKafkaCluster)

		return m, func() tea.Msg {
			return initList{}
		}
	case contracts.UserWantsDeleteClusterFromList:
		slog.Info("in main: event contracts.UserWantsDeleteClusterFromList")
		m.activePane = CLUSTERS_DELETE_CLUSTER_WINDOW

		return m, m.stackOfOverlayWindows.Push(m.deleteClusterWindow.ProvideNewOverlayWindowForDeleteNewServer(), overlay.OverlayConfig{
			Placement:           overlay.Center(),
			DimOpacity:          0.35,
			CloseOnEscape:       false,
			CloseOnClickOutside: false,
		})
	case contracts.CloseWindowForDeleteNewServer:
		slog.Info("in main: event contracts.CloseWindowForDeleteNewServer, depth before = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

		var c tea.Cmd

		if m.stackOfOverlayWindows.Depth() > 0 {
			_, c = m.stackOfOverlayWindows.Pop()
		}

		if m.stackOfOverlayWindows.Depth() == 0 {
			slog.Info("in main: event contracts.CloseWindowForDeleteNewServer, depth ZERO = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))
			m.activePane = CLUSTERS
		}

		slog.Info("in main: event contracts.CloseWindowForDeleteNewServer, depth after = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()) + ", m.activePane = " + strconv.Itoa(int(m.activePane)))

		return m, c
	case contracts.UserWantsEditClusterInList:
		slog.Info("in main: event contracts.UserWantsEditClusterInList")
	case contracts.CloseWindowForEditCluster:
		slog.Info("in main: event contracts.CloseWindowForEditCluster, depth before = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

		var c tea.Cmd

		if m.stackOfOverlayWindows.Depth() > 0 {
			_, c = m.stackOfOverlayWindows.Pop()
		}

		if m.stackOfOverlayWindows.Depth() == 0 {
			slog.Info("in main: event contracts.CloseWindowForEditCluster, depth ZERO = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))
			m.activePane = TABS_READ_WRITE
		}

		slog.Info("in main: event contracts.CloseWindowForEditCluster, depth after = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()) + ", m.activePane = " + strconv.Itoa(int(m.activePane)))

		return m, c
	case contracts.UserWantsAddClusterToList:
		slog.Info("in main: event contracts.UserWantsAddClusterToList")
	case contracts.CloseWindowForAppendNewCluster:
		slog.Info("in main: event contracts.CloseWindowForAppendNewCluster, depth before = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

		var c tea.Cmd

		if m.stackOfOverlayWindows.Depth() > 0 {
			_, c = m.stackOfOverlayWindows.Pop()
		}

		if m.stackOfOverlayWindows.Depth() == 0 {
			slog.Info("in main: event contracts.CloseWindowForAppendNewCluster, depth ZERO = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))
			m.activePane = TABS_READ_WRITE
		}

		slog.Info("in main: event contracts.CloseWindowForAppendNewCluster, depth after = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()) + ", m.activePane = " + strconv.Itoa(int(m.activePane)))

		return m, c
	case contracts.CloseWindowForDetailedReadMessage:
		slog.Info("in main: event contracts.CloseOverlayWindow, depth before = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

		var c tea.Cmd

		if m.stackOfOverlayWindows.Depth() > 0 {
			_, c = m.stackOfOverlayWindows.Pop()
		}

		if m.stackOfOverlayWindows.Depth() == 0 {
			slog.Info("in main: event contracts.CloseOverlayWindow, depth ZERO = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))
			m.activePane = TABS_READ_WRITE
		}

		slog.Info("in main: event contracts.CloseOverlayWindow, depth after = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()) + ", m.activePane = " + strconv.Itoa(int(m.activePane)))

		return m, c
	case tea.KeyPressMsg:
		slog.Info("in main: нажата кнопка " + msg.Keystroke())

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
			case TABS_READ_WRITE_WINDOW_DETAILED_MESSAGE:
			}
		case "?":
			m.isHelpWindowNeedToOpen = true
			slog.Info("in main: выпущено событие CreateAskForHelpEvent")
			return m, contracts.CreateAskForHelpEvent()
		}
	case contracts.HelpMessage:
		if m.isHelpWindowNeedToOpen == false {
			return m, nil
		}

		switch m.activePane {
		case CLUSTERS:
			m.activePane = CLUSTERS_HELP_WINDOW
		case TOPICS:
			m.activePane = TOPICS_HELP_WINDOW
		case PARTITIONS:
			m.activePane = PARTITIONS_HELP_WINDOW
		case TABS_READ_WRITE:
			m.activePane = TABS_READ_WRITE_HELP_WINDOW
		}

		slog.Info("in main: event contracts.HelpMessage, value = " + ", создание окна помощи, активная панель теперь " +
			strconv.Itoa(int(m.activePane)))

		return m, m.stackOfOverlayWindows.Push(m.helpWindow.ProvideNewOverlayWindowForHelp(msg.Values), overlay.OverlayConfig{
			Placement:           overlay.Center(),
			DimOpacity:          0.35,
			CloseOnEscape:       false,
			CloseOnClickOutside: false,
		})
	case contracts.CloseWindowForHelp:
		m.isHelpWindowNeedToOpen = false
		slog.Info("in main: event contracts.CloseWindowForHelp, depth before = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

		var c tea.Cmd

		if m.stackOfOverlayWindows.Depth() > 0 {
			_, c = m.stackOfOverlayWindows.Pop()
		}

		if m.stackOfOverlayWindows.Depth() == 0 {
			slog.Info("in main: event contracts.CloseWindowForHelp, depth ZERO = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()))

			switch m.activePane {
			case CLUSTERS_HELP_WINDOW:
				m.activePane = CLUSTERS
			case TOPICS_HELP_WINDOW:
				m.activePane = TOPICS
			case PARTITIONS_HELP_WINDOW:
				m.activePane = PARTITIONS
			case TABS_READ_WRITE_HELP_WINDOW:
				m.activePane = TABS_READ_WRITE
			}
		}

		slog.Info("in main: event contracts.CloseWindowForHelp, depth after = " + strconv.Itoa(m.stackOfOverlayWindows.Depth()) + ", m.activePane = " + strconv.Itoa(int(m.activePane)))

		return m, c
	case contracts.ReadMessagesTableChosenRow:
		slog.Info("in main: event contracts.ReadMessagesTableChosenRow")
		m.activePane = TABS_READ_WRITE_WINDOW_DETAILED_MESSAGE
		return m, m.stackOfOverlayWindows.Push(m.overlayWindowProvider.ProvideNewWindowForKafkaReadDetailedMessage(msg.Body, msg.Header, msg.Offset, msg.Datetime), overlay.OverlayConfig{
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
		slog.Info("in main: TABS_READ_WRITE")
		_, cmd := m.kafkaReadWriteTabs.Update(msg, m)
		return m, tea.Batch(cmd)
	case TABS_READ_WRITE_WINDOW_DETAILED_MESSAGE, CLUSTERS_HELP_WINDOW, TOPICS_HELP_WINDOW, PARTITIONS_HELP_WINDOW,
		CLUSTERS_ERROR_WINDOW, TOPICS_ERROR_WINDOW, PARTITIONS_ERROR_WINDOW, TABS_READ_WRITE_ERROR_WINDOW, CLUSTERS_DELETE_CLUSTER_WINDOW:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "in main: активная панель %v: event %v", m.activePane, msg)
		slog.Info(buf.String())

		cmd := m.stackOfOverlayWindows.Update(msg)
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
	kcl.SetSize(m.width/4, m.height)
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
	ktl.SetSize(m.width/4, m.height)

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
	kpl.SetSize(m.width/4, m.height)

	// Update the model and list styles.
	kpl.Styles.Title = stylesKpl.GetTitle()

	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	kclString := kcl.View()
	ktlString := ktl.View()
	kplString := kpl.View()
	kafkaReadWriteTabsString := m.kafkaReadWriteTabs.View()

	switch m.activePane {
	case CLUSTERS, CLUSTERS_HELP_WINDOW:
		kclString = s.Render(kclString)
	case TOPICS, TOPICS_HELP_WINDOW:
		ktlString = s.Render(ktlString)
	case PARTITIONS, PARTITIONS_HELP_WINDOW:
		kplString = s.Render(kplString)
	case TABS_READ_WRITE, TABS_READ_WRITE_WINDOW_DETAILED_MESSAGE, TABS_READ_WRITE_HELP_WINDOW:
		kafkaReadWriteTabsString = s.Render(kafkaReadWriteTabsString)
	}

	v3 := lipgloss.JoinHorizontal(lipgloss.Left, kclString, ktlString, kplString, kafkaReadWriteTabsString)

	if m.stackOfOverlayWindows.Depth() > 0 {
		v4 := m.stackOfOverlayWindows.CompositeView(v3, m.width, m.height)
		v4.AltScreen = true
		return v4
	} else {
		vv := tea.NewView(v3)
		vv.AltScreen = true

		return vv
	}
}
