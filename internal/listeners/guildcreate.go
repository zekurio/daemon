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
	"github.com/zekurio/daemon/pkg/arrayutils"
	"github.com/zekurio/daemon/pkg/discordutils"
)

type GuildCreate struct {
	cfg models.Config
	db  database.Database
}

func NewGuildCreate(ctn di.Container) *GuildCreate {
	return &GuildCreate{
		cfg: ctn.Get(static.DiConfig).(models.Config),
		db:  ctn.Get(static.DiDatabase).(database.Database),
	}
}

func (g *GuildCreate) GuildLimit(s *discordgo.Session, e *discordgo.GuildCreate) {

	// check if the joinedAt is older than the time
	if e.JoinedAt.Unix() <= time.Now().Unix() {
		return
	}

	log.Debug("Guild limit triggered", "GuildID", e.Guild.ID)

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

		log.Debug("Left guild due to guild limit", "GuildID", e.Guild.ID)

	}

	log.Debug("Finished running guild limit")

}

func (g *GuildCreate) AVCleanup(s *discordgo.Session, e *discordgo.GuildCreate) {

	log.Debug("AutoVoice cleanup triggered", "GuildID", e.Guild.ID)

	// get all the auto voice channels
	createdAutoVoice, err := g.db.GetCreatedAV(e.Guild.ID)
	if err != nil {
		return
	}

	if len(createdAutoVoice) == 0 {
		return
	}

	log.Debug("Found auto voice channels", "GuildID", e.Guild.ID, "Channels", createdAutoVoice)

	// check if the channel still exists
	for _, cID := range createdAutoVoice {
		channel, err := discordutils.GetChannel(s, cID)
		if err != nil {
			// channel doesn't exist, delete it from the array
			createdAutoVoice = arrayutils.RemoveLazy(createdAutoVoice, cID)

			// delete the channel from the database
			if err := g.db.SetAutoRoles(e.Guild.ID, createdAutoVoice); err != nil {
				log.With(err).Error("Failed to delete auto voice channel")
				return
			}
		}

		// check if the channel is empty
		members, err := discordutils.GetVoiceMembers(s, e.Guild.ID, channel.ID)
		if err != nil {
			log.With(err).Error("Failed to get voice members")
			return
		}

		if len(members) == 0 {
			// delete the channel
			if _, err := s.ChannelDelete(cID); err != nil {
				log.With(err).Error("Failed to delete channel")
				return
			}

			// delete the channel from the array
			createdAutoVoice = arrayutils.RemoveLazy(createdAutoVoice, cID)

			// delete the channel from the database
			if err := g.db.SetAutoRoles(e.Guild.ID, createdAutoVoice); err != nil {
				log.With(err).Error("Failed to delete auto voice channel")
				return
			}
		}

	}

	log.Debug("Finished running auto voice cleanup")

}
