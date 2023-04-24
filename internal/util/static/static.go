package static

import "github.com/bwmarrin/discordgo"

const (
	ColorError   = 0xd32f2f
	ColorDefault = 0xffc107
	ColorUpdated = 0x8bc34a
	ColorGrey    = 0xb0bec5
	ColorOrange  = 0xfb8c00
	ColorGreen   = 0x8BC34A
	ColorCyan    = 0x00BCD4
	ColorYellow  = 0xFFC107
	ColorViolet  = 0x6A1B9A

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
