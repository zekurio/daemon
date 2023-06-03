package listeners

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/autovoice"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
)

type ListenerAutovoice struct {
	db              database.Database
	avhandler       *autovoice.AutovoiceHandler
	voiceStateCache map[string]*discordgo.VoiceState
}

func NewListenerAutovoice(ctn di.Container) *ListenerAutovoice {
	return &ListenerAutovoice{
		db:              ctn.Get(static.DiDatabase).(database.Database),
		avhandler:       ctn.Get(static.DiAutovoice).(*autovoice.AutovoiceHandler),
		voiceStateCache: map[string]*discordgo.VoiceState{},
	}
}

func (l *ListenerAutovoice) Handler(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	vsOld, _ := l.voiceStateCache[e.UserID]
	vsNew := e.VoiceState

	l.voiceStateCache[e.UserID] = vsNew

	if vsOld == nil || (vsOld != nil && vsOld.ChannelID == "") {

		if err := l.avhandler.Join(s, vsNew); err != nil {
			return
		}

	} else if vsOld != nil && vsNew.ChannelID != "" && vsOld.ChannelID != vsNew.ChannelID {

		/* TODO implement move
		if err := l.avhandler.Move(s, vsNew); err != nil {
			return
		}
		*/

	} else if vsOld != nil && vsNew.ChannelID == "" {
		if err := l.avhandler.Leave(s, vsOld); err != nil {
			return
		}
	}
}
