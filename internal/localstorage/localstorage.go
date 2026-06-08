package localstorage

import (
	"marat/fayz/kafka_reader_writer/internal/database"
)

type kafkaCluster struct {
	title              string
	url                string
	trustStorePath     string
	trustStorePassword string
	username           string
	password           string
	saslMechanism      string
}

func (kc kafkaCluster) Title() string {
	return kc.title
}

func (kc kafkaCluster) Url() string {
	return kc.url
}
func (kc kafkaCluster) TrustStorePath() string {
	return kc.trustStorePath
}
func (kc kafkaCluster) TrustStorePassword() string {
	return kc.trustStorePassword
}
func (kc kafkaCluster) Username() string {
	return kc.username
}
func (kc kafkaCluster) Password() string {
	return kc.password
}
func (kc kafkaCluster) SaslMechanism() string {
	return kc.saslMechanism
}

type LocalStorage interface {
	GetKafkaClusters() []kafkaCluster
}

type localStorage struct {
	db database.DB
}

func (ls localStorage) GetKafkaClusters() []kafkaCluster {
	clusters := make([]kafkaCluster, 0, 1)
	clusters = append(clusters, kafkaCluster{title: "sfa", url: "172.16.15.171:9093"})
	return clusters
}

func NewLocalStorage(db database.DB) LocalStorage {
	return localStorage{db: db}
}
