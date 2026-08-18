package app

import (
	"context"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/closer"
	"marat/fayz/kafka_reader_writer/internal/components"
	readwritetabs "marat/fayz/kafka_reader_writer/internal/components/readWriteTabs"
	"marat/fayz/kafka_reader_writer/internal/config"
	"marat/fayz/kafka_reader_writer/internal/database"
	"marat/fayz/kafka_reader_writer/internal/kafka/connection"
	"marat/fayz/kafka_reader_writer/internal/localstorage"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"os"
	"sync"
)

type diContainer struct {
	appConfig                                  *config.Config
	db                                         *database.Db
	localStorage                               *localstorage.LocalStorage
	model                                      *windows.Model
	fullModel                                  *windows.Model
	kafkaClusterList                           *components.KafkaClusterList
	kafkaTopicList                             *components.KafkaTopicList
	kafkaConnector                             *connection.KafkaConnectorProvider
	kafkaPartitionList                         *components.KafkaPartitionList
	kafkaReadWriteTabs                         *readwritetabs.KafkaReadWriteTabsComponent
	kafkaSendMessageTextArea                   *readwritetabs.KafkaSendMessageTextAreaComponent
	kafkaSendMessageTableComponent             *readwritetabs.KafkaSendMessageTableComponent
	kafkaReadMessagesTableComponent            *readwritetabs.KafkaReadMessagesTableComponent
	overlayForKafkaReadDetailedMessageProvider *components.OverlayWindowForKafkaReadDetailedMessageProvider
	spinnerWhenLoadMessagesFromKafkaComponent  *readwritetabs.SpinnerWhenLoadMessagesFromKafkaComponent
	overlayWindowForAppendNewServerProvider    *components.OverlayWindowForAppendNewServerProvider
	overlayWindowForHelpProvider               *components.OverlayWindowForHelpProvider
	overlayWindowForDeleteNewServerProvider    *components.OverlayWindowForDeleteNewServerProvider

	muAppConfig                                  sync.Mutex
	muDb                                         sync.Mutex
	muLocalStorage                               sync.Mutex
	muModel                                      sync.Mutex
	muFullModel                                  sync.Mutex
	muKafkaClusterList                           sync.Mutex
	muKafkaTopicList                             sync.Mutex
	muKafkaConnector                             sync.Mutex
	muKafkaPartitionList                         sync.Mutex
	muKafkaReadWriteTabs                         sync.Mutex
	muKafkaSendMessageTextArea                   sync.Mutex
	muKafkaReadMessagesTable                     sync.Mutex
	muOverlayForKafkaReadDetailedMessageProvider sync.Mutex
	muSpinnerWhenLoadMessagesFromKafkaComponent  sync.Mutex
	muOverlayWindowForAppendNewServerProvider    sync.Mutex
	muOverlayWindowForHelpProvider               sync.Mutex
	muOverlayWindowForDeleteNewServerProvider    sync.Mutex
}

func newDIContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) AppConfig() *config.Config {
	if d.appConfig == nil {
		d.muAppConfig.Lock()
		defer d.muAppConfig.Unlock()

		if d.appConfig != nil {
			return d.appConfig
		}

		d.appConfig = config.AppConfig()
	}

	return d.appConfig
}

func (d *diContainer) DB() *database.Db {
	if d.db == nil {
		d.muDb.Lock()
		defer d.muDb.Unlock()

		if d.db != nil {
			return d.db
		}

		slog.Info("Init db", "db", d.db)
		db, err := database.CreateOrGetDb(d.AppConfig().DatabaseUrl)
		if err != nil {
			slog.Error("не удалось подключиться к БД", "err", err)
			os.Exit(1)
		}

		closer.Add("база данных", func(_ context.Context) error {
			return db.Close()
		})

		d.db = db
	}

	return d.db
}

