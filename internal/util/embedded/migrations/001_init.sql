-- +goose Up

CREATE TABLE IF NOT EXISTS guilds (
	guild_id VARCHAR(25) NOT NULL DEFAULT '',
	autorole_ids text NOT NULL DEFAULT '',
	autovoice_ids text NOT NULL DEFAULT '',
	created_av_ids TEXT NOT NULL DEFAULT '',
	PRIMARY KEY (guild_id)
);

CREATE TABLE IF NOT EXISTS permissions (
	role_id VARCHAR(25) NOT NULL DEFAULT '',
	guild_id VARCHAR(25) NOT NULL DEFAULT '',
	perms text NOT NULL DEFAULT '',
	PRIMARY KEY (role_id)
);

CREATE TABLE IF NOT EXISTS roleselection (
	guild_id VARCHAR(25) NOT NULL DEFAULT '',
	channel_id VARCHAR(25) NOT NULL DEFAULT '',
	message_id VARCHAR(25) NOT NULL DEFAULT '',
	role_id VARCHAR(25) NOT NULL DEFAULT '',
	PRIMARY KEY (guild_id, channel_id, message_id, role_id)
);

-- +goose Down
DROP TABLE IF EXISTS guilds;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roleselection;