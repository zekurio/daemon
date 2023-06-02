-- +goose Up

CREATE TABLE IF NOT EXISTS guilds
(
    guild_id      VARCHAR(25) NOT NULL DEFAULT '',
    autorole_ids  text        NOT NULL DEFAULT '',
    autovoice_ids text        NOT NULL DEFAULT '',
    PRIMARY KEY (guild_id)
);

CREATE TABLE IF NOT EXISTS permissions
(
    role_id  VARCHAR(25) NOT NULL DEFAULT '',
    guild_id VARCHAR(25) NOT NULL DEFAULT '',
    perms    text        NOT NULL DEFAULT '',
    PRIMARY KEY (role_id)
);

CREATE TABLE IF NOT EXISTS votes
(
    id        VARCHAR(25) NOT NULL DEFAULT '',
    json_data TEXT        NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS autovoice
(
    user_id        VARCHAR(25) NOT NULL DEFAULT '',
    json_data TEXT        NOT NULL,
    PRIMARY KEY (user_id)
);

-- +goose Down
DROP TABLE IF EXISTS guilds;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS autovoice;