func (d *diContainer) LocalStorage() *localstorage.LocalStorage {
	if d.localStorage == nil {
		d.muLocalStorage.Lock()
		defer d.muLocalStorage.Unlock()

		if d.localStorage != nil {
			return d.localStorage
		}

		slog.Info("Init LocalStorage", "LocalStorage", d.model)
		ls := localstorage.NewLocalStorage(d.DB(), d.DB())
		d.localStorage = ls
	}

	return d.localStorage
}

func (d *diContainer) Model() *windows.Model {
	if d.model == nil {
		d.muModel.Lock()
		defer d.muModel.Unlock()

		if d.model != nil {
			return d.model
		}

		slog.Info("Init Model", "model", d.model)
		model := windows.InitialModel(d.LocalStorage())
		d.model = model
	}

	return d.model
}

func (d *diContainer) FullModel() *windows.Model {
	if d.fullModel == nil {
		d.muFullModel.Lock()
		defer d.muFullModel.Unlock()

		if d.fullModel != nil {
			return d.fullModel
		}

		slog.Info("Init Full Model", "model", d.fullModel)
		fullModel := windows.PostInitModel(d.Model(), d.KafkaClusterListComponent(), d.KafkaTopicListComponent(), d.KafkaPartitionListComponent(),
			d.KafkaReadWriteTabsComponent(), d.OverlayWindowForKafkaReadDetailedMessageProvider(),
			d.OverlayWindowForHelpProvider())
		d.fullModel = fullModel
	}

	return d.fullModel
}

func (d *diContainer) KafkaClusterListComponent() windows.KafkaClusterList {
	if d.kafkaClusterList == nil {
		d.muKafkaClusterList.Lock()
		defer d.muKafkaClusterList.Unlock()

		if d.kafkaClusterList != nil {
			return d.kafkaClusterList
		}

		slog.Info("Init KafkaClusterListComponent", "KafkaClusterListComponent", d.kafkaClusterList)
		d.kafkaClusterList = components.CreateKafkaClustersListAddValues(d.LocalStorage(), d.Model())
	}

	return d.kafkaClusterList
}

func (d *diContainer) KafkaTopicListComponent() *components.KafkaTopicList {
	if d.kafkaTopicList == nil {
		d.muKafkaTopicList.Lock()
		defer d.muKafkaTopicList.Unlock()

		if d.kafkaTopicList != nil {
			return d.kafkaTopicList
		}

		slog.Info("Init KafkaTopicListComponent", "KafkaTopicListComponent", d.kafkaTopicList)
		d.kafkaTopicList = components.CreateKafkaTopicsList(d.Model(), d.KafkaTopicsProvider())
	}

	return d.kafkaTopicList
}

func (d *diContainer) KafkaTopicsProvider() components.KafkaTopicsProvider {
	slog.Info("Init KafkaTopicsProvider", "KafkaTopicsProvider", d.kafkaConnector)
	return d.createKafkaConnector()
}

func (d *diContainer) createKafkaConnector() *connection.KafkaConnectorProvider {
	if d.kafkaConnector == nil {
		d.muKafkaConnector.Lock()
		defer d.muKafkaConnector.Unlock()

		if d.kafkaConnector != nil {
			return d.kafkaConnector
		}

		// closer.Add("кафка", func(_ context.Context) error {
		// 	return d.kafkaConnector.Close()
		// })

		d.kafkaConnector = connection.NewKafkaConnector()
	}

	return d.kafkaConnector
}

func (d *diContainer) KafkaPartitionsProvider() *connection.KafkaConnectorProvider {
	slog.Info("Init KafkaPartitionsProvider", "KafkaPartitionsProvider", d.kafkaConnector)
	return d.createKafkaConnector()
}

func (d *diContainer) KafkaPartitionListComponent() *components.KafkaPartitionList {
	if d.kafkaPartitionList == nil {
		d.muKafkaPartitionList.Lock()
		defer d.muKafkaPartitionList.Unlock()

		if d.kafkaPartitionList != nil {
			return d.kafkaPartitionList
		}

		slog.Info("Init KafkaPartitionListComponent", "KafkaPartitionListComponent", d.kafkaPartitionList)
		d.kafkaPartitionList = components.CreateKafkaPartitionsList(d.Model(), d.KafkaPartitionsProvider())
	}

	return d.kafkaPartitionList
}

