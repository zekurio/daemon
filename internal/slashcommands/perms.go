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
	"github.com/zekurio/daemon/pkg/perms"
	"github.com/zekurio/daemon/pkg/roleutils"
)

type Perms struct {
	ken.EphemeralCommand
}

const (
	permAllow = "+"
	permDeny  = "-"
)

var (
	_ ken.SlashCommand         = (*Perms)(nil)
	_ permissions.CommandPerms = (*Perms)(nil)
)

func (c *Perms) Name() string {
	return "perms"
}

func (c *Perms) Description() string {
	return "Set the permissions for groups on your guild."
}

func (c *Perms) Version() string {
	return "1.0.0"
}

func (c *Perms) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *Perms) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List the current permission definitions.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "set",
			Description: "Set a permission rule for specific roles.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "mode",
					Description: "Set the permission as allow or deny.",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "allow",
							Value: permAllow,
						},
						{
							Name:  "deny",
							Value: permDeny,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "perm",
					Description: "Permission Domain Name Specifier",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to apply the permission to.",
					Required:    true,
				},
			},
		},
	}
}

func (c *Perms) Perm() string {
	return "dm.guild.config.perms"
}

func (c *Perms) SubPerms() []permissions.SubCommandPerms {
	return nil
}

func (c *Perms) Run(ctx ken.Context) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "list", Run: c.list},
		ken.SubCommandHandler{Name: "set", Run: c.set},
	)

	return
}

func (c *Perms) list(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)
	s := ctx.Get(static.DiDiscord).(*discordgo.Session)

	gPerms, err := db.GetPermissions(ctx.GetEvent().GuildID)
	if err != nil && err != dberr.ErrNotFound {
		return
	}

	sortedGuildRoles, err := roleutils.GetSortedGuildRoles(s, ctx.GetEvent().GuildID, true)
	if err != nil {
		return err
	}

	if len(gPerms) == 0 {
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Title:       "Permissions",
			Description: "No permissions set.",
		}).Send().Error
	}

	msgstr := ""

	for _, role := range sortedGuildRoles {
		if pa, ok := gPerms[role.ID]; ok {
			msgstr += fmt.Sprintf("**<@&%s>**\n%s\n\n", role.ID, strings.Join(pa, "\n"))
		}
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Title:       "Permissions",
		Description: msgstr,
	}).Send().Error

}

func (c *Perms) set(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	mode := ctx.Options().GetByName("mode").StringValue()
	nPerm := ctx.Options().GetByName("perm").StringValue()
	role := ctx.Options().GetByName("role").RoleValue(ctx)

	nPerm = mode + nPerm

	gPerms, err := db.GetPermissions(ctx.GetEvent().GuildID)
	if err != nil {
		return err
	}

	cPerm, ok := gPerms[role.ID]
	if !ok {
		cPerm = make(perms.PermsArray, 0)
	}

	cPerm, changed := cPerm.Update(nPerm, false)
	if changed {
		err := db.SetPermissions(ctx.GetEvent().GuildID, role.ID, cPerm)
		if err != nil {
			return err
		}
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Permissions set",
		Description: fmt.Sprintf("Set permission `%s` for role `%s`", nPerm, role.Name),
	})

}
