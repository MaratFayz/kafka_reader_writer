package contracts

type KafkaCluster struct {
	Title          string
	Url            string
	TrustStorePath string
	// TrustStorePassword string
	Username string
	Password string
	// SaslMechanism      string
}

type KafkaClusterProvider interface {
	GetKafkaClusters() []*KafkaCluster
}

type UserChosenKafkaCluster struct {
	ChosenCluster string
}

type UserWantsDeleteClusterFromList struct {
	ChosenCluster string
}
