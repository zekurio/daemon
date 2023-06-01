package middlewares

import (
	"github.com/zekrotja/ken"
	"time"
)

type CooldownMiddleware struct {
	cooldowns map[string]map[string]int // map[userID]map[commandName]commandInvokeTime
}

func NewCooldownMiddleware() *CooldownMiddleware {
	return &CooldownMiddleware{
		cooldowns: make(map[string]map[string]int),
	}
}

func (m *CooldownMiddleware) Before(ctx *ken.Ctx) (next bool, err error) {
	next = true

	if m.isOnCooldown(ctx) {
		next = false
		err = ctx.RespondError("You are on cooldown.", "")
	}

	return
}

func (m *CooldownMiddleware) isOnCooldown(ctx *ken.Ctx) bool {
	var (
		userID      string
		commandName string
		cmdCooldown int
		now         = time.Now().Second()
	)

	userID = ctx.User().ID
	commandName = ctx.Command.Name()
	if cmd, ok := ctx.Command.(CommandCooldown); ok {
		cmdCooldown = cmd.Cooldown()
	} else {
		return true // no cooldown enabled, so go next
	}

	// check for nil map
	if _, ok := m.cooldowns[userID]; !ok {
		m.cooldowns[userID] = make(map[string]int)
	}

	// check for nil map
	if _, ok := m.cooldowns[userID][commandName]; !ok {
		m.cooldowns[userID][commandName] = now
	}

	// check if cooldown is over
	if now-m.cooldowns[userID][commandName] <= cmdCooldown {
		m.cooldowns[userID][commandName] = now // reset cooldown, might be changed in future
		return false
	}

	return true
}

type CommandCooldown interface {
	// Cooldown returns the cooldown of the command in seconds.
	Cooldown() int
}
