package contracts

type UpdateListOfTopicsEvent struct {
	Topics []string
}

type ErrorLoadedKafkaTopicsErrMsg struct {
	Err error
}

type SuccessLoadedKafkaTopicsMsg struct {
	Cluster *KafkaCluster
	Topics  []string
	//	KafkaClusters     map[string]*KafkaCluster
	// KafkaClustersList []*KafkaCluster
}

type UserChosenKafkaTopic struct {
	ChosenTopic string
}
