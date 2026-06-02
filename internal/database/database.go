package database

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type db struct {
	conn *sql.DB
	mu   sync.Mutex
}

type DB interface {
	Close() error
}

func CreateOrGetDb(databaseUrl string) (DB, error) {
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

	//TODO delete later
	dirPath := "/Users/mfayz/Desktop/filesToSynchronize"

	// Get the file info structure for the directory
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	// .ModTime() returns the last modification time of the directory's metadata
	modTime := dirInfo.ModTime()

	sqlStmt := `
	create table if not exists epoch_number (finished_epoch_number integer not null default 0);
	INSERT into epoch_number(finished_epoch_number) values (0);
	create table if not exists watched_dirs (id integer primary key autoincrement, path varchar(1024) not null, status varchar(7) NOT NULL, date_time_edited TEXT);
	create table if not exists files_local (id integer primary key autoincrement, dir integer, name varchar(255), FOREIGN KEY(dir) REFERENCES watched_dirs(id));

	INSERT INTO watched_dirs(path, status, date_time_edited) VALUES ("/Users/mfayz/Desktop/filesToSynchronize", 'ACTIVE', $1);
	`
	//status = ENUM('ACTIVE', 'STOPPED', 'DELETED')

	_, err = conn.Exec(sqlStmt, modTime)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	return &db{conn: conn}, nil
}

// Close закрывает подключение к базе данных.
func (d *db) Close() error {
	err := d.conn.Close()
	if err != nil {
		slog.Error("подключение к базе данных было закрыто с ошибкой", "err", err.Error())
	} else {
		slog.Info("подключение к базе данных закрыто")
	}

	return nil
}
