-- +goose Up
CREATE TABLE clusters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE, 
    password TEXT NOT NULL,
    trust_store_path TEXT NOT NULL
); 

CREATE TABLE sent_messages_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    body TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('OK', 'ERROR')),
    date_time_sent TEXT NOT NULL
);

-- +goose Down
DROP TABLE clusters;
DROP TABLE sent_messages_history;