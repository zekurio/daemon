-- +goose Up

CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    data jsonb NOT NULL DEFAULT '{}',
    PRIMARY KEY (id)
);

-- +goose Down

DROP TABLE IF EXISTS votes;