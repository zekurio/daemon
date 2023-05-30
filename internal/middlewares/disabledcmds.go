package middlewares

import (
	"github.com/sarulabs/di/v2"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/arrayutils"
)

type DisableCommandsMiddleware struct {
	cfg models.Config
}

var (
	_ ken.MiddlewareBefore = (*DisableCommandsMiddleware)(nil)
)

func NewDisableCommandsMiddleware(ctn di.Container) *DisableCommandsMiddleware {
	return &DisableCommandsMiddleware{
		cfg: ctn.Get(static.DiConfig).(models.Config),
	}
}

func (m *DisableCommandsMiddleware) Before(ctx *ken.Ctx) (next bool, err error) {
	next = true

	if m.isDisabled(ctx.Command.Name()) {
		next = false
		err = ctx.RespondError("This command is disabled by config.", "")
	}

	return
}

func (m *DisableCommandsMiddleware) isDisabled(invoke string) bool {
	disabledCmds := m.cfg.Discord.DisabledCommands
	return len(disabledCmds) != 0 && arrayutils.Contains(disabledCmds, invoke)
}
