package messagecommands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/services/codeexec"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util/static"
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

	exec := ctx.Get(static.DiCodeexec).(codeexec.ExecutorWrapper)
	db := ctx.Get(static.DiDatabase).(database.Database)

	ok, err := db.GetExecEnabled(ctx.GetEvent().GuildID)
	if err != nil {
		return
	}

	// check if jdoodle is enabled
	if !ok {
		err = ctx.FollowUpError("JDoodle is not enabled", "").Send().Error
		return
	}

	jdoodle, err := exec.NewExecutor(ctx.GetEvent().GuildID)
	if err != nil {
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

	p := codeexec.Payload{
		Language: lang,
		Code:     code,
	}

	// execute code
	res, err := jdoodle.Execute(p)
	if err != nil {
		err = ctx.FollowUpError("Failed to execute code", "").Send().Error
		return
	}

	if res.StdErr != "" {
		err = ctx.FollowUpError("An error occurred: "+res.StdErr, "").Send().Error
		return
	}

	// send result
	err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Title: "Result",
		// either use StdErr or StdOut, depending on if there was an error
		Description: fmt.Sprintf("```%s```", res.StdOut),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Memory",
				Value: res.MemUsed + " bytes",
			},
			{
				Name:  "CPU Time",
				Value: res.CPUTime + " seconds",
			},
			{
				Name:  "Execute time",
				Value: fmt.Sprintf("%d ms", res.ExecTime),
			},
		},
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
