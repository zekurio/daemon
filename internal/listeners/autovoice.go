package listeners

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
)

type ListenerAutovoice struct {
	db              database.Database
	voiceStateCache map[string]*discordgo.VoiceState
}

func NewListenerAutovoice(ctn di.Container) *ListenerAutovoice {
	return &ListenerAutovoice{
		db:              ctn.Get(static.DiDatabase).(database.Database),
		voiceStateCache: map[string]*discordgo.VoiceState{},
	}
}

func (l *ListenerAutovoice) Handler(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {

}
