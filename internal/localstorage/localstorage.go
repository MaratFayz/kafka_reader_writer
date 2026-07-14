package localstorage

import (
	"marat/fayz/kafka_reader_writer/internal/contracts"
)

type LocalStorage struct {
	kafkaClusterProvider contracts.KafkaClusterProvider
	db                   DB
}

type DB interface {
	SaveMessage(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, messageToSave string) error
	GetMessages(cluster *contracts.KafkaCluster, topic string) ([]*contracts.SentMessagesRow, error)
}

func (ls *LocalStorage) GetKafkaClusters() []*contracts.KafkaCluster {
	a := ls.kafkaClusterProvider.GetKafkaClusters()

	return a
}

func NewLocalStorage(kafkaClusterProvider contracts.KafkaClusterProvider, db DB) *LocalStorage {
	return &LocalStorage{kafkaClusterProvider: kafkaClusterProvider, db: db}
}

func (ls *LocalStorage) SaveMessage(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, messageToSave string) error {
	return ls.db.SaveMessage(kafkaCluster, kafkaTopic, messageToSave)
}

func (ls *LocalStorage) GetMessages(cluster *contracts.KafkaCluster, topic string) ([]*contracts.SentMessagesRow, error) {
	return ls.db.GetMessages(cluster, topic)
}
