-- +goose Up

CREATE TABLE IF NOT EXISTS votes (
    id SERIAL,
    json_data JSON NOT NULL,
    PRIMARY KEY (id)
);

-- +goose Down

DROP TABLE IF EXISTS votes;