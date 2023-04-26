package inits

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"
	"github.com/zekrotja/ken"

	"github.com/zekurio/daemon/internal/middlewares"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/slashcommands"
	"github.com/zekurio/daemon/internal/usercommands"
	"github.com/zekurio/daemon/internal/util/static"
)

func InitKen(ctn di.Container) (*ken.Ken, error) {

	s := ctn.Get(static.DiDiscord).(*discordgo.Session)
	p := ctn.Get(static.DiPermissions).(*permissions.Permissions)

	k, err := ken.New(s, ken.Options{
		EmbedColors: ken.EmbedColors{
			Default: static.ColorDefault,
			Error:   static.ColorRed,
		},
		DependencyProvider: ctn,
		OnSystemError:      systemErrorHandler,
		OnCommandError:     commandErrorHandler,
	})

	if err != nil {
		return nil, err
	}

	// register slashcommands
	err = k.RegisterCommands(
		// slashcommands
		new(slashcommands.Profile),
		new(slashcommands.Autorole),
		new(slashcommands.Autovoice),
		new(slashcommands.Guild),
		new(slashcommands.Perms),
		new(slashcommands.Vote),

		// usercommands
		new(usercommands.About),
	)
	if err != nil {
		return nil, err
	}

	err = k.RegisterMiddlewares(
		p,
		middlewares.NewDisableCommandsMiddleware(ctn),
	)

	return k, err

}

func systemErrorHandler(context string, err error, args ...interface{}) {
	log.Error("Ken system error")
}

func commandErrorHandler(err error, ctx *ken.Ctx) {
	ctx.Defer()

	if err == ken.ErrNotDMCapable {
		ctx.FollowUpError("This command can not be used in DMs.", "").Send()
		return
	}

	ctx.FollowUpError(
		fmt.Sprintf("The command execution failed unexpectedly:\n```\n%s\n```", err.Error()),
		"Command execution failed").Send()
}
