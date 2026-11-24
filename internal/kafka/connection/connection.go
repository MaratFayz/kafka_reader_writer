package connection

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type KafkaConnectorProvider struct {
	mu         sync.Mutex
	dialers    map[string]*kafka.Dialer
	writers    map[writerKey]*kafka.Writer
	topics     map[string][]string
	partitions map[string]map[string][]int
}

// writerKey идентифицирует writer: он привязан к кластеру и фиксированной партиции.
type writerKey struct {
	clusterTitle string
	partition    int
}

func NewKafkaConnector() *KafkaConnectorProvider {
	return &KafkaConnectorProvider{
		dialers:    make(map[string]*kafka.Dialer),
		writers:    make(map[writerKey]*kafka.Writer),
		topics:     make(map[string][]string),
		partitions: make(map[string]map[string][]int),
	}
}

// dialerFor возвращает dialer для кластера, лениво создавая его при первом обращении.
// Так сертификат читается один раз за сессию, а не на каждую операцию.
func (k *KafkaConnectorProvider) dialerFor(cluster *contracts.KafkaCluster) (*kafka.Dialer, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if dialer, ok := k.dialers[cluster.Title]; ok {
		return dialer, nil
	}

	dialer, err := createDialer(cluster.Username, cluster.Password, cluster.TrustStorePath)
	if err != nil {
		return nil, err
	}

	k.dialers[cluster.Title] = dialer

	return dialer, nil
}

// writerFor возвращает writer для кластера и партиции, лениво создавая его.
// Партиция фиксируется балансером, потому что kafka.Writer не позволяет задать
// её через kafka.Message.Partition (это поле read-only при записи).
func (k *KafkaConnectorProvider) writerFor(cluster *contracts.KafkaCluster, partition int) (*kafka.Writer, error) {
	dialer, err := k.dialerFor(cluster)
	if err != nil {
		return nil, err
	}

	key := writerKey{clusterTitle: cluster.Title, partition: partition}

	k.mu.Lock()
	defer k.mu.Unlock()

	if writer, ok := k.writers[key]; ok {
		return writer, nil
	}

	writer := &kafka.Writer{
		Addr: kafka.TCP(cluster.Url),
		Balancer: kafka.BalancerFunc(func(_ kafka.Message, _ ...int) int {
			return partition
		}),
		Transport: &kafka.Transport{
			SASL: dialer.SASLMechanism,
			TLS:  dialer.TLS,
		},
	}

	k.writers[key] = writer

	return writer, nil
}

func (k *KafkaConnectorProvider) GetTopicsByClusterName(cluster *contracts.KafkaCluster) ([]string, error) {
	k.mu.Lock()
	cached, ok := k.topics[cluster.Title]
	k.mu.Unlock()

	if ok {
		return cached, nil
	}

	dialer, err := k.dialerFor(cluster)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", cluster.Url)
	if err != nil {
		slog.Error("Connection to Kafka error", "cluster", cluster.Title, "err", err)
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		slog.Error("Partitions in Kafka error", "cluster", cluster.Title, "err", err)
		return nil, err
	}

	topics := map[string]struct{}{}
	partitionMap := map[string][]int{}
	for _, p := range partitions {
		topics[p.Topic] = struct{}{}

		partitionMap[p.Topic] = append(partitionMap[p.Topic], p.ID)
	}

	for _, ids := range partitionMap {
		sort.Ints(ids)
	}

	topicsList := make([]string, 0, len(topics))
	for topic := range topics {
		topicsList = append(topicsList, topic)
	}

	sort.Strings(topicsList)

	k.mu.Lock()
	defer k.mu.Unlock()

	k.topics[cluster.Title] = topicsList
	k.partitions[cluster.Title] = partitionMap

	return topicsList, nil
}

