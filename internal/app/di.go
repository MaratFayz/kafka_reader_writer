package app

import (
	"context"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/closer"
	"marat/fayz/kafka_reader_writer/internal/components"
	"marat/fayz/kafka_reader_writer/internal/config"
	"marat/fayz/kafka_reader_writer/internal/database"
	"marat/fayz/kafka_reader_writer/internal/localstorage"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"os"
	"sync"
)

type diContainer struct {
	appConfig        *config.Config
	db               database.DB
	localStorage     localstorage.LocalStorage
	model            *windows.Model
	fullModel        *windows.Model
	kafkaClusterList *components.KafkaClusterList
	kafkaTopicList   *components.KafkaTopicList

	muAppConfig        sync.Mutex
	muDb               sync.Mutex
	muLocalStorage     sync.Mutex
	muModel            sync.Mutex
	muFullModel        sync.Mutex
	muKafkaClusterList sync.Mutex
	muKafkaTopicList   sync.Mutex
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
		fullModel := windows.PostInitModel(d.Model(), d.KafkaClusterListComponent(), d.KafkaTopicListComponent())
		d.fullModel = fullModel
	}

	return d.fullModel
}

func (d *diContainer) KafkaClusterListComponent() *components.KafkaClusterList {
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
		d.kafkaTopicList = components.CreateKafkaTopicListAddValues(d.LocalStorage(), d.Model())
	}

	return d.kafkaTopicList
}
