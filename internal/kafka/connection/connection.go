package connection

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"os"
	"sort"
	"strconv"
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

func (k *KafkaConnectorProvider) GetTopicsByClusterName(cluster *contracts.KafkaCluster) ([]string, error) {
	if v, ok := k.topics[cluster.Title]; ok == true {
		slog.Error("in map have topics", "v", v)

		return v, nil
	}

	var conn *kafka.Conn
	if v, ok := k.connections[cluster.Title]; ok == true {
		slog.Error("in map have connection", "v", v)

		conn = v
	}

	conn, err := createConnection(cluster.Username, cluster.Password, cluster.TrustStorePath, cluster.Url)
	if err != nil {
		slog.Error("Connection to Kafka error")
		return nil, err
	}

	k.connections[cluster.Title] = conn

	partitions, err := conn.ReadPartitions()
	if err != nil {
		slog.Error("Partitions in Kafka error")
		return nil, err
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

	k.topics[cluster.Title] = topicsList
	if _, ok := k.partitions[cluster.Title]; ok != true {
		k.partitions[cluster.Title] = make(map[string][]int)
	}

	k.partitions[cluster.Title] = partitionMap

	return topicsList, nil
}

func createDialer(username, password, certificateFilePath, brokerUrl string) *kafka.Dialer {
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
		// InsecureSkipVerify: false, // Всегда false в production
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
		TLS:           tlsConfig,
	}

	return dialer
}

func createConnection(username, password, certificateFilePath, brokerUrl string) (*kafka.Conn, error) {
	dialer := createDialer(username, password, certificateFilePath, brokerUrl)

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

	return t2
}

func (k *KafkaConnectorProvider) Send(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, kafkaPartition string, textToSend string) error {
	dialer := createDialer(kafkaCluster.Username, kafkaCluster.Password, kafkaCluster.TrustStorePath, kafkaCluster.Url)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	i, err := strconv.Atoi(kafkaPartition)

	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return err
	}

	conn, err := dialer.DialLeader(ctx, "tcp", kafkaCluster.Url, kafkaTopic, i)

	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return err
	}

	_, err = conn.WriteMessages(
		kafka.Message{Topic: kafkaTopic, Value: []byte(textToSend)},
	)

	if err != nil {
		log.Fatal("failed to write messages:", err) //         2026…2026/07/19 23:10:29 failed to write messages:[29] Topic Authorization Failed: the client is not authorized to access the requested topic                       │  │
	}

	return nil
}

func (k *KafkaConnectorProvider) Read(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, kafkaPartition string, numberOfReadMessages int) ([]*contracts.ReadMessagesRow, error) {
	dialer := createDialer(kafkaCluster.Username, kafkaCluster.Password, kafkaCluster.TrustStorePath, kafkaCluster.Url)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	i, err := strconv.Atoi(kafkaPartition)

	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return nil, err
	}

	conn, err := dialer.DialLeader(ctx, "tcp", kafkaCluster.Url, kafkaTopic, i)

	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return nil, err
	}
	defer conn.Close()

	// 2. Получаем последний доступный оффсет
	lastOffset, err := conn.ReadLastOffset()
	if err != nil {
		return nil, fmt.Errorf("failed to read last offset: %w", err)
	}

	// Если сообщений в партиции нет, последний оффсет будет -1
	if lastOffset < 0 {
		return nil, fmt.Errorf("no messages in partition")
	}

	// 3. Создаем Reader для чтения с найденного оффсета
	// Важно: используем тот же контекст или новый для чтения
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaCluster.Url},
		Topic:   kafkaTopic,
		Dialer:  dialer,
	})
	defer reader.Close()

	// // Устанавливаем оффсет на последнее сообщение
	err = reader.SetOffset(lastOffset - int64(numberOfReadMessages))
	if err != nil {
		return nil, fmt.Errorf("failed to set offset: %w", err)
	}

	// 4. Читаем одно сообщение
	// Используем новый контекст для операции чтения
	readCtx, readCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer readCancel()

	res := make([]*contracts.ReadMessagesRow, 0)

	for range numberOfReadMessages {
		message, err := reader.ReadMessage(readCtx)

		if err != nil {
			break
		}

		res = append(res, &contracts.ReadMessagesRow{Row: []string{strconv.Itoa((int(message.Offset))), message.Time.String(), string(message.Key), string(message.Value)}})
	}

	return res, nil
}
