package contracts

type UpdateListOfTopicsEvent struct {
	Topics []string
}

type LoadingKafkaTopicErrMsg struct {
	Err error
}

type LoadedKafkaTopicsMsg struct {
	Cluster string
	Topics  []string
}

type UserChosenKafkaTopic struct {
	Topic string
}
