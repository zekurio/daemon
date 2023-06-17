package listeners

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"

	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/scheduler"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
)

type ListenerReady struct {
	db    database.Database
	sched scheduler.Provider
}

func NewListenerReady(ctn di.Container) *ListenerReady {
	return &ListenerReady{
		db:    ctn.Get(static.DiDatabase).(database.Database),
		sched: ctn.Get(static.DiScheduler).(scheduler.Provider),
	}
}

func (l *ListenerReady) Handler(s *discordgo.Session, e *discordgo.Ready) {
	err := s.UpdateListeningStatus("slash commands [WIP]")
	if err != nil {
		return
	}
	log.Info("Signed in!", "Username", fmt.Sprintf("%s#%s", e.User.Username, e.User.Discriminator), "ID", e.User.ID)
	log.Infof("Invite link: %s", discordutils.GetInviteLink(s))

	l.sched.Start()
}
