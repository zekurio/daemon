package slashcommands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"

	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
)

const nRoleOptions = 5

type RoleSelection struct{}

var (
	_ ken.SlashCommand         = (*RoleSelection)(nil)
	_ permissions.CommandPerms = (*RoleSelection)(nil)
)

func (c *RoleSelection) Name() string {
	return "roleselection"
}

func (c *RoleSelection) Description() string {
	return "Create a role selector."
}

func (c *RoleSelection) Version() string {
	return "1.0.0"
}

func (c *RoleSelection) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *RoleSelection) Options() []*discordgo.ApplicationCommandOption {
	roleOptions := make([]*discordgo.ApplicationCommandOption, 0, nRoleOptions+1)

	for i := 0; i < nRoleOptions; i++ {
		roleOptions = append(roleOptions, &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionRole,
			Name:        fmt.Sprintf("role%d", i+1),
			Description: fmt.Sprintf("Role %d", i+1),
			Required:    i == 0,
		})
	}

	options := []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "new",
			Description: "Create a message with attached role select buttons.",
			Options: append([]*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "content",
				Description: "The content of the message.",
				Required:    true,
			}}, roleOptions...),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "append",
			Description: "Append a role select button to an existing message.",
			Options: append([]*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "id",
				Description: "The ID of the message.",
				Required:    true,
			}}, roleOptions...),
		},
	}

	return options
}

func (c *RoleSelection) Perm() string {
	return "dm.guild.mod.roleselect"
}

func (c *RoleSelection) SubPerms() []permissions.SubCommandPerms {
	return nil
}

func (c *RoleSelection) Run(ctx ken.Context) (err error) {
	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{
			Name: "new",
			Run:  c.new,
		},
		ken.SubCommandHandler{
			Name: "append",
			Run:  c.append,
		},
	)

	return err
}

func (c *RoleSelection) new(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		return err
	}

	content := ctx.Options().GetByName("content").StringValue()

	fum := ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: content,
	}).Send()

	b := fum.AddComponents()

	roleIDs, err := c.attachRoleButtons(ctx, b)
	if err != nil {
		return err
	}

	roleSelects := mapRoleSelects(ctx.GetEvent().GuildID, fum.ChannelID, fum.ID, roleIDs)

	db := ctx.Get(static.DiDatabase).(database.Database)
	return db.AddRoleSelections(roleSelects)

}

func (c *RoleSelection) append(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	if err := ctx.Defer(); err != nil {
		return err
	}

	id := ctx.Options().GetByName("id").StringValue()

	s := ctx.Get(static.DiDiscord).(*discordgo.Session)

	msg, err := discordutils.GetMessage(s, ctx.GetEvent().ChannelID, id)
	if err != nil {
		return ctx.FollowUpError("Message could not be found in this channel.", "").
			Send().Error
	}

	b := ctx.GetKen().Components().Add(msg.ID, msg.ChannelID)

	roleIDs, err := c.attachRoleButtons(ctx, b)
	if err != nil {
		return err
	}

	roleSelects := mapRoleSelects(ctx.GetEvent().GuildID, msg.ChannelID, msg.ID, roleIDs)

	db := ctx.Get(static.DiDatabase).(database.Database)
	err = db.AddRoleSelections(roleSelects)
	if err != nil {
		return err
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "Role buttons have been attached.",
	}).Send().DeleteAfter(6 * time.Second).Error
}

func (c *RoleSelection) attachRoleButtons(ctx ken.SubCommandContext, b *ken.ComponentBuilder) ([]string, error) {
	roles := make([]*discordgo.Role, 0, nRoleOptions)
	for i := 0; i < nRoleOptions; i++ {
		r, ok := ctx.Options().GetByNameOptional(fmt.Sprintf("role%d", i+1))
		if ok {
			roles = append(roles, r.RoleValue(ctx))
		}
	}

	return util.AttachRoleSelectButtons(b, roles)
}

func mapRoleSelects(guildID, channelID, msgID string, roleIDs []string) []models.RoleSelection {
	roleSelects := make([]models.RoleSelection, 0, len(roleIDs))
	for _, rid := range roleIDs {
		roleSelects = append(roleSelects, models.RoleSelection{
			GuildID:   guildID,
			ChannelID: channelID,
			MessageID: msgID,
			RoleID:    rid,
		})
	}
	return roleSelects
}
