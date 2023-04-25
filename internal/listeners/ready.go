package listeners

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"

	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/scheduler"
	"github.com/zekurio/daemon/internal/util/vote"
	"github.com/zekurio/daemon/pkg/discordutils"
)

type ListenerReady struct {
	db    database.Database
	sched scheduler.Provider
}

func NewReady() *ListenerReady {
	return &ListenerReady{}
}

func (l *ListenerReady) Ready(s *discordgo.Session, e *discordgo.Ready) {
	err := s.UpdateListeningStatus("slash commands [WIP]")
	if err != nil {
		return
	}
	log.Info("Signed in!", "Username", fmt.Sprintf("%s#%s", s.State.User.Username, s.State.User.Discriminator), "ID", s.State.User.ID)
	log.Infof("Invite link: %s", discordutils.GetInviteLink(s))

	votes, err := l.db.GetVotes()
	if err != nil {
	} else {
		vote.VotesRunning = votes
		_, err = l.sched.Schedule("*/10 * * * * *", func() {
			now := time.Now()
			for _, v := range vote.VotesRunning {
				if (v.Expires != time.Time{}) && v.Expires.Before(now) {
					v.Close(s, vote.VoteStateExpired)
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

}
