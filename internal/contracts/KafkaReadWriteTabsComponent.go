package contracts

type UpdateHistoryMessagesTableEvent struct {
	Topic      string
	Partitions []int
}

type ClearHistoryMessagesTableEvent struct {
}

type ErrorWhenSendMessageToKafka struct {
	Error error
}

type SuccessWhenSendMessageToKafka struct {
}

type ErrorWhenSaveMessageToHistory struct {
	Error error
}

type SuccessWhenSaveMessageToHistory struct {
}

type ErrorWhenGetMessagesFromHistory struct {
	Error error
}

type SuccessWhenGetMessagesFromHistory struct {
	Messages []*SentMessagesRow
}

type ErrorWhenSetMessagesHistoryMessagesTableCmd struct {
	Error error
}

type SuccessWhenSetMessagesHistoryMessagesTableCmd struct {
}

type SuccessWhenReadMessagesFromKafka struct {
	Messages []*ReadMessagesRow
}

type ErrorWhenReadMessagesFromKafka struct {
	Error error
}

type SuccessWhenSetMessagesToReadMessagesTableCmd struct {
}

type ErrorWhenSetMessagesToReadMessagesTableCmd struct {
	Error error
}

type UpdateListOfMessagesFromHistoryEvent struct {
	Messages []*SentMessagesRow
}

type SendMessageToKafka struct {
	Message string
}

type UserWantsReadMessagesFromKafka struct {
}

type StartSpinningWhenReadMessagesFromKafka struct {
}

type UpdateListOfReadMessagesFromKafkaEvent struct {
	Messages []*ReadMessagesRow
}

type StopSpinningWhenReadMessagesFromKafka struct {
}
