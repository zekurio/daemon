package listeners

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/autovoice"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
)

type ListenerAutovoice struct {
	db        database.Database
	avhandler *autovoice.AutovoiceHandler
}

func NewListenerAutovoice(ctn di.Container) *ListenerAutovoice {
	return &ListenerAutovoice{
		db:        ctn.Get(static.DiDatabase).(database.Database),
		avhandler: ctn.Get(static.DiAutovoice).(*autovoice.AutovoiceHandler),
	}
}

func (l *ListenerAutovoice) Handler(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.BeforeUpdate == nil && e.VoiceState != nil {
		err := l.avhandler.Join(s, e)
		if err != nil {
			return
		}
	} else {
		err := l.avhandler.Leave(s, e)
		if err != nil {
			return
		}
	}
}
