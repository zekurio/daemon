// TODO rewrite this piece of crap to use the new auto voice system
package listeners

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"

	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/autovoice"
	"github.com/zekurio/daemon/internal/util/static"
)

type VoiceStateUpdate struct {
	db              database.Database
	voiceStateCache map[string]*discordgo.VoiceState
}

func NewVoiceStateUpdate(ctn di.Container) *VoiceStateUpdate {
	return &VoiceStateUpdate{
		db:              ctn.Get(static.DiDatabase).(database.Database),
		voiceStateCache: map[string]*discordgo.VoiceState{},
	}
}

func (v *VoiceStateUpdate) AutoVoice(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {

	oldVState := v.voiceStateCache[e.UserID]
	newVState := e.VoiceState

	v.voiceStateCache[e.UserID] = newVState

	ids, err := v.db.GetAutoVoice(e.GuildID)
	if err != nil {
		return
	}
	idString := strings.Join(ids, ";")

	if oldVState == nil || (oldVState != nil && oldVState.ChannelID == "") {

		var av autovoice.AVChannel

		if !strings.Contains(idString, newVState.ChannelID) {
			return
		}

		if av, err = autovoice.Create(s, e.GuildID, e.UserID, newVState.ChannelID); err != nil {
			return
		}

		if err = v.db.AddUpdateAVChannel(av); err != nil {
			return
		}

	} else if oldVState != nil && newVState.ChannelID != "" && oldVState.ChannelID != newVState.ChannelID {

		avChannel, ok := autovoice.Get(e.UserID)

		if ok && newVState.ChannelID == avChannel.CreatedChannelID {

		} else if strings.Contains(idString, newVState.ChannelID) && (!ok || avChannel.CreatedChannelID == "") {
			if !ok || avChannel.CreatedChannelID == "" {
				if _, err := autovoice.Create(s, e.GuildID, e.UserID, newVState.ChannelID); err != nil {
					return
				}
			} else {
				if err := avChannel.Delete(s); err != nil {
					return
				}
			}
		} else if ok && avChannel.CreatedChannelID != "" {
			if err := avChannel.Delete(s); err != nil {
				return
			}
		}

	} else if oldVState != nil && oldVState.ChannelID != "" && newVState.ChannelID == "" {
		if avChannel, ok := autovoice.Get(e.UserID); ok && avChannel.CreatedChannelID != "" {
			if err := avChannel.Delete(s); err != nil {
				return
			}
		}
		// Add a new else if branch to handle when a user leaves the guild while in a voice channel
	} else if oldVState != nil && oldVState.ChannelID != "" && newVState.GuildID == "" {
		if avChannel, ok := autovoice.Get(e.UserID); ok && avChannel.CreatedChannelID != "" {
			if err := avChannel.Delete(s); err != nil {
				return
			}
		}
	}
}
