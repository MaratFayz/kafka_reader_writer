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
	// defer func() {
	// 	if recover() != nil {
	// 		slog.Error("EEEEEEEEEEEEE")
	// 	}
	// }()

	if v, ok := k.topics[cluster.Title]; ok == true {
		slog.Error("in map have topics", "v", v)

		return v, nil
	}

	// slog.Error("in map NOT topics")

	var conn *kafka.Conn
	if v, ok := k.connections[cluster.Title]; ok == true {
		slog.Error("in map have connection", "v", v)

		conn = v
	}

	// slog.Error("in map NOT connections")

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

	// slog.Error("in kafka", "v", k)

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
	slog.Error("cccccccccccccccccccccccccccccc", "c", t2)

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
	// conn := k.connections[kafkaCluster.Title]
	// conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// // conn.
	// _, err := conn.WriteMessages(
	// 	kafka.Message{Topic: kafkaTopic, Value: []byte(textToSend)},
	// 	// kafka.Message{Value: []byte("two!")},
	// 	// kafka.Message{Value: []byte("three!")},
	// )

	// if err != nil {
	// 	log.Fatal("failed to write messages:", err)
	// }

	// return nil

	// slog.Error("NULL", "cluster", kafkaCluster)
	// b, err := conn.Brokers()

	// brokers := make([]string, 0, len(b))

	// for _, br := range b {
	// 	brokers = append(brokers, br.Host)
	// }

	// if err != nil {
	// 	log.Fatal("failed to write messages:", err)
	// }

	// writer := kafka.NewWriter(kafka.WriterConfig{
	// 	// Addr:     kafka.TCP(kafkaCluster.Url),
	// 	// Username: kafkaCluster.Username,
	// 	// Password: kafkaCluster.Password,
	// 	// Balancer: &kafka.LeastBytes{},
	// 	Dialer: createDialer(kafkaCluster.Username, kafkaCluster.Password, kafkaCluster.TrustStorePath, kafkaCluster.Url),
	// 	Topic:  kafkaTopic,
	// 	// Brokers: []string{kafkaCluster.Url},
	// 	Brokers: brokers,
	// })
	dialer := createDialer(kafkaCluster.Username, kafkaCluster.Password, kafkaCluster.TrustStorePath, kafkaCluster.Url)
	// writer := kafka.Writer{
	// 	Addr: kafka.TCP(kafkaCluster.Url),
	// 	// Dialer: createDialer(kafkaCluster.Username, kafkaCluster.Password, kafkaCluster.TrustStorePath, kafkaCluster.Url),
	// 	Topic: kafkaTopic,
	// 	// Brokers: []string{kafkaCluster.Url},
	// 	// Brokers: brokers,
	// 	Transport: &kafka.Transport{
	// 		Dial: dialer.DialFunc, // Используем диалер для транспорта
	// 		TLS:  dialer.TLS,
	// 	},
	// 	Logger: kafka.LoggerFunc(log.Printf),
	// }

	// defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	// err := writer.WriteMessages(ctx, kafka.Message{Value: []byte(textToSend)}, kafka.Message{Value: []byte(textToSend)})

	// if err != nil {
	// 	log.Fatalf("Error when write to kafka: %s %s % X", err, textToSend, []byte(textToSend))
	// 	return err
	// }

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
		// kafka.Message{Value: []byte("two!")},
		// kafka.Message{Value: []byte("three!")},
	)

	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	return nil
}

// func (k *KafkaConnectorProvider) Close() error {
// 	var err error
// 	for _, c := range k.connections {
// 		err = c.Close()
// 	}

// 	if err != nil {

// 	}
// }

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

	slog.Error("l", "l", lastOffset)

	// 3. Создаем Reader для чтения с найденного оффсета
	// Важно: используем тот же контекст или новый для чтения
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaCluster.Url},
		Topic:   kafkaTopic,
		Dialer:  dialer,
	})
	defer reader.Close()

	// // Устанавливаем оффсет на последнее сообщение
	err = reader.SetOffset(lastOffset - 3)
	if err != nil {
		return nil, fmt.Errorf("failed to set offset: %w", err)
	}

	// 4. Читаем одно сообщение
	// Используем новый контекст для операции чтения
	readCtx, readCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer readCancel()

	// msg, err := reader.ReadMessage(readCtx)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read message: %w", err)
	// }

	// log.Fatal("offset", "offset", offset, "w", w)

	res := make([]*contracts.ReadMessagesRow, 0)

	for range 3 {
		// message, err := conn.ReadMessage(2147483566)
		message, err := reader.ReadMessage(readCtx)

		if err != nil {
			slog.Error("failed to read message:", "err", err)
		}

		// slog.Error("m", "mm", string(message.Value))

		res = append(res, &contracts.ReadMessagesRow{Row: []string{strconv.Itoa((int(message.Offset))), message.Time.GoString(), string(message.Value)}})
	}

	// batch := conn.ReadBatch(1000, 2147483566)
	// defer batch.Close()
	// if err := batch.Close(); err != nil {
	// 	slog.Error("Batch is failed when close", "err", err)
	// 	return nil, err
	// }

	// if err := batch.Err(); err != nil {
	// 	log.Fatal("Batch is failed", "err", err)
	// 	return nil, err
	// }

	// for range 50 {
	// 	message, err := batch.ReadMessage()

	// 	if err != nil {
	// 		log.Fatal("failed to read message:", err)
	// 	}

	// 	m := string(message.Value)

	// 	res = append(res, &contracts.ReadMessagesRow{Row: []string{strconv.Itoa((int(message.Offset))), message.Time.GoString(), m}})
	// }

	// slog.Error("Mes read kafka------------------------>", "mes", res)

	// res = append(res, &contracts.ReadMessagesRow{Row: []string{"1", "2", "3"}})
	return res, nil
}
