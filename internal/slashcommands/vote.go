package slashcommands

import (
	"github.com/zekurio/daemon/internal/middlewares"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/services/permissions"
)

type Vote struct{}

var (
	_ ken.SlashCommand            = (*Vote)(nil)
	_ permissions.CommandPerms    = (*Vote)(nil)
	_ middlewares.CommandCooldown = (*Vote)(nil)
)

func (c *Vote) Name() string {
	return "vote"
}

func (c *Vote) Description() string {
	return "Create a vote."
}

func (c *Vote) Version() string {
	return "1.1.0"
}

func (c *Vote) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *Vote) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "Create a new vote.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "body",
					Description: "The vote body content.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "choices",
					Description: "The choices - split by `,`.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "imageurl",
					Description: "An optional image URL.",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timeout",
					Description: "Timeout of the vote (i.e. `1h`, `30m`, ...)",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List currently running votes.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "expire",
			Description: "Set the expiration of a running vote.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "The ID of the vote or `all` if you want to close all.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timeout",
					Description: "Timeout of the vote (i.e. `1h`, `30m`, ...)",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "close",
			Description: "Close a running vote.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "The ID of the vote or `all` if you want to close all.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "chart",
					Description: "Display chart (default `true`).",
				},
			},
		},
	}
}

func (c *Vote) Perm() string {
	return "dm.chat.vote"
}

func (c *Vote) SubPerms() []permissions.SubCommandPerms {
	return []permissions.SubCommandPerms{
		{
			Perm:        "close",
			Explicit:    true,
			Description: "Allows closing votes of other users.",
		},
	}
}

func (c *Vote) Cooldown() int {
	return 120
}

func (c *Vote) Run(ctx ken.Context) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "create", Run: c.create},
		ken.SubCommandHandler{Name: "list", Run: c.list},
		ken.SubCommandHandler{Name: "expire", Run: c.expire},
		ken.SubCommandHandler{Name: "close", Run: c.close},
	)

	return
}

func (c *Vote) create(ctx ken.SubCommandContext) (err error) {
	return
}

func (c *Vote) list(ctx ken.SubCommandContext) (err error) {
	return
}

func (c *Vote) expire(ctx ken.SubCommandContext) (err error) {
	return
}

func (c *Vote) close(ctx ken.SubCommandContext) (err error) {
	return
}
