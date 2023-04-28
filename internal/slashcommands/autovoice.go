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

type Autovoice struct {
	ken.EphemeralCommand
}

var (
	_ ken.SlashCommand         = (*Autovoice)(nil)
	_ permissions.CommandPerms = (*Autovoice)(nil)
)

func (c *Autovoice) Name() string {
	return "autovoice"
}

func (c *Autovoice) Description() string {
	return "Configure the autovoice current guild."
}

func (c *Autovoice) Version() string {
	return "1.0.0"
}

func (c *Autovoice) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *Autovoice) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "Display the current autovoice channels.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add a channel to autovoice channels.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionChannel,
					Name:         "channel",
					Description:  "The autovoice to be set.",
					ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildVoice},
					Required:     true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "remove",
			Description: "Remove a channel from autovoice.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The autovoice to be removed.",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "purge",
			Description: "Unset all autovoice channels.",
		},
	}
}

func (c *Autovoice) Perm() string {
	return "dm.guild.config.autovoice"
}

func (c *Autovoice) SubPerms() []permissions.SubCommandPerms {
	return nil
}

func (c *Autovoice) Run(ctx ken.Context) (err error) {
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

func (c *Autovoice) list(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	autovoice, err := db.GetAutoVoice(ctx.GetEvent().GuildID)
	if err != nil && err != dberr.ErrNotFound {
		return err
	}

	if len(autovoice) == 0 {
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: "No autovoice are set.",
		}).Send().Error
	}

	var res strings.Builder
	for _, id := range autovoice {
		res.WriteString(fmt.Sprintf("- <#%s>\n", id))
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "Currently following channels are set as autovoice:\n" + res.String(),
	}).Send().Error

}

func (c *Autovoice) add(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	channel := ctx.Options().Get(0).
		ChannelValue(ctx)

	if channel.Type != discordgo.ChannelTypeGuildVoice {
		return ctx.FollowUpError("The given channel is not a voice channel.", "Argument Error").Send().Error
	}

	autovoice, err := db.GetAutoVoice(ctx.GetEvent().GuildID)
	if err != nil && err != dberr.ErrNotFound {
		return ctx.FollowUpError("An error occurred while fetching autovoice channels.", "Database Error").Send().Error
	}

	if arrayutils.Contains(autovoice, channel.ID) {
		return ctx.FollowUpError("The given autovoice is already assigned.", "").Send().Error
	}

	if err = db.SetAutoVoice(ctx.GetEvent().GuildID, append(autovoice, channel.ID)); err != nil {
		return
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Color:       static.ColorGreen,
		Description: "Channel was successfully added as autovoice.",
	}).Send().Error

}

func (c *Autovoice) remove(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	channel := ctx.Options().Get(0).
		ChannelValue(ctx)

	autovoice, err := db.GetAutoVoice(ctx.GetEvent().GuildID)
	if err != nil && err != dberr.ErrNotFound {
		return
	}

	if !arrayutils.Contains(autovoice, channel.ID) {
		err = ctx.FollowUpError("The given channel is not assigned as autovoice.", "").Send().Error
		return
	}

	autovoice = arrayutils.RemoveLazy(autovoice, channel.ID)
	if err = db.SetAutoVoice(ctx.GetEvent().GuildID, autovoice); err != nil {
		return
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Color:       static.ColorGreen,
		Description: "Channel was successfully removed as autovoice.",
	}).Send().Error

}

func (c *Autovoice) purge(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	if err = db.SetAutoVoice(ctx.GetEvent().GuildID, []string{}); err != nil && err != dberr.ErrNotFound {
		return
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Color:       static.ColorGreen,
		Description: "All autovoice were successfully removed.",
	}).Send().Error

}
