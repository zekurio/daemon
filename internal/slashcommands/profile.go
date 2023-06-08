package slashcommands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"

	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
	"github.com/zekurio/daemon/pkg/quickembed"
	"github.com/zekurio/daemon/pkg/stringutils"
)

type Profile struct {
	ken.EphemeralCommand
}

var (
	_ ken.SlashCommand         = (*Profile)(nil)
	_ permissions.CommandPerms = (*Profile)(nil)
)

func (c *Profile) Name() string {
	return "profile"
}

func (c *Profile) Description() string {
	return "Shows the profile of a user."
}

func (c *Profile) Version() string {
	return "1.0.0"
}

func (c *Profile) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *Profile) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to be displayed.",
		},
	}
}

func (c *Profile) Perm() string {
	return "dm.chat.profile"
}

func (c *Profile) SubPerms() []permissions.SubCommandPerms {
	return nil
}

func (c *Profile) Run(ctx ken.Context) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	s := ctx.Get(static.DiDiscord).(*discordgo.Session)
	p := ctx.Get(static.DiPermissions).(*permissions.Permissions)

	userID := ctx.GetEvent().Member.User.ID
	if userV, ok := ctx.Options().GetByNameOptional("user"); ok {
		user := userV.UserValue(ctx)
		userID = user.ID
	}

	member, err := discordutils.GetMember(s, ctx.GetEvent().GuildID, userID)
	if err != nil {
		return
	}

	guild, err := discordutils.GetGuild(s, ctx.GetEvent().GuildID)
	if err != nil {
		return
	}

	membRoleIDs := make(map[string]struct{})
	for _, rID := range member.Roles {
		membRoleIDs[rID] = struct{}{}
	}

	maxPos := len(guild.Roles)
	roleColor := static.ColorGrey
	for _, guildRole := range guild.Roles {
		if _, ok := membRoleIDs[guildRole.ID]; ok && guildRole.Position < maxPos && guildRole.Color != 0 {
			maxPos = guildRole.Position
			roleColor = guildRole.Color
		}
	}

	createdTime, err := discordutils.GetDiscordSnowflakeCreationTime(member.User.ID)
	if err != nil {
		return
	}

	perms, _, err := p.GetPerms(s, ctx.GetEvent().GuildID, member.User.ID)
	if err != nil {
		return
	}

	roles := make([]string, len(member.Roles))
	for i, rID := range member.Roles {
		roles[i] = "<@&" + rID + ">"
	}

	emb := quickembed.New().
		SetTitle("Profile of "+member.User.Username).
		SetThumbnail(member.User.AvatarURL("256"), "", 100, 100).
		SetColor(roleColor).
		AddField("Nickname", member.Nick).
		AddField("ID", fmt.Sprintf("`%s`", member.User.ID)).
		AddField("Joined at", stringutils.EnsureNotEmpty(member.JoinedAt.Format("02.01.2006, 15:04"),
			"*failed parsing timestamp*")).
		AddField("Created at", stringutils.EnsureNotEmpty(createdTime.Format("02.01.2006, 15:04"),
			"*failed parsing timestamp*")).
		AddField("Bot Permissions", stringutils.EnsureNotEmpty(strings.Join(perms, "\n"), "*no perms set*")).
		AddField("Roles", stringutils.EnsureNotEmpty(strings.Join(roles, ", "), "*no roles set*"))

	if member.User.Bot {
		emb.SetDescription(":robot:  **this is a bot account**")
	}

	return ctx.FollowUpEmbed(emb.Build()).Send().Error
}
