package localstorage

import "marat/fayz/kafka_reader_writer/internal/database"

type LocalStorage interface {
}

type localStorage struct {
	db database.DB
}

func NewLocalStorage(db database.DB) LocalStorage {
	return localStorage{db: db}
}
