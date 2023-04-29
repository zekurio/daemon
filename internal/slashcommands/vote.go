package slashcommands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/internal/util/vote"
	"github.com/zekurio/daemon/pkg/timeutils"
)

type Vote struct{}

var (
	_ ken.SlashCommand         = (*Vote)(nil)
	_ permissions.CommandPerms = (*Vote)(nil)
)

func (c *Vote) Name() string {
	return "vote"
}

func (c *Vote) Description() string {
	return "Create a vote."
}

func (c *Vote) Version() string {
	return "1.0.0"
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
	db := ctx.Get(static.DiDatabase).(database.Database)

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
				"Choices can not be empty.", "").
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

	ivote := vote.Vote{
		ID:          ctx.GetEvent().ID,
		CreatorID:   ctx.User().ID,
		GuildID:     ctx.GetEvent().GuildID,
		ChannelID:   ctx.GetEvent().ChannelID,
		Description: body,
		Choices:     split,
		ImageURL:    imgLink,
		Expires:     expires,
		Buttons:     map[string]vote.ChoiceButton{},
		CurrentVote: map[string]vote.CurrentVote{},
	}

	emb, err := ivote.AsEmbed(ctx.GetSession())
	if err != nil {
		return err
	}

	fum := ctx.FollowUpEmbed(emb).Send()
	err = fum.Error
	if err != nil {
		return
	}

	b := fum.AddComponents()

	ivote.MsgID = fum.Message.ID
	_, err = ivote.AddButtons(b)
	if err != nil {
		return err
	}

	err = db.AddUpdateVote(ivote)
	if err != nil {
		return err
	}

	vote.VotesRunning[ivote.ID] = ivote
	return
}

func (c *Vote) list(ctx ken.SubCommandContext) (err error) {
	emb := &discordgo.MessageEmbed{
		Description: "Your open votes on this guild:",
		Color:       static.ColorDefault,
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}
	for _, v := range vote.VotesRunning {
		if v.GuildID == ctx.GetEvent().GuildID && v.CreatorID == ctx.User().ID {
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
	db, _ := ctx.Get(static.DiDatabase).(database.Database)

	expireDuration, err := timeutils.ParseDuration(ctx.Options().GetByName("timeout").StringValue())
	if err != nil {
		return ctx.FollowUpError(
			"Invalid duration format. Please take a look "+
				"[here](https://golang.org/pkg/time/#ParseDuration) how to format duration parameter.", "").
			Send().Error
	}

	id := ctx.Options().Get(0).StringValue()
	var ivote *vote.Vote
	for _, v := range vote.VotesRunning {
		if v.GuildID == ctx.GetEvent().GuildID && v.ID == id {
			ivote = &v
		}
	}

	if err = ivote.SetExpire(ctx.GetSession(), expireDuration); err != nil {
		return err
	}
	if err = db.AddUpdateVote(*ivote); err != nil {
		return err
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: fmt.Sprintf("Vote will expire <t:%d:R>", ivote.Expires.Unix()),
	}).Send().Error
}

func (c *Vote) close(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)

	state := vote.StateClosed

	if showChartV, ok := ctx.Options().GetByNameOptional("chart"); ok && !showChartV.BoolValue() {
		state = vote.StateClosedNC
	}

	id := ctx.Options().GetByName("id").StringValue()

	if strings.ToLower(id) == "all" {
		var i int
		for _, v := range vote.VotesRunning {
			if v.GuildID == ctx.GetEvent().GuildID && v.CreatorID == ctx.User().ID {
				go func(vC vote.Vote) {
					err := db.DeleteVote(vC.ID)
					if err != nil {
						return
					}
					err = vC.Close(ctx.GetSession(), state)
					if err != nil {
						return
					}
				}(v)
				i++
			}
		}
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: fmt.Sprintf("Closed %d votes.", i),
		}).Send().Error
	}

	var ivote *vote.Vote
	for _, v := range vote.VotesRunning {
		if v.GuildID == ctx.GetEvent().GuildID && v.ID == id {
			ivote = &v
			break
		}
	}

	p := ctx.Get(static.DiPermissions).(*permissions.Permissions)
	ok, override, err := p.HasPerms(ctx.GetSession(), ctx.GetEvent().GuildID, ctx.User().ID, "!"+ctx.GetCommand().(permissions.CommandPerms).Perm()+".close")
	if err != nil {
		return err
	}

	if ivote.CreatorID != ctx.User().ID && !ok && !override {
		return ctx.FollowUpError(
			"You do not have the permission to close another ones votes.", "").
			Send().DeleteAfter(5 * time.Second).Error
	}

	err = db.DeleteVote(ivote.ID)
	if err != nil {
		return err
	}

	if err = ivote.Close(ctx.GetSession(), state); err != nil {
		return
	}

	err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "Vote closed.",
	}).Send().DeleteAfter(5 * time.Second).Error
	return
}
