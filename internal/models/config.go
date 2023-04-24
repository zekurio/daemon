package models

import "github.com/zekurio/daemon/internal/util/static"

var DefaultConfig = Config{
	Discord: DiscordConfig{
		Token:      "",
		OwnerID:    "",
		GuildLimit: -1,
	},
	Postgres: PostgresConfig{
		Host: "localhost",
		Port: 5432,
	},
	Permissions: PermissionRules{
		UserRules:  static.DefaultUserRules,
		AdminRules: static.DefaultAdminRules,
	},
}

type DiscordConfig struct {
	Token      string
	OwnerID    string
	GuildLimit int
}

type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type PermissionRules struct {
	UserRules  []string
	AdminRules []string
}

type Config struct {
	Discord     DiscordConfig
	Postgres    PostgresConfig
	Permissions PermissionRules
}
