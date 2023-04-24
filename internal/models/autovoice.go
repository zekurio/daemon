package models

import "github.com/bwmarrin/discordgo"

type AVChannel struct {
	OriginChannel  *discordgo.Channel
	CreatedChannel *discordgo.Channel
}
