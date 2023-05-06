package messagecommands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/jdoodle"
	"strings"
)

type Exec struct {
}

var (
	_ ken.MessageCommand       = (*Exec)(nil)
	_ permissions.CommandPerms = (*Exec)(nil)
)

func (c *Exec) Name() string {
	return "exec"
}

func (c *Exec) Description() string {
	return "Execute code using JDoodle"
}

func (c *Exec) Version() string {
	return "1.0.0"
}

func (c *Exec) Type() discordgo.ApplicationCommandType {
	return discordgo.MessageApplicationCommand
}

func (c *Exec) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (c *Exec) Perm() string {
	return "dm.chat.exec"
}

func (c *Exec) SubPerms() []permissions.SubCommandPerms {
	return nil
}

func (c *Exec) TypeMessage() {}

func (c *Exec) Run(ctx ken.Context) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	db := ctx.Get(static.DiDatabase).(database.Database)

	ok, err := db.GetExecEnabled(ctx.GetEvent().GuildID)
	if err != nil {
		return
	}

	// check if wrapper is enabled
	if !ok {
		err = ctx.FollowUpError("JDoodle is not enabled", "").Send().Error
		return
	}

	var content string

	if resolved := ctx.GetEvent().ApplicationCommandData().Resolved; resolved != nil {
		if resolved.Messages != nil {
			for _, msg := range resolved.Messages {
				content = msg.Content
				break
			}
		}
	}

	lang, code, ok := parseMessageContent(content)
	if !ok {
		err = ctx.FollowUpError("Invalid code block", "").Send().Error
		return
	}

	creds, err := db.GetJDoodleKey(ctx.GetEvent().GuildID)
	if err != nil {
		return
	}

	split := strings.Split(creds, "::")

	wrapper := jdoodle.New(split[0], split[1])

	output, err := wrapper.Execute(lang, code)
	if err != nil {
		err = ctx.FollowUpError("Error executing code", "").Send().Error
		return
	}

	err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Title:       "JDoodle Output",
		Description: "```\n" + output.Output + "\n```",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Memory",
				Value: output.Memory,
			},
			{
				Name:  "CPU Time",
				Value: output.CPUTime,
			},
			{
				Name:  "Status",
				Value: output.Status,
			},
		},
		Color: static.ColorCyan,
	}).Send().Error

	return
}

func parseMessageContent(content string) (lang string, script string, ok bool) {
	spl := strings.Split(content, "```")
	if len(spl) < 3 {
		return
	}

	inner := spl[1]
	iFirstLineBreak := strings.Index(inner, "\n")
	if iFirstLineBreak < 0 || len(inner)+1 <= iFirstLineBreak {
		return
	}

	lang = inner[:iFirstLineBreak]
	script = inner[iFirstLineBreak+1:]
	ok = len(lang) > 0 && len(script) > 0

	return
}
