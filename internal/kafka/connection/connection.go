package connection

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"os"
	"sort"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type KafkaConnectorProvider struct {
	topics      map[string][]string
	partitions  map[string]map[string][]int
	connections map[string]*kafka.Conn
}

func NewKafkaConnector() *KafkaConnectorProvider {
	return &KafkaConnectorProvider{
		topics:      make(map[string][]string),
		partitions:  make(map[string]map[string][]int),
		connections: make(map[string]*kafka.Conn),
	}
}

func (k *KafkaConnectorProvider) GetTopicsByClusterName(clusterName *contracts.KafkaCluster) ([]string, error) {
	if v, ok := k.topics[clusterName.Title]; ok == true {
		slog.Error("in map have topics", "v", v)

		return v, nil
	}

	// slog.Error("in map NOT topics")

	var conn *kafka.Conn
	if v, ok := k.connections[clusterName.Title]; ok == true {
		slog.Error("in map have connection", "v", v)

		conn = v
	}

	// slog.Error("in map NOT connections")

	conn, err := createConnection(clusterName.Username, clusterName.Password, clusterName.TrustStorePath, clusterName.Url)
	if err != nil {
		slog.Error("Connection to Kafka error")
	}

	partitions, err := conn.ReadPartitions()
	if err != nil {
		slog.Error("Partitions in Kafka error")
	}

	topics := map[string]struct{}{}
	partitionMap := map[string][]int{}
	for _, p := range partitions {
		topics[p.Topic] = struct{}{}

		pv, ok := partitionMap[p.Topic]
		if !ok {
			partitionMap[p.Topic] = make([]int, 0)
		}
		pv = partitionMap[p.Topic]
		pv = append(pv, p.ID)
		partitionMap[p.Topic] = pv
	}

	for _, v := range partitionMap {
		sort.Ints(v)
	}

	topicsList := make([]string, 0, len(topics))
	for k := range topics {
		topicsList = append(topicsList, k)
	}

	sort.Strings(topicsList)

	k.topics[clusterName.Title] = topicsList
	if _, ok := k.partitions[clusterName.Title]; ok != true {
		k.partitions[clusterName.Title] = make(map[string][]int)
	}

	k.partitions[clusterName.Title] = partitionMap

	// slog.Error("in kafka", "v", k)

	return topicsList, nil
}

func createConnection(username, password, certificateFilePath, brokerUrl string) (*kafka.Conn, error) {
	mechanism := plain.Mechanism{
		Username: username,
		Password: password,
	}

	ca, err := os.Open(certificateFilePath)
	if err != nil {
		panic(err)
	}
	cab, err := ioutil.ReadAll(ca)
	if err != nil {
		panic(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cab)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
		// ClientCAs: ,
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
		TLS:           tlsConfig,
	}

	conn, err := dialer.Dial("tcp", brokerUrl)

	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return nil, err
	}

	return conn, nil
}

func (k *KafkaConnectorProvider) GetPartitionsByClusterNameAndTopic(topicName string, cluster *contracts.KafkaCluster) []int {
	t := k.partitions[cluster.Title]

	t2 := t[topicName]
	// slog.Error("cccccccccccccccccccccccccccccc", "c", t2)

	return t2
}

// type Kafka interface {
// }

// func sendMessage(config *Config, message string) {
// 	// conn, err := dialer.DialLeader(context.Background(), "tcp", config.broker, config.topic, config.partition)
// 	if err != nil {
// 		log.Fatal("failed to dial leader:", err)
// 	}

// 	partitions, err := conn.ReadPartitions()
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// Собираем имена топиков
// 	m := map[string]struct{}{}
// 	for _, p := range partitions {
// 		m[p.Topic] = struct{}{}
// 	}

// 	// Выводим имена топиков
// 	for k := range m {
// 		fmt.Println(k)
// 	}

// 	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
// 	_, err = conn.WriteMessages(
// 		kafka.Message{Topic: config.topic, Value: []byte(message)},
// 		// kafka.Message{Value: []byte("two!")},
// 		// kafka.Message{Value: []byte("three!")},
// 	)
// 	if err != nil {
// 		log.Fatal("failed to write messages:", err)
// 	}

// 	if err := conn.Close(); err != nil {
// 		log.Fatal("failed to close writer:", err)
// 	}
// }

func (k *KafkaConnectorProvider) Send(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, kafkaPartition string, textToSend string) error {
	return nil
}

func (k *KafkaConnectorProvider) Close() error {
	var err error
	for _, c := range k.connections {
		err = c.Close()
	}

	if err != nil {

	}
}
