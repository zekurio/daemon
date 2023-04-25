-- +goose Up

CREATE TABLE IF NOT EXISTS votes (
    id SERIAL,
    json_data JSON NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS autovoice (
    id VARCHAR(25) NOT NULL DEFAULT '',
    json_data JSON NOT NULL,
    PRIMARY KEY (guild_id)
);

DROP COLUMN IF EXISTS created_av_ids FROM guilds;

-- +goose Down

DROP TABLE IF EXISTS votes;

DROP TABLE IF EXISTS autovoice;

ALTER TABLE guilds ADD COLUMN created_av_ids TEXT NOT NULL DEFAULT '';