func createDialer(username, password, certificateFilePath string) (*kafka.Dialer, error) {
	mechanism := plain.Mechanism{
		Username: username,
		Password: password,
	}

	cab, err := os.ReadFile(certificateFilePath)
	if err != nil {
		return nil, fmt.Errorf("read certificate file %q: %w", certificateFilePath, err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(cab); !ok {
		return nil, fmt.Errorf("certificate file %q does not contain valid PEM certificates", certificateFilePath)
	}

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

	return dialer, nil
}

func (k *KafkaConnectorProvider) GetPartitionsByClusterNameAndTopic(topicName string, cluster *contracts.KafkaCluster) ([]int, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	return k.partitions[cluster.Title][topicName], nil
}

func (k *KafkaConnectorProvider) Send(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, kafkaPartition string, textToSend string) error {
	partition, err := strconv.Atoi(kafkaPartition)
	if err != nil {
		slog.Error("kafka partition is not a number", "partition", kafkaPartition, "err", err)
		return err
	}

	writer, err := k.writerFor(kafkaCluster, partition)
	if err != nil {
		slog.Error("Connection is not initiated", "cluster", kafkaCluster.Title, "err", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	err = writer.WriteMessages(ctx, kafka.Message{
		Topic: kafkaTopic,
		Value: []byte(textToSend),
	})
	if err != nil {
		slog.Error("failed to write messages", "cluster", kafkaCluster.Title, "topic", kafkaTopic, "partition", partition, "err", err)
		return err
	}

	return nil
}

func (k *KafkaConnectorProvider) Read(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, kafkaPartition string, numberOfReadMessages int) ([]*contracts.ReadMessagesRow, error) {
	slog.Info("PETUSHIYI KRIK!")

	dialer, err := k.dialerFor(kafkaCluster)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	i, err := strconv.Atoi(kafkaPartition)
	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return nil, err
	}

	slog.Info("before DialLeader")

	conn, err := dialer.DialLeader(ctx, "tcp", kafkaCluster.Url, kafkaTopic, i)
	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	slog.Info("before lastOffset")

	// 2. Получаем последний доступный оффсет
	lastOffset, err := conn.ReadLastOffset()
	slog.Info("lastOffset", "lastOffset", lastOffset)

	lastOffset = max(lastOffset-1-int64(numberOfReadMessages), 0)

	slog.Info("lastOffset after - 10 - 1", "lastOffset", lastOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read last offset: %w", err)
	}

	// Если сообщений в партиции нет, последний оффсет будет -1
	// if lastOffset == 0 {
	// if lastOffset < 0 {
	// 	return nil, fmt.Errorf("no messages in partition")
	// }

	// 3. Создаем Reader для чтения с найденного оффсета
	// Важно: используем тот же контекст или новый для чтения
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaCluster.Url},
		Topic:     kafkaTopic,
		Dialer:    dialer,
		Partition: i,
	})
	defer func() {
		_ = reader.Close()
	}()

	// // Устанавливаем оффсет на последнее сообщение
	// err = reader.SetOffset(185)
	// err = reader.SetOffset(lastOffset - int64(numberOfReadMessages) - 1)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to set offset: %w", err)
	// }

	// 4. Читаем одно сообщение
	// Используем новый контекст для операции чтения
	readCtx, readCancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer readCancel()

	res := make([]*contracts.ReadMessagesRow, 0)

	for range numberOfReadMessages {
		slog.Info("Try to read message from kafka")
		message, err := reader.ReadMessage(readCtx)
		if err != nil {
			return nil, err
		}
		slog.Info("Read message from kafka", "message", message)

		res = append(res, &contracts.ReadMessagesRow{Row: []string{strconv.Itoa(int(message.Offset)), message.Time.String(), string(message.Key), string(message.Value)}})
	}

	return res, nil
}

// Close закрывает все открытые kafka-writer'ы. Регистрируется в closer при создании коннектора.
func (k *KafkaConnectorProvider) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	var errs []error

	for key, writer := range k.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close kafka writer (cluster=%s, partition=%d): %w", key.clusterTitle, key.partition, err))
		}

		delete(k.writers, key)
	}

	return errors.Join(errs...)
}
