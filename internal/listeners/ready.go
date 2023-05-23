package listeners

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"

	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/scheduler"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/internal/util/vote"
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

	_, err = l.sched.Schedule("*/30 * * * * *", func() {
		votes, err := l.db.GetVotes()
		if err != nil {
			log.Error("Failed getting votes from database: %s", err.Error())
			return
		}
		vote.VotesRunning = votes
		now := time.Now()
		for _, v := range vote.VotesRunning {
			if (v.Expires != time.Time{}) && v.Expires.Before(now) {
				err := v.Close(s, vote.StateExpired)
				if err != nil {
					log.Error("Failed closing vote: %s", err.Error())
				}
				if err = l.db.DeleteVote(v.ID); err != nil {
					log.Error("Failed deleting vote from database: %s", err.Error())
				}
			}
		}
	})
	if err != nil {
		log.Error("Failed scheduling vote cleanup: %s", err.Error())
	}
}
