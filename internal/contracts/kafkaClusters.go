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
	GetKafkaClusters() ([]*KafkaCluster, error)
}

type UserChosenKafkaCluster struct {
	ChosenCluster string
}

type UserWantsDeleteClusterFromList struct {
	ChosenCluster string
}

type UpdateListOfClustersEvent struct {
	Clusters []*KafkaCluster
}

type ViewOfCluster struct {
	title    string
	subtitle string
}

type SuccessClusterLoadedEvent struct {
	KafkaClusters     map[string]*KafkaCluster
	KafkaClustersList []*KafkaCluster
}

type ErrorClusterLoadedEvent struct {
	Err error
}
