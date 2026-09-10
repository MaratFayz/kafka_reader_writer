package contracts

type UpdateListOfPartitionsEvent struct {
	Topic      string
	Partitions []int
}

type ErrorLoadedKafkaPartitionsErrMsg struct {
	Err error
}

type SuccessLoadedKafkaPartitionsMsg struct {
	Cluster    string
	Topic      string
	Partitions []int
}

type UserChosenPartition struct {
	Partition string
}
