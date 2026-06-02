package connection

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func CreateConnection() (Kafka, error) {
	mechanism := plain.Mechanism{
		Username: config.username,
		Password: config.password,
	}

	ca, err := os.Open(config.certificateFilePath)
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

	conn, err := dialer.Dial("tcp", config.broker)

	if err != nil {
		slog.Error("Connection is not initiated", "err", err)
		return nil, err
	}

	return conn, nil
}

type Kafka interface {
}

func sendMessage(config *Config, message string) {
	// conn, err := dialer.DialLeader(context.Background(), "tcp", config.broker, config.topic, config.partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	partitions, err := conn.ReadPartitions()
	if err != nil {
		panic(err.Error())
	}

	// Собираем имена топиков
	m := map[string]struct{}{}
	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}

	// Выводим имена топиков
	for k := range m {
		fmt.Println(k)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Topic: config.topic, Value: []byte(message)},
		// kafka.Message{Value: []byte("two!")},
		// kafka.Message{Value: []byte("three!")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
