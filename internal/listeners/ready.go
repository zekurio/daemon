package listeners

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"

	"github.com/zekurio/daemon/pkg/discordutils"
)

type Ready struct {
}

func NewReady() *Ready {
	return &Ready{}
}

func (r *Ready) Ready(s *discordgo.Session, e *discordgo.Ready) {
	err := s.UpdateListeningStatus("slash commands [WIP]")
	if err != nil {
		return
	}
	log.Info("Signed in!", "Username", fmt.Sprintf("%s#%s", s.State.User.Username, s.State.User.Discriminator), "ID", s.State.User.ID)
	log.Infof("Invite link: %s", discordutils.GetInviteLink(s))

}
