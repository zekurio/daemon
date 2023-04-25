package listeners

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/autovoice"
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

	oldVState := l.voiceStateCache[e.UserID]
	newVState := e.VoiceState

	l.voiceStateCache[e.UserID] = newVState

	ids, err := l.db.GetAutoVoice(e.GuildID)
	if err != nil {
		return
	}
	idString := strings.Join(ids, ";")

	if oldVState == nil || (oldVState != nil && oldVState.ChannelID == "") {

		if !strings.Contains(idString, newVState.ChannelID) {
			return
		}

		av, err := autovoice.Create(s, e.GuildID, e.UserID, newVState.ChannelID)
		if err != nil {
			return
		}

		if err = l.db.AddUpdateAVChannel(av); err != nil {
			return
		}

	} else if oldVState != nil && newVState.ChannelID != "" && oldVState.ChannelID != newVState.ChannelID {

		avChannel, ok := autovoice.Get(e.UserID)

		if ok && newVState.ChannelID == avChannel.CreatedChannelID {

		} else if strings.Contains(idString, newVState.ChannelID) && (!ok || avChannel.CreatedChannelID == "") {
			if !ok || avChannel.CreatedChannelID == "" {
				av, err := autovoice.Create(s, e.GuildID, e.UserID, newVState.ChannelID)
				if err != nil {
					return
				}

				if err = l.db.AddUpdateAVChannel(av); err != nil {
					return
				}
			} else {
				if err := avChannel.Delete(s); err != nil {
					return
				}
			}
		} else if ok && avChannel.CreatedChannelID != "" {
			if avChannel, ok := autovoice.Get(e.UserID); ok && avChannel.CreatedChannelID != "" {
				err := avChannel.Delete(s)
				if err != nil {
					return
				}

				if err = l.db.DeleteAVChannel(avChannel.CreatedChannelID); err != nil {
					return
				}
			}
		}
	} else if oldVState != nil && oldVState.ChannelID != "" && newVState.ChannelID == "" {
		if avChannel, ok := autovoice.Get(e.UserID); ok && avChannel.CreatedChannelID != "" {
			err := avChannel.Delete(s)
			if err != nil {
				return
			}

			if err = l.db.DeleteAVChannel(avChannel.CreatedChannelID); err != nil {
				return
			}
		}
	} else if oldVState != nil && oldVState.ChannelID != "" && newVState.GuildID == "" {
		if avChannel, ok := autovoice.Get(e.UserID); ok && avChannel.CreatedChannelID != "" {
			err := avChannel.Delete(s)
			if err != nil {
				return
			}

			if err = l.db.DeleteAVChannel(avChannel.CreatedChannelID); err != nil {
				return
			}
		}
	}
}
