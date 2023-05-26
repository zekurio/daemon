package models

type GuildMap map[string]*AVChannel

type AVChannel struct {
	GuildID          string
	OwnerID          string
	OriginChannelID  string
	CreatedChannelID string
}
