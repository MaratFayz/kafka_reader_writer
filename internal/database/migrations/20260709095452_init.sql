-- +goose Up
CREATE TABLE clusters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    username TEXT NOT NULL, 
    password TEXT NOT NULL,
    trust_store_path TEXT NOT NULL
); 

INSERT INTO clusters(title, url, username, password, trust_store_path) VALUES ('sfa', '172.16.15.171:9093', 'SFA', 'SFADEV123', './c/certificate.pem');

CREATE TABLE message_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT
);

-- +goose Down
DROP TABLE clusters;
DROP TABLE message_history;