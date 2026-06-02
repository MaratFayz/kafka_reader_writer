package app

import (
	"context"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/closer"
	"marat/fayz/kafka_reader_writer/internal/config"
	"marat/fayz/kafka_reader_writer/internal/database"
	"marat/fayz/kafka_reader_writer/internal/localstorage"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"os"
	"sync"
)

type diContainer struct {
	appConfig    *config.Config
	db           database.DB
	localStorage localstorage.LocalStorage
	model        *windows.Model

	muAppConfig    sync.Mutex
	muDb           sync.Mutex
	muLocalStorage sync.Mutex
	muModel        sync.Mutex
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
