-- +goose Up

DROP TABLE IF EXISTS roleselection;

-- +goose Down

CREATE TABLE IF NOT EXISTS roleselection (
    guild_id VARCHAR(25) NOT NULL DEFAULT '',
    channel_id VARCHAR(25) NOT NULL DEFAULT '',
    message_id VARCHAR(25) NOT NULL DEFAULT '',
    role_id VARCHAR(25) NOT NULL DEFAULT '',
    PRIMARY KEY (guild_id, channel_id, message_id, role_id)
);
