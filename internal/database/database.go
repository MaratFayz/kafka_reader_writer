package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Db struct {
	conn *sql.DB
	mu   sync.Mutex
}

type kafkaClusterDb struct {
	title              string
	url                string
	trustStorePath     string
	trustStorePassword string
	username           string
	password           string
	saslMechanism      string
}

func (kc *kafkaClusterDb) Title() string {
	return kc.title
}

func (kc *kafkaClusterDb) Url() string {
	return kc.url
}
func (kc *kafkaClusterDb) TrustStorePath() string {
	return kc.trustStorePath
}
func (kc *kafkaClusterDb) TrustStorePassword() string {
	return kc.trustStorePassword
}
func (kc *kafkaClusterDb) Username() string {
	return kc.username
}
func (kc *kafkaClusterDb) Password() string {
	return kc.password
}
func (kc *kafkaClusterDb) SaslMechanism() string {
	return kc.saslMechanism
}

func CreateOrGetDb(databaseUrl string) (*Db, error) {
	//TODO delete later
	filePath := "database.db"
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("Error deleting file:", err)
	} else {
		fmt.Println("File successfully deleted")
	}

	conn, err := sql.Open("sqlite3", databaseUrl)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(conn, "migrations"); err != nil {
		panic(err)
	}

	fmt.Println("Миграции успешно применены!")

	return &Db{conn: conn}, nil
}

// Close закрывает подключение к базе данных.
func (d *Db) Close() error {
	err := d.conn.Close()
	if err != nil {
		slog.Error("подключение к базе данных было закрыто с ошибкой", "err", err.Error())
	} else {
		slog.Info("подключение к базе данных закрыто")
	}

	return nil
}

type kafkaClusters struct {
	clusters []*kafkaClusterDb
}

func (k *kafkaClusters) Clusters() []*kafkaClusterDb {
	return k.clusters
}

func (d *Db) GetKafkaClusters() []*contracts.KafkaCluster {
	clusters := make([]*contracts.KafkaCluster, 0, 1)
	// clusters = append(clusters, &contracts.KafkaCluster{Title: "sfa", Url: "172.16.15.171:9093", Username: "SFA", Password: "SFADEV123", TrustStorePath: "./c/certificate.pem"})

	d.mu.Lock()
	defer d.mu.Unlock()

	rows, err := d.conn.Query("select title, url,trust_store_path,username,password from clusters")

	if err != nil {
		slog.Error("Ошибка при получении кластеров из БД", "err", err)
	}

	for rows.Next() {
		var title string
		var url string
		var trust_store_path string
		var username string
		var password string

		err = rows.Scan(&title, &url, &trust_store_path, &username, &password)

		if err != nil {
			slog.Error("Ошибка при получении кластеров из БД 2", "err", err)
		}

		clusters = append(clusters, &contracts.KafkaCluster{Title: title, Url: url, TrustStorePath: trust_store_path, Username: username, Password: password})
	}

	// if err != nil {
	// 	slog.Error("Ошибка при получении существования файла из БД", "err", err)
	// 	return false, ExistFileOrNotError
	// } else {
	// 	return true, nil
	// }

	return clusters
}

func (d *Db) SaveMessage(kafkaCluster *contracts.KafkaCluster, kafkaTopic string, messageToSave string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	res, err := d.conn.Exec(`insert into sent_messages_history(cluster_id, body, status, date_time_sent) 
	values ((select c.id from clusters c where c.url = $1 and c.username = $2 and c.title = $3), $4, 'OK', $5)`,
		kafkaCluster.Url, kafkaCluster.Username, kafkaCluster.Title, messageToSave, time.Now())

	if err != nil {
		slog.Error("Ошибка при добавлении строки истории отправленных сообщений в БД", "err", err)
		return err
	}

	r, err := res.RowsAffected()
	if err != nil {
		slog.Error("Ошибка при добавлении строки истории отправленных сообщений в БД 2", "err", err)
		return err
	}

	if r == 0 {
		slog.Error("Ошибка при добавлении строки истории отправленных сообщений в БД 3", "r", r)
		return err
	}

	return nil
}
func (d *Db) GetMessages(kafkaCluster *contracts.KafkaCluster, topic string) ([]*contracts.SentMessagesRow, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	messages := make([]*contracts.SentMessagesRow, 0)

	rows, err := d.conn.Query(`select body, status, date_time_sent from sent_messages_history
	where cluster_id = (select c.id from clusters c where c.url = $1 and c.username = $2 and c.title = $3)
	limit 100`, kafkaCluster.Url, kafkaCluster.Username, kafkaCluster.Title)

	if err != nil {
		slog.Error("Ошибка при получении строк истории отправленных сообщений из БД", "err", err)
		return nil, err
	}

	for rows.Next() {
		var body string
		var status string
		var dateTimeSent string

		err = rows.Scan(&body, &status, &dateTimeSent)

		if err != nil {
			slog.Error("Ошибка при получении строк истории отправленных сообщений из БД 2", "err", err)
			return nil, err
		}

		messages = append(messages, &contracts.SentMessagesRow{Row: []string{body, status, dateTimeSent}})
	}

	return messages, nil
}
