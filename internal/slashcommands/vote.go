package slashcommands

import (
	"fmt"
	"strings"
	"time"

	"github.com/zekurio/daemon/internal/middlewares"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/timeutils"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/services/vote"
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
					Name:        "options",
					Description: "The options - split by `,`.",
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
	vh := ctx.Get(static.DiVotes).(vote.VotesProvider)

	body := ctx.Options().GetByName("body").StringValue()
	choices := ctx.Options().GetByName("choices").StringValue()
	split := strings.Split(choices, ",")
	if len(split) < 2 || len(split) > 10 {
		return ctx.FollowUpError(
			"Invalid arguments. Please use `help vote` go get help about how to use this command.", "").
			Send().Error
	}
	for i, e := range split {
		if len(e) < 1 {
			return ctx.FollowUpError(
				"Possibilities can not be empty.", "").
				Send().Error
		}
		split[i] = strings.Trim(e, " \t")
	}

	var imgLink string
	if imgLinkV, ok := ctx.Options().GetByNameOptional("imageurl"); ok {
		imgLink = imgLinkV.StringValue()
	}

	var expires time.Time
	if expiresV, ok := ctx.Options().GetByNameOptional("timeout"); ok {
		expiresDuration, err := timeutils.ParseDuration(expiresV.StringValue())
		if err != nil {
			return ctx.FollowUpError(
				"Invalid duration format. Please take a look "+
					"[here](https://golang.org/pkg/time/#ParseDuration) how to format duration parameter.", "").
				Send().Error
		}
		expires = time.Now().Add(expiresDuration)
	}

	v, err := vh.CreateVote(ctx, body, imgLink, split, expires)
	if err != nil {
		return ctx.FollowUpError(
			"Failed to create vote.", "").
			Send().Error
	}

	emb, err := v.AsEmbed(ctx.GetSession())
	if err != nil {
		return err
	}

	fum := ctx.FollowUpEmbed(emb).Send()
	err = fum.Error
	if err != nil {
		return
	}

	b := fum.AddComponents()

	v.MsgID = fum.Message.ID
	_, err = v.AddButtons(b)
	if err != nil {
		return err
	}

	return
}

func (c *Vote) list(ctx ken.SubCommandContext) (err error) {
	vh := ctx.Get(static.DiVotes).(vote.VotesProvider)

	emb := &discordgo.MessageEmbed{
		Description: "Your open votes on this guild:",
		Color:       static.ColorDefault,
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	votes, err := vh.GetVotes()
	if err != nil {
		return err
	}

	for _, v := range votes {
		if v.GuildID == ctx.GetEvent().GuildID {
			emb.Fields = append(emb.Fields, v.AsField())
		}
	}

	if len(emb.Fields) == 0 {
		emb.Description = "You don't have any open votes on this guild."
	}
	err = ctx.FollowUpEmbed(emb).Send().Error
	return err
}

func (c *Vote) expire(ctx ken.SubCommandContext) (err error) {
	vh := ctx.Get(static.DiVotes).(vote.VotesProvider)
	db := ctx.Get(static.DiDatabase).(database.Database)

	expireDuration, err := timeutils.ParseDuration(ctx.Options().GetByName("timeout").StringValue())
	if err != nil {
		return ctx.FollowUpError(
			"Invalid duration format. Please take a look "+
				"[here](https://golang.org/pkg/time/#ParseDuration) how to format duration parameter.", "").
			Send().Error
	}

	id := ctx.Options().Get(0).StringValue()

	if id == "all" {
		return c.expireAllVotes(ctx, vh, db, expireDuration)
	}

	return c.expireSingleVote(ctx, vh, db, id, expireDuration)
}

func (c *Vote) expireAllVotes(ctx ken.SubCommandContext, vh vote.VotesProvider, db database.Database, expireDuration time.Duration) (err error) {
	votes, err := vh.GetVotes()
	if err != nil {
		return err
	}
	for _, v := range votes {
		if v.GuildID == ctx.GetEvent().GuildID {
			err := c.expireVote(ctx, db, &v, expireDuration)
			if err != nil {
				return err
			}
		}
	}
	return ctx.FollowUpError("No vote found.", "").Send().Error
}

func (c *Vote) expireSingleVote(ctx ken.SubCommandContext, vh vote.VotesProvider, db database.Database, id string, expireDuration time.Duration) (err error) {
	ivote, err := vh.GetVote(id)
	if err != nil {
		return err
	}
	return c.expireVote(ctx, db, ivote, expireDuration)
}

func (c *Vote) expireVote(ctx ken.SubCommandContext, db database.Database, ivote *models.Vote, expireDuration time.Duration) (err error) {
	if err := ivote.SetExpire(ctx.GetSession(), expireDuration); err != nil {
		return err
	}
	if err := db.AddUpdateVote(*ivote); err != nil {
		return err
	}
	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: fmt.Sprintf("Vote will expire <t:%d:R>", ivote.Expires.Unix()),
	}).Send().Error
}

func (c *Vote) close(ctx ken.SubCommandContext) (err error) {
	vh := ctx.Get(static.DiVotes).(vote.VotesProvider)

	id := ctx.Options().GetByName("id").StringValue()

	if strings.ToLower(id) == "all" {
		var i int
		votes, err := vh.GetVotes()
		if err != nil {
			return err
		}

		for _, v := range votes {
			if v.GuildID == ctx.GetEvent().GuildID {
				if err := vh.DeleteVote(ctx.GetSession(), v.ID); err != nil {
					return err
				}

				i++
			}

			return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
				Description: fmt.Sprintf("Closed %d votes.", i),
			}).Send().Error
		}

		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: fmt.Sprintf("Closed %d votes.", i),
		}).Send().Error
	}

	v, err := vh.GetVote(id)
	if err != nil {
		return err
	}

	if v.GuildID != ctx.GetEvent().GuildID {
		return ctx.FollowUpError("Vote not found.", "").Send().Error
	}

	if err := vh.DeleteVote(ctx.GetSession(), v.ID); err != nil {
		return err
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: fmt.Sprintf("Closed vote %s.", v.ID),
	}).Send().Error
}
