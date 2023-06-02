package models

type ChannelMap map[string]AVChannel

type AVChannel struct {
	GuildID          string
	OwnerID          string
	OriginChannelID  string
	CreatedChannelID string
	Members          []string
}
