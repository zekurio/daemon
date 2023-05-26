package inits

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/autovoice"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
)

func InitAutvoice(ctn di.Container) *autovoice.AutovoiceHandler {

	_ = ctn.Get(static.DiDatabase).(database.Database)
	_ = ctn.Get(static.DiDiscord).(*discordgo.Session)

	// TODO populate guilds map with data from database

	return nil
}
