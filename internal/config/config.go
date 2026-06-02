package config

type Config struct {
	username            string
	password            string
	broker              string
	topic               string
	partition           int
	certificateFilePath string
	DatabaseUrl         string
}

var appConfig = &Config{
	// PeriodOfSynchronization:                  5 * time.Second,
	// ServerPort:                               ":8019",
	// MaxGoroutinesToUpdateLocalStorage:        1,
	// MaxGoroutinesToSearchChangesInLocalFiles: 1,
	// // PathsToSynchronize:                       []string{"/Users/mfayz/Desktop/filesToSynchronize"},
	DatabaseUrl: "./database.db",
}

func AppConfig() *Config {
	return appConfig
}
