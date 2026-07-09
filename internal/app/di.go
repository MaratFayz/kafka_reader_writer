package app

import (
	"context"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/closer"
	"marat/fayz/kafka_reader_writer/internal/components"
	"marat/fayz/kafka_reader_writer/internal/config"
	"marat/fayz/kafka_reader_writer/internal/database"
	"marat/fayz/kafka_reader_writer/internal/kafka/connection"
	"marat/fayz/kafka_reader_writer/internal/localstorage"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"os"
	"sync"
)

type diContainer struct {
	appConfig                      *config.Config
	db                             database.DB
	localStorage                   localstorage.LocalStorage
	model                          *windows.Model
	fullModel                      *windows.Model
	kafkaClusterList               *components.KafkaClusterList
	kafkaTopicList                 *components.KafkaTopicList
	kafkaConnector                 *connection.KafkaConnectorProvider
	kafkaPartitionList             *components.KafkaPartitionList
	kafkaReadWriteTabs             *components.KafkaReadWriteTabsComponent
	kafkaSendMessageTextArea       *components.KafkaSendMessageTextAreaComponent
	kafkaSendMessageTableComponent *components.KafkaSendMessageTableComponent

	muAppConfig                sync.Mutex
	muDb                       sync.Mutex
	muLocalStorage             sync.Mutex
	muModel                    sync.Mutex
	muFullModel                sync.Mutex
	muKafkaClusterList         sync.Mutex
	muKafkaTopicList           sync.Mutex
	muKafkaConnector           sync.Mutex
	muKafkaPartitionList       sync.Mutex
	muKafkaReadWriteTabs       sync.Mutex
	muKafkaSendMessageTextArea sync.Mutex
}

// newDIContainer создаёт новый пустой контейнер.
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

func (d *diContainer) DB() database.DB {
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

func (d *diContainer) LocalStorage() localstorage.LocalStorage {
	if d.localStorage == nil {
		d.muLocalStorage.Lock()
		defer d.muLocalStorage.Unlock()

		if d.localStorage != nil {
			return d.localStorage
		}

		slog.Info("Init LocalStorage", "LocalStorage", d.model)
		ls := localstorage.NewLocalStorage(d.DB())
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
			d.KafkaReadWriteTabsComponent())
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
	if d.kafkaConnector == nil {
		d.muKafkaConnector.Lock()
		defer d.muKafkaConnector.Unlock()

		if d.kafkaConnector != nil {
			return d.kafkaConnector
		}

		slog.Info("Init KafkaTopicsProvider", "KafkaTopicsProvider", d.kafkaConnector)
		d.kafkaConnector = connection.NewKafkaConnector()
	}

	return d.kafkaConnector
}

func (d *diContainer) KafkaPartitionsProvider() components.KafkaPartitionsProvider {
	if d.kafkaConnector == nil {
		d.muKafkaConnector.Lock()
		defer d.muKafkaConnector.Unlock()

		if d.kafkaConnector != nil {
			return d.kafkaConnector
		}

		slog.Info("Init KafkaPartitionsProvider", "KafkaPartitionsProvider", d.kafkaConnector)
		d.kafkaConnector = connection.NewKafkaConnector()
	}

	return d.kafkaConnector
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

func (d *diContainer) KafkaReadWriteTabsComponent() *components.KafkaReadWriteTabsComponent {
	if d.kafkaReadWriteTabs == nil {
		d.muKafkaReadWriteTabs.Lock()
		defer d.muKafkaReadWriteTabs.Unlock()

		if d.kafkaReadWriteTabs != nil {
			return d.kafkaReadWriteTabs
		}

		slog.Info("Init KafkaReadWriteTabsComponent", "KafkaReadWriteTabsComponent", d.kafkaReadWriteTabs)
		d.kafkaReadWriteTabs = components.CreateKafkaReadWriteTabsComponent(d.Model(), d.KafkaSendMessageTextAreaComponent(),
			d.KafkaSendMessageTableComponent())
	}

	return d.kafkaReadWriteTabs
}

func (d *diContainer) KafkaSendMessageTextAreaComponent() *components.KafkaSendMessageTextAreaComponent {
	if d.kafkaSendMessageTextArea == nil {
		d.muKafkaSendMessageTextArea.Lock()
		defer d.muKafkaSendMessageTextArea.Unlock()

		if d.kafkaSendMessageTextArea != nil {
			return d.kafkaSendMessageTextArea
		}

		slog.Info("Init KafkaSendMessageTextAreaComponent", "KafkaSendMessageTextAreaComponent", d.kafkaSendMessageTextArea)
		d.kafkaSendMessageTextArea = components.CreateKafkaSendMessageTextArea(d.Model())
	}

	return d.kafkaSendMessageTextArea
}

func (d *diContainer) KafkaSendMessageTableComponent() *components.KafkaSendMessageTableComponent {
	if d.kafkaSendMessageTableComponent == nil {
		d.muKafkaSendMessageTextArea.Lock()
		defer d.muKafkaSendMessageTextArea.Unlock()

		if d.kafkaSendMessageTableComponent != nil {
			return d.kafkaSendMessageTableComponent
		}

		slog.Info("Init KafkaSendMessageTableComponent", "KafkaSendMessageTableComponent", d.kafkaSendMessageTableComponent)
		d.kafkaSendMessageTableComponent = components.CreateKafkaSendMessageTable(d.Model())
	}

	return d.kafkaSendMessageTableComponent
}
