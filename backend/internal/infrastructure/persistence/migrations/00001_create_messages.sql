-- +goose Up
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL
);

INSERT INTO messages (id, content) VALUES ('1', 'Hello World');

-- +goose Down
DROP TABLE IF EXISTS messages;
