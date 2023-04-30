package slashcommands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
	"strings"
)

type ExecSetup struct {
	ken.EphemeralCommand
}

var (
	_ ken.SlashCommand         = (*ExecSetup)(nil)
	_ permissions.CommandPerms = (*ExecSetup)(nil)
)

func (c *ExecSetup) Name() string {
	return "exec-setup"
}

func (c *ExecSetup) Description() string {
	return "Setup code execution with JDoodle"
}

func (c *ExecSetup) Version() string {
	return "1.0.0"
}

func (c *ExecSetup) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *ExecSetup) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setup",
			Description: "Setup code execution.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "reset",
			Description: "Disable code execution and remove stored credentials.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "toogle",
			Description: "Toggle code execution on or off.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "check",
			Description: "Show the status of the current code execution setup.",
		},
	}
}

func (c *ExecSetup) Perm() string {
	return "dm.guild.config.exec"
}

func (c *ExecSetup) SubPerms() []permissions.SubCommandPerms {
	return nil
}

func (c *ExecSetup) Run(ctx ken.Context) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	err = ctx.HandleSubCommands()

	return
}

func (c *ExecSetup) setup(ctx ken.SubCommandContext) (err error) {

	dmMsg, err := discordutils.SendEmbedMessageDM(ctx.GetSession(), ctx.GetEvent().User.ID, &discordgo.MessageEmbed{
		Title: "JDoodle Code Execution Setup",
		Description: "Please enter your JDoodle Client ID and Client Secret. You can get them [here](https://www.jdoodle.com/compiler-api)\n" +
			"The values will be stored in plain text. Please enter your credentials in the following format:\n" +
			"`<client_id>::<client_secret>` or type `cancel` to cancel the setup.",
		Color: static.ColorYellow,
	})
	if err != nil {
		return ctx.FollowUpError("Failed sending setup message, make sure DMs are enabled, or contact the bot owner.", "").Send().Error
	}

	err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "Since you are setting up an API, you need to go to DMs to enter your credentials.",
	}).Send().Error

	var removeHandler func()
	var jdoodleCreds string
	removeHandler = ctx.GetSession().AddHandler(func(s *discordgo.Session, e *discordgo.MessageCreate) {
		self, err := discordutils.GetUser(s, s.State.User.ID)
		if err != nil {
			return
		}

		if e.ChannelID != dmMsg.ChannelID || e.Author.ID == self.ID {
			return
		}

		if strings.ToLower(e.Content) == "cancel" {
			_, err = discordutils.SendEmbedMessageDM(ctx.GetSession(), ctx.GetEvent().User.ID, &discordgo.MessageEmbed{
				Description: "Setup cancelled.",
			})
		}

		jdoodleCreds = e.Content
		// TODO: Check if credentials are valid
		db := ctx.Get(static.DiDatabase).(database.Database)

		err = db.SetJDoodleKey(ctx.GetEvent().GuildID, jdoodleCreds)
		if err != nil {
			_, err = discordutils.SendEmbedMessageDM(ctx.GetSession(), ctx.GetEvent().User.ID, &discordgo.MessageEmbed{
				Description: "Failed setting up JDoodle Creds.",
			})
			return
		}

		err = db.SetExecEnabled(ctx.GetEvent().GuildID, true)
		if err != nil {
			_, err = discordutils.SendEmbedMessageDM(ctx.GetSession(), ctx.GetEvent().User.ID, &discordgo.MessageEmbed{
				Description: "Failed enabling JDoodle API.",
			})
			return
		}

		_, err = discordutils.SendEmbedMessageDM(ctx.GetSession(), ctx.GetEvent().User.ID, &discordgo.MessageEmbed{
			Description: "Successfully set up JDoodle API.",
		})

		if removeHandler != nil {
			removeHandler()
		}

	})

	return nil
}

func (c *ExecSetup) reset(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)
	err = db.SetJDoodleKey(ctx.GetEvent().GuildID, "")
	if err != nil {
		return err
	}

	err = db.SetExecEnabled(ctx.GetEvent().GuildID, false)
	if err != nil {
		return err
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "API key was deleted from database and system was disabled.",
	}).Send().Error
}

func (c *ExecSetup) toggle(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)
	enabled, err := db.GetExecEnabled(ctx.GetEvent().GuildID)
	if err != nil {
		return err
	}

	err = db.SetExecEnabled(ctx.GetEvent().GuildID, !enabled)
	if err != nil {
		return err
	}

	if !enabled {
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: "Code execution was enabled.",
		}).Send().Error
	} else {
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: "Code execution was disabled.",
		}).Send().Error
	}
}

func (c *ExecSetup) check(ctx ken.SubCommandContext) (err error) {
	db := ctx.Get(static.DiDatabase).(database.Database)
	jdoodlecreds, err := db.GetJDoodleKey(ctx.GetEvent().GuildID)
	if err != nil {
		return err
	}

	enabled, err := db.GetExecEnabled(ctx.GetEvent().GuildID)
	if err != nil {
		return err
	}

	if jdoodlecreds == "" {
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: "No JDoodle API key was set up. Use `/exec setup` to set up the API.",
		}).Send().Error
	}

	if !enabled && jdoodlecreds != "" {
		return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Description: "JDoodle API key was set up, but code execution is disabled. Use `/exec toggle` to enable code execution.",
		}).Send().Error
	}

	return ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "JDoodle API key was set up and code execution is enabled.",
	}).Send().Error
}
