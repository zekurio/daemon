package static

import "github.com/bwmarrin/discordgo"

const (
	ColorDefault = 0x7169ba
	ColorRed     = 0xff2b66
	ColorGreen   = 0x92f026
	ColorYellow  = 0xffff38
	ColorGray    = 0x929292

	OAuthScopes = "bot%20applications.commands"

	InvitePermission = discordgo.PermissionEmbedLinks |
		discordgo.PermissionManageRoles |
		discordgo.PermissionManageChannels |
		discordgo.PermissionVoiceMoveMembers

	Intents = discordgo.IntentsGuilds |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildEmojis |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildVoiceStates
)

var (
	DefaultAdminRules = []string{
		"+dm.guild.*",
		"+dm.etc.*",
		"+dm.chat.*",
	}

	DefaultUserRules = []string{
		"+dm.etc.*",
		"+dm.chat.*",
	}
)
