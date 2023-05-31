package static

import "github.com/bwmarrin/discordgo"

const (
	ColorRed     = 0xEA8A8A
	ColorDefault = 0x7091FF
	ColorGrey    = 0xC2D2D9
	ColorOrange  = 0xFAC47F
	ColorGreen   = 0xAED785
	ColorCyan    = 0x4FDCE7
	ColorYellow  = 0xFFDC79
	ColorViolet  = 0x9A7BB7

	OAuthScopes = "bot%20applications.commands"

	InvitePermission = discordgo.PermissionEmbedLinks |
		discordgo.PermissionManageRoles |
		discordgo.PermissionManageChannels |
		discordgo.PermissionVoiceMoveMembers

	Intents = discordgo.IntentsGuilds |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildEmojis |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsDirectMessages
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
