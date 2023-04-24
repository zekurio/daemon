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

	s.AddHandler(listeners.NewReady().Ready)

	s.AddHandler(listeners.NewGuildCreate(ctn).GuildLimit)

	s.AddHandler(listeners.NewGuildRemove(ctn).FlushGuildData)

	s.AddHandler(listeners.NewGuildMemberAdd(ctn).AutoRole)

	s.AddHandler(listeners.NewVoiceStateUpdate(ctn).AutoVoice)

	return s, nil
}
