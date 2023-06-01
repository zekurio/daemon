package middlewares

import (
	"github.com/zekrotja/ken"
)

type CooldownMiddleware struct {
	cooldowns map[string]map[string]int // map[userID]map[commandName]cooldown
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
	userID := ctx.User().ID
	commandName := ctx.Command.Name()

	if _, ok := m.cooldowns[userID]; !ok {
		m.cooldowns[userID] = make(map[string]int)
	}

	if _, ok := m.cooldowns[userID][commandName]; !ok {
		m.cooldowns[userID][commandName] = 0
	}

	if m.cooldowns[userID][commandName] > 0 {
		return true
	}

	m.cooldowns[userID][commandName] = ctx.Command.(CommandCooldown).Cooldown() // TODO: Add cooldown to command struct
	return false
}

type CommandCooldown interface {
	// Cooldown returns the cooldown of the command in seconds.
	Cooldown() int
}