func (d *diContainer) KafkaReadWriteTabsComponent() *readwritetabs.KafkaReadWriteTabsComponent {
	if d.kafkaReadWriteTabs == nil {
		d.muKafkaReadWriteTabs.Lock()
		defer d.muKafkaReadWriteTabs.Unlock()

		if d.kafkaReadWriteTabs != nil {
			return d.kafkaReadWriteTabs
		}

		slog.Info("Init KafkaReadWriteTabsComponent", "KafkaReadWriteTabsComponent", d.kafkaReadWriteTabs)
		d.kafkaReadWriteTabs = readwritetabs.CreateKafkaReadWriteTabsComponent(d.Model(), d.KafkaSendMessageTextAreaComponent(),
			d.KafkaSendMessageTableComponent(), d.KafkaSenderReader(), d.LocalStorage(), d.KafkaReadMessagesTableComponent(),
			d.SpinnerWhenLoadMessagesFromKafkaComponent())
	}

	return d.kafkaReadWriteTabs
}

func (d *diContainer) KafkaSendMessageTextAreaComponent() *readwritetabs.KafkaSendMessageTextAreaComponent {
	if d.kafkaSendMessageTextArea == nil {
		d.muKafkaSendMessageTextArea.Lock()
		defer d.muKafkaSendMessageTextArea.Unlock()

		if d.kafkaSendMessageTextArea != nil {
			return d.kafkaSendMessageTextArea
		}

		slog.Info("Init KafkaSendMessageTextAreaComponent", "KafkaSendMessageTextAreaComponent", d.kafkaSendMessageTextArea)
		d.kafkaSendMessageTextArea = readwritetabs.CreateKafkaSendMessageTextArea(d.Model())
	}

	return d.kafkaSendMessageTextArea
}

func (d *diContainer) KafkaSendMessageTableComponent() *readwritetabs.KafkaSendMessageTableComponent {
	if d.kafkaSendMessageTableComponent == nil {
		d.muKafkaSendMessageTextArea.Lock()
		defer d.muKafkaSendMessageTextArea.Unlock()

		if d.kafkaSendMessageTableComponent != nil {
			return d.kafkaSendMessageTableComponent
		}

		slog.Info("Init KafkaSendMessageTableComponent", "KafkaSendMessageTableComponent", d.kafkaSendMessageTableComponent)
		d.kafkaSendMessageTableComponent = readwritetabs.CreateKafkaSendMessageTable(d.Model())
	}

	return d.kafkaSendMessageTableComponent
}

func (d *diContainer) KafkaReadMessagesTableComponent() *readwritetabs.KafkaReadMessagesTableComponent {
	if d.kafkaReadMessagesTableComponent == nil {
		d.muKafkaReadMessagesTable.Lock()
		defer d.muKafkaReadMessagesTable.Unlock()

		if d.kafkaReadMessagesTableComponent != nil {
			return d.kafkaReadMessagesTableComponent
		}

		slog.Info("Init KafkaReadMessagesTableComponent", "KafkaReadMessagesTableComponent", d.kafkaReadMessagesTableComponent)
		d.kafkaReadMessagesTableComponent = readwritetabs.CreateKafkaReadMessagesTable(d.Model())
	}

	return d.kafkaReadMessagesTableComponent
}

func (d *diContainer) KafkaSenderReader() *connection.KafkaConnectorProvider {
	slog.Info("Init KafkaPartitionsProvider", "KafkaPartitionsProvider", d.kafkaConnector)
	return d.createKafkaConnector()
}

