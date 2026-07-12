package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/contracts"
	"os"
	"sync"

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
