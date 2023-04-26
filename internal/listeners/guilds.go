package listeners

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"

	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
)

type ListenerGuilds struct {
	cfg models.Config
	db  database.Database
}

func NewListenerGuilds(ctn di.Container) *ListenerGuilds {
	return &ListenerGuilds{
		cfg: ctn.Get(static.DiConfig).(models.Config),
		db:  ctn.Get(static.DiDatabase).(database.Database),
	}
}

func (g *ListenerGuilds) Handler(s *discordgo.Session, e *discordgo.GuildCreate) {
	// check if the joinedAt is older than the time
	if e.JoinedAt.Unix() <= time.Now().Unix() {
		return
	}

	limit := g.cfg.Discord.GuildLimit
	if limit == -1 {
		return
	}

	if len(s.State.Guilds) > limit {
		_, err := discordutils.SendMessageDM(s, e.OwnerID,
			fmt.Sprintf("Sorry, the instance owner disallowed me to join more than %d guilds.", limit))
		if err != nil {
			log.With(err).Error("Failed to send message", "GuildID", e.Guild.ID, "UserID", e.OwnerID)
			return
		}
		err = s.GuildLeave(e.Guild.ID)
		if err != nil {
			log.With(err).Error("Failed to leave guild", "GuildID", e.Guild.ID)
			return
		}
	}
}
