package slashcommands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"

	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/database/dberr"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/arrayutils"
)

type Autorole struct {
	ken.EphemeralCommand
}

var (
	_ ken.SlashCommand         = (*Autorole)(nil)
	_ permissions.CommandPerms = (*Autorole)(nil)
)

func (c *Autorole) Name() string {
	return "autorole"
}

func (c *Autorole) Description() string {
	return "Configure the autorole current guild."
}

func (c *Autorole) Version() string {
	return "1.0.0"
}

func (c *Autorole) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *Autorole) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "Display the current autorole roles.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add a role to autorole.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The autorole to be set.",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "remove",
			Description: "Remove a role from autorole.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The autorole to be removed.",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "purge",
			Description: "Unset all autorole roles.",
		},
	}
}

func (c *Autorole) Perm() string {
	return "dm.guild.config.autorole"
}

func (c *Autorole) SubPerms() []permissions.SubCommandPerms {
	return nil
}

func (c *Autorole) Run(ctx ken.Context) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "list", Run: c.list},
		ken.SubCommandHandler{Name: "add", Run: c.add},
		ken.SubCommandHandler{Name: "remove", Run: c.remove},
		ken.SubCommandHandler{Name: "purge", Run: c.purge},
	)

	return
}

func (c *Autorole) list(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	autoroles, err := db.GetAutoRoles(ctx.GetEvent().GuildID)
	if err != nil && err != dberr.ErrNotFound {
		return err
	}

	if len(autoroles) == 0 {
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: "No autoroles are set.",
		}).Send().Error
	}

	var res strings.Builder
	for _, id := range autoroles {
		res.WriteString(fmt.Sprintf("- <@&%s>\n", id))
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "Currently following roles are set as autoroles:\n" + res.String(),
	}).Send().Error
}

func (c *Autorole) add(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	role := ctx.Options().Get(0).
		RoleValue(ctx)

	autoroles, err := db.GetAutoRoles(ctx.GetEvent().GuildID)
	if err != nil && err != dberr.ErrNotFound {
		return
	}

	if arrayutils.Contains(autoroles, role.ID) {
		err = ctx.FollowUpError("The given autorole is already assigned.", "").Send().Error
		return
	}

	if err = db.SetAutoRoles(ctx.GetEvent().GuildID, append(autoroles, role.ID)); err != nil {
		return
	}

	err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Color:       static.ColorGreen,
		Description: "Role was successfully added as autorole.",
	}).Send().Error

	return
}

func (c *Autorole) remove(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	role := ctx.Options().Get(0).
		RoleValue(ctx)

	autoroles, err := db.GetAutoRoles(ctx.GetEvent().GuildID)
	if err != nil && err != dberr.ErrNotFound {
		return
	}

	if !arrayutils.Contains(autoroles, role.ID) {
		err = ctx.FollowUpError("The given role is not assigned as autorole.", "").Send().Error
		return
	}

	autoroles = arrayutils.RemoveLazy(autoroles, role.ID)
	if err = db.SetAutoRoles(ctx.GetEvent().GuildID, autoroles); err != nil {
		return
	}

	err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Color:       static.ColorGreen,
		Description: "Role was successfully removed as autorole.",
	}).Send().Error

	return
}

func (c *Autorole) purge(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	if err = db.SetAutoRoles(ctx.GetEvent().GuildID, []string{}); err != nil && err != dberr.ErrNotFound {
		return
	}

	err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Color:       static.ColorGreen,
		Description: "All autoroles were successfully removed.",
	}).Send().Error

	return
}