func (d *diContainer) OverlayWindowForKafkaReadDetailedMessageProvider() *components.OverlayWindowForKafkaReadDetailedMessageProvider {
	if d.overlayForKafkaReadDetailedMessageProvider == nil {
		d.muOverlayForKafkaReadDetailedMessageProvider.Lock()
		defer d.muOverlayForKafkaReadDetailedMessageProvider.Unlock()

		if d.overlayForKafkaReadDetailedMessageProvider != nil {
			return d.overlayForKafkaReadDetailedMessageProvider
		}

		slog.Info("Init overlayForKafkaReadDetailedMessageProvider", "overlayForKafkaReadDetailedMessageProvider", d.overlayForKafkaReadDetailedMessageProvider)
		d.overlayForKafkaReadDetailedMessageProvider = components.CreateNewWindowForKafkaReadDetailedMessageProvider()
	}

	return d.overlayForKafkaReadDetailedMessageProvider
}

func (d *diContainer) SpinnerWhenLoadMessagesFromKafkaComponent() *readwritetabs.SpinnerWhenLoadMessagesFromKafkaComponent {
	if d.spinnerWhenLoadMessagesFromKafkaComponent == nil {
		d.muSpinnerWhenLoadMessagesFromKafkaComponent.Lock()
		defer d.muSpinnerWhenLoadMessagesFromKafkaComponent.Unlock()

		if d.spinnerWhenLoadMessagesFromKafkaComponent != nil {
			return d.spinnerWhenLoadMessagesFromKafkaComponent
		}

		slog.Info("Init SpinnerWhenLoadMessagesFromKafkaComponent", "SpinnerWhenLoadMessagesFromKafkaComponent", d.spinnerWhenLoadMessagesFromKafkaComponent)
		d.spinnerWhenLoadMessagesFromKafkaComponent = readwritetabs.NewSpinnerWhenLoadMessagesFromKafkaComponent()
	}

	return d.spinnerWhenLoadMessagesFromKafkaComponent
}

func (d *diContainer) OverlayWindowForAppendNewServerProvider() *components.OverlayWindowForAppendNewServerProvider {
	if d.overlayWindowForAppendNewServerProvider == nil {
		d.muOverlayWindowForAppendNewServerProvider.Lock()
		defer d.muOverlayWindowForAppendNewServerProvider.Unlock()

		if d.overlayWindowForAppendNewServerProvider != nil {
			return d.overlayWindowForAppendNewServerProvider
		}

		slog.Info("Init overlayWindowForAppendNewServerProvider", "overlayWindowForAppendNewServerProvider", d.overlayWindowForAppendNewServerProvider)
		d.overlayWindowForAppendNewServerProvider = components.CreateNewOverlayWindowForAppendNewServerProvider()
	}

	return d.overlayWindowForAppendNewServerProvider
}

func (d *diContainer) OverlayWindowForDeleteNewServerProvider() *components.OverlayWindowForDeleteNewServerProvider {
	if d.overlayWindowForDeleteNewServerProvider == nil {
		d.muOverlayWindowForDeleteNewServerProvider.Lock()
		defer d.muOverlayWindowForDeleteNewServerProvider.Unlock()

		if d.overlayWindowForDeleteNewServerProvider != nil {
			return d.overlayWindowForDeleteNewServerProvider
		}

		slog.Info("Init overlayWindowForDeleteNewServerProvider", "overlayWindowForDeleteNewServerProvider", d.overlayWindowForDeleteNewServerProvider)
		d.overlayWindowForDeleteNewServerProvider = components.CreateNewOverlayWindowForDeleteNewServerProvider()
	}

	return d.overlayWindowForDeleteNewServerProvider
}

func (d *diContainer) OverlayWindowForHelpProvider() *components.OverlayWindowForHelpProvider {
	if d.overlayWindowForHelpProvider == nil {
		d.muOverlayWindowForHelpProvider.Lock()
		defer d.muOverlayWindowForHelpProvider.Unlock()

		if d.overlayWindowForHelpProvider != nil {
			return d.overlayWindowForHelpProvider
		}

		slog.Info("Init overlayWindowForHelpProvider", "overlayWindowForHelpProvider", d.overlayWindowForHelpProvider)
		d.overlayWindowForHelpProvider = components.CreateNewOverlayWindowForHelpProvider()
	}

	return d.overlayWindowForHelpProvider
}
