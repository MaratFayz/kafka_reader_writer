package localstorage

import "marat/fayz/kafka_reader_writer/internal/contracts"

// type kafkaCluster struct {
// 	title              string
// 	url                string
// 	trustStorePath     string
// 	trustStorePassword string
// 	username           string
// 	password           string
// 	saslMechanism      string
// }

// type KafkaCluster interface {
// 	Title() string
// 	Url() string
// 	TrustStorePath() string
// 	TrustStorePassword() string
// 	Username() string
// 	Password() string
// 	SaslMechanism() string
// }

// func (kc kafkaCluster) Title() string {
// 	return kc.title
// }

// func (kc kafkaCluster) Url() string {
// 	return kc.url
// }
// func (kc kafkaCluster) TrustStorePath() string {
// 	return kc.trustStorePath
// }
// func (kc kafkaCluster) TrustStorePassword() string {
// 	return kc.trustStorePassword
// }
// func (kc kafkaCluster) Username() string {
// 	return kc.username
// }
// func (kc kafkaCluster) Password() string {
// 	return kc.password
// }
// func (kc kafkaCluster) SaslMechanism() string {
// 	return kc.saslMechanism
// }

type LocalStorage struct {
	kafkaClusterProvider contracts.KafkaClusterProvider
}

// type DB interface {
// 	Close() error
// 	GetKafkaClusters() []KafkaCluster
// }

// type KafkaClusters interface {
// 	Clusters()
// }

// type kafkaClusters struct {
// 	clusters []*kafkaCluster
// }

// func (k *kafkaClusters) Clusters() []*kafkaCluster {
// 	return k.clusters
// }

func (ls *LocalStorage) GetKafkaClusters() []*contracts.KafkaCluster {
	a := ls.kafkaClusterProvider.GetKafkaClusters()

	// clusters := make([]*kafkaCluster, 0, 1)

	// for _, v := range a.Clusters() {
	// 	cluster := ls.mapKafkaClusters(v)
	// 	clusters = append(clusters, cluster)
	// }

	// return &kafkaClusters{clusters: clusters}
	return a
}

// func (ls *LocalStorage) mapKafkaClusters(v KafkaCluster) *kafkaCluster {
// 	return &kafkaCluster{title: v.Title(), url: v.Url(), username: v.Username(),
// 		password: v.Password(), trustStorePath: v.TrustStorePath()}
// }

func NewLocalStorage(kafkaClusterProvider contracts.KafkaClusterProvider) *LocalStorage {
	return &LocalStorage{kafkaClusterProvider: kafkaClusterProvider}
}

func (ls *LocalStorage) SaveMessage(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, messageToSave string) error {
	return nil
}
