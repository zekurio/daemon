package listeners

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"

	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/database/dberr"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/arrayutils"
	"github.com/zekurio/daemon/pkg/discordutils"
)

type VoiceStateUpdate struct {
	db              database.Database
	avCache         map[string]models.AVChannel
	voiceStateCache map[string]*discordgo.VoiceState
}

func NewVoiceStateUpdate(ctn di.Container) *VoiceStateUpdate {
	return &VoiceStateUpdate{
		db:              ctn.Get(static.DiDatabase).(database.Database),
		avCache:         map[string]models.AVChannel{},
		voiceStateCache: map[string]*discordgo.VoiceState{},
	}
}

func (v *VoiceStateUpdate) AutoVoice(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {

	log.Debug("AutoVoice triggered")

	oldVState := v.voiceStateCache[e.UserID]
	newVState := e.VoiceState

	v.voiceStateCache[e.UserID] = newVState

	ids, err := v.db.GetAutoVoice(e.GuildID)
	if err != nil {
		return
	}
	idString := strings.Join(ids, ";")

	if oldVState == nil || (oldVState != nil && oldVState.ChannelID == "") {

		if !strings.Contains(idString, newVState.ChannelID) {
			return
		}

		if err := v.createAutoVoice(s, e.GuildID, e.UserID, newVState.ChannelID); err != nil {
			return
		}

	} else if oldVState != nil && newVState.ChannelID != "" && oldVState.ChannelID != newVState.ChannelID {

		avChannel, ok := v.avCache[e.UserID]

		if ok && newVState.ChannelID == avChannel.CreatedChannel.ID {

		} else if strings.Contains(idString, newVState.ChannelID) && (!ok || avChannel.CreatedChannel.ID == "") {
			if !ok || avChannel.CreatedChannel.ID == "" {
				if err := v.createAutoVoice(s, e.GuildID, e.UserID, newVState.ChannelID); err != nil {
					return
				}
			} else {
				if err := v.deleteAutoVoice(s, e.UserID); err != nil {
					return
				}
			}
		} else if ok && avChannel.CreatedChannel.ID != "" {
			if err := v.deleteAutoVoice(s, e.UserID); err != nil {
				return
			}
		}

	} else if oldVState != nil && oldVState.ChannelID != "" && newVState.ChannelID == "" {
		if avChannel, ok := v.avCache[e.UserID]; ok && avChannel.CreatedChannel.ID != "" {
			if err := v.deleteAutoVoice(s, e.UserID); err != nil {
				return
			}
		}
		// Add a new else if branch to handle when a user leaves the guild while in a voice channel
	} else if oldVState != nil && oldVState.ChannelID != "" && newVState.GuildID == "" {
		if avChannel, ok := v.avCache[e.UserID]; ok && avChannel.CreatedChannel.ID != "" {
			if err := v.deleteAutoVoice(s, e.UserID); err != nil {
				return
			}
		}
	}
}

func (v *VoiceStateUpdate) createAutoVoice(s *discordgo.Session, guildID, userID, parentChannelID string) error {

	log.Debug("createAutoVoice triggered", "GuildID", guildID, "UserID", userID, "ParentChannelID", parentChannelID)

	var chName string

	pChannel, err := discordutils.GetChannel(s, parentChannelID)
	if err != nil {
		return err
	}

	member, err := discordutils.GetMember(s, guildID, userID)
	if err != nil {
		return err
	}

	if member.Nick == "" {
		chName = member.User.Username + "'s " + pChannel.Name
	} else {
		chName = member.Nick + "'s " + pChannel.Name
	}

	nChannel, err := s.GuildChannelCreate(guildID, chName, discordgo.ChannelTypeGuildVoice)
	if err != nil {
		return err
	}

	nChannel, err = s.ChannelEditComplex(nChannel.ID, &discordgo.ChannelEdit{
		ParentID: pChannel.ParentID,
		Position: pChannel.Position + 1,
	})
	if err != nil {
		return err
	}

	v.avCache[userID] = models.AVChannel{
		OriginChannel:  pChannel,
		CreatedChannel: nChannel,
	}

	if err := s.GuildMemberMove(guildID, userID, &nChannel.ID); err != nil {
		return err
	}

	// add the channel to the database
	createdAutoVoice, err := v.db.GetCreatedAV(guildID)
	if err != nil && err != dberr.ErrNotFound {
		return err
	}

	// dont why, but we have to check if the channel is already in the database
	if arrayutils.Contains(createdAutoVoice, nChannel.ID) {
		return nil
	}

	createdAutoVoice = append(createdAutoVoice, nChannel.ID)
	if err := v.db.SetCreatedAV(guildID, createdAutoVoice); err != nil {
		return err
	}

	return nil

}

// deleteAutoVoice deletes the auto voice channel of a user
func (v *VoiceStateUpdate) deleteAutoVoice(s *discordgo.Session, userID string) error {

	log.Debug("deleteAutoVoice triggered", "UserID", userID)

	channel := v.avCache[userID].CreatedChannel

	// check if the channel still has members
	members, err := discordutils.GetVoiceMembers(s, channel.GuildID, channel.ID)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		_, err := s.ChannelDelete(channel.ID)
		if err != nil {
			return err
		}
		delete(v.avCache, userID)

		// remove the channel from the database
		createdAutoVoice, err := v.db.GetCreatedAV(channel.GuildID)
		if err != nil && err != dberr.ErrNotFound {
			return err
		}

		createdAutoVoice = arrayutils.RemoveLazy(createdAutoVoice, channel.ID)

		if err := v.db.SetCreatedAV(channel.GuildID, createdAutoVoice); err != nil {
			return err
		}

		log.Debug("deleted auto voice channel", "ChannelID", channel.ID, "GuildID", channel.GuildID)

		return nil

	}

	log.Debug("channel still has members", "ChannelID", channel.ID, "GuildID", channel.GuildID, "Members:", members)

	// members are still in the channel, so we just edit the channel again to the origin channel with the new owner name
	newOwner := members[0]

	// Create a new cache entry for the new owner
	newCacheEntry := v.avCache[userID]

	// Assign the new cache entry to the new owner
	v.avCache[newOwner.User.ID] = newCacheEntry

	// Delete the old cache entry
	delete(v.avCache, userID)

	var chName string

	if newOwner.Nick == "" {
		chName = newOwner.User.Username + "'s " + newCacheEntry.OriginChannel.Name
	} else {
		chName = newOwner.Nick + "'s " + newCacheEntry.OriginChannel.Name
	}

	// change the avCache entry to the new owner, aka creating a new one and deleting the old one
	delete(v.avCache, userID)

	_, err = s.ChannelEdit(channel.ID, &discordgo.ChannelEdit{
		Name:     chName,
		ParentID: newCacheEntry.OriginChannel.ParentID,
		Position: newCacheEntry.OriginChannel.Position + 1,
	})

	log.Debug("edited auto voice channel", "ChannelID", channel.ID, "GuildID", channel.GuildID, "NewOwner:", newOwner.User.ID)

	return err

}
