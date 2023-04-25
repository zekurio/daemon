package listeners

import (
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"

	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
)

type GuildRemove struct {
	db database.Database
}

func NewGuildRemove(ctn di.Container) *GuildRemove {
	return &GuildRemove{
		db: ctn.Get(static.DiDatabase).(database.Database),
	}
}

func (g *GuildRemove) FlushGuildData(s *discordgo.Session, e *discordgo.GuildDelete) {
	err := g.db.FlushGuildData(e.ID)
	if err != nil {
		log.With(err).Error("Failed to flush guild data")
		return
	}
}
