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
	DeleteKafkaCluster(clusterName string) error
	SaveNewKafkaCluster(Title, Password, Trust_store_path, Url, Username string) error
	SaveEditedKafkaCluster(Title, Password, Trust_store_path, Url, Username string) error
}

func (ls *LocalStorage) GetKafkaClusters() ([]*contracts.KafkaCluster, error) {
	a, err := ls.kafkaClusterProvider.GetKafkaClusters()

	if err != nil {
		return nil, err
	}

	return a, nil
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

func (ls *LocalStorage) DeleteKafkaCluster(clusterName string) error {
	return ls.db.DeleteKafkaCluster(clusterName)
}

func (ls *LocalStorage) SaveNewKafkaCluster(Title, Password, Trust_store_path, Url, Username string) error {
	return ls.db.SaveNewKafkaCluster(Title, Password, Trust_store_path, Url, Username)
}

func (ls *LocalStorage) SaveEditedKafkaCluster(Title, Password, Trust_store_path, Url, Username string) error {
	return ls.db.SaveEditedKafkaCluster(Title, Password, Trust_store_path, Url, Username)
}
