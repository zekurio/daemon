package inits

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/listeners"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/util/static"
)

func InitDiscord(ctn di.Container) (*discordgo.Session, error) {
	var err error

	cfg := ctn.Get(static.DiConfig).(models.Config)

	s, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, err
	}

	s.Identify.Intents = discordgo.MakeIntent(static.Intents)

	s.StateEnabled = true
	s.State.TrackChannels = true
	s.State.TrackMembers = true
	s.State.TrackVoice = true

	s.AddHandler(listeners.NewListenerReady(ctn).Handler)

	s.AddHandler(listeners.NewListenerMemberAdd(ctn).Handler)

	s.AddHandler(listeners.NewListenerGuildCreate(ctn).Handler)

	s.AddHandler(listeners.NewListenerAutovoice(ctn).Handler)

	return s, nil
}
