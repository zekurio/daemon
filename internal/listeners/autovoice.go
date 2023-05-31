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

func (l *ListenerAutovoice) Handler(_ *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	vsOld, _ := l.voiceStateCache[e.UserID]
	vsNew := e.VoiceState

	l.voiceStateCache[e.UserID] = vsNew

	ids, err := l.db.GetAutoVoice(e.GuildID)
	if err != nil {
		return
	}
	idString := strings.Join(ids, ";")

	if vsOld == nil || (vsOld != nil && vsOld.ChannelID == "") { // user joined a channel

		// check if user joined a channel that is in the auto voice list
		if strings.Contains(idString, vsNew.ChannelID) {

			// create a new channel
			if _, err := l.autovoiceHandler.CreateChannel(e.GuildID, e.UserID, vsNew.ChannelID); err != nil {
				return
			}

		}

		// check if user joined a created channel to add them to the channel
		if l.autovoiceHandler.IsCreatedChannel(e.GuildID, vsNew.ChannelID) {

			// add user to the channel
			if err := l.autovoiceHandler.AddMember(e.GuildID, e.UserID, vsNew.ChannelID); err != nil {
				return
			}
		}

	} else if (vsOld != nil && vsOld.ChannelID != "") && (vsNew == nil || (vsNew != nil && vsNew.ChannelID == "")) { // user left a channel

		// check if user left a channel that is in the auto voice list
		if strings.Contains(idString, vsOld.ChannelID) {

			// delete the channel
			if err := l.autovoiceHandler.DeleteChannel(e.GuildID, vsOld.ChannelID); err != nil {
				return
			}

			// remove user from the channel
			if err := l.autovoiceHandler.RemoveMember(e.GuildID, e.UserID, vsOld.ChannelID); err != nil {
				return
			}

		}

	} else if (vsOld != nil && vsOld.ChannelID != "") && (vsNew != nil && vsNew.ChannelID != "") { // user switched channels

		// check if user left a channel that is in the auto voice list
		if strings.Contains(idString, vsOld.ChannelID) {

			// delete the channel
			if err := l.autovoiceHandler.DeleteChannel(e.GuildID, vsOld.ChannelID); err != nil {
				return
			}

			// remove user from the channel
			if err := l.autovoiceHandler.RemoveMember(e.GuildID, e.UserID, vsOld.ChannelID); err != nil {
				return
			}

		}

		// check if user joined a channel that is in the auto voice list
		if strings.Contains(idString, vsNew.ChannelID) {

			// create a new channel
			if _, err := l.autovoiceHandler.CreateChannel(e.GuildID, e.UserID, vsNew.ChannelID); err != nil {
				return
			}

		}
	}
}
