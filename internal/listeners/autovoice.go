package listeners

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/autovoice"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
	"strings"
)

type ListenerAutovoice struct {
	db               database.Database
	voiceStateCache  map[string]*discordgo.VoiceState
	autovoiceHandler *autovoice.AutovoiceHandler
}

func NewListenerAutovoice(ctn di.Container) *ListenerAutovoice {
	return &ListenerAutovoice{
		db:               ctn.Get(static.DiDatabase).(database.Database),
		voiceStateCache:  map[string]*discordgo.VoiceState{},
		autovoiceHandler: ctn.Get(static.DiAutovoice).(*autovoice.AutovoiceHandler),
	}
}

func (l *ListenerAutovoice) Handler(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	vsOld, _ := l.voiceStateCache[e.UserID]
	vsNew := e.VoiceState

	l.voiceStateCache[e.UserID] = vsNew

	ids, err := l.db.GetAutoVoice(e.GuildID)
	if err != nil {
		return
	}

	idString := strings.Join(ids, ";")

	if vsOld == nil || (vsOld != nil && vsOld.ChannelID == "") { // User joined channel

		if strings.Contains(idString, vsNew.ChannelID) {
			if _, err := l.autovoiceHandler.CreateChannel(s, e.GuildID, e.UserID, vsNew.ChannelID); err != nil {
				return
			}
		}

		if ok := l.autovoiceHandler.IsCreatedChannel(e.GuildID, vsNew.ChannelID); ok {
			if err := l.autovoiceHandler.AddMember(e.GuildID, e.UserID, vsNew.ChannelID); err != nil {
				return
			}
		}

	} else if (vsOld != nil && vsOld.ChannelID != "") && (vsNew == nil && vsNew.ChannelID == "") { // User left channel

		if err := l.autovoiceHandler.RemoveMember(e.GuildID, e.UserID, vsOld.ChannelID); err != nil {
			return
		}

		// delete channel
		if err := l.autovoiceHandler.DeleteChannel(s, e.GuildID, vsOld.ChannelID); err != nil {
			return
		}

	} else if (vsOld != nil && vsOld.ChannelID != "") && (vsNew != nil && vsNew.ChannelID != "") { // User switched channel
		isAVChannel := l.autovoiceHandler.IsCreatedChannel(e.GuildID, vsOld.ChannelID)

		if isAVChannel {
			// DO LITERALLY NOTHING
		} else if strings.Contains(idString, vsNew.ChannelID) { // User switched to a channel that is in the autovoice list
			avChannel, err := l.autovoiceHandler.GetChannelFromOwner(e.GuildID, e.UserID)
			if err != nil || avChannel == nil { // User has no channel
				if _, err := l.autovoiceHandler.CreateChannel(s, e.GuildID, e.UserID, vsNew.ChannelID); err != nil {
					return
				}
			} else { // User has a channel
				// remove member from old channel
				if err := l.autovoiceHandler.RemoveMember(e.GuildID, e.UserID, vsOld.ChannelID); err != nil {
					return
				}

				// delete channel
				if err := l.autovoiceHandler.DeleteChannel(s, e.GuildID, vsOld.ChannelID); err != nil {
					return
				}
			}
		}
	}
}
