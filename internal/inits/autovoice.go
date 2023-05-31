package inits

import (
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/autovoice"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
)

func InitAutovoice(ctn di.Container) *autovoice.AutovoiceHandler {

	db := ctn.Get(static.DiDatabase).(database.Database)
	s := ctn.Get(static.DiDiscord).(*discordgo.Session)

	handler := autovoice.NewAutvoiceHandler(db, s)

	// populate guilds
	guilds, err := db.GetAllGuildMaps()
	if err != nil {
		log.Error("Failed to get guilds from database", err)
	}

	handler.SetGuilds(guilds)

	return handler
}
