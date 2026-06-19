package connection

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/windows"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type KafkaConnectorProvider struct {
}

func NewKafkaConnectorProvider() *KafkaConnectorProvider {
	return &KafkaConnectorProvider{}
}

func (k *KafkaConnectorProvider) GetTopicsByClusterName(clusterName windows.KafkaCluster) []string {
	conn, err := createConnection(clusterName.Username(), clusterName.Password(), clusterName.TrustStorePath(), clusterName.Url())
	if err != nil {
		slog.Error("COnenction to Kafak error")
	}
	partitions, err := conn.ReadPartitions()
	if err != nil {
		slog.Error("Partitions in Kafka error")
	}

	// Собираем имена топиков
	m := map[string]struct{}{}
	for _, p := range partitions {
		if p.Error != nil {
			continue
		}

		if len(strings.TrimSpace(p.Topic)) > 0 {
			m[strings.TrimSpace(p.Topic)] = struct{}{}
		}
	}

	partitions2 := make([]string, 0, len(m))
	for k := range m {
		partitions2 = append(partitions2, k)
	}

	sort.Strings(partitions2)

	// slog.Info("Topics", "topics", partitions2)
	// log.Fatal("XXX")
	return partitions2
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
