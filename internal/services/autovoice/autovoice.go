package autovoice

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/pkg/arrayutils"
	"github.com/zekurio/daemon/pkg/discordutils"
	"strings"
)

// AutovoiceHandler is the struct that handles the autovoice service
type AutovoiceHandler struct {
	db       database.Database
	channels map[string]models.AVChannel // userID -> AVChannel
}

var _ AutovoiceProvider = (*AutovoiceHandler)(nil)

func NewAutovoiceHandler(db database.Database) *AutovoiceHandler {
	return &AutovoiceHandler{
		db:       db,
		channels: make(map[string]models.AVChannel),
	}
}

func (a *AutovoiceHandler) Join(s *discordgo.Session, e *discordgo.VoiceStateUpdate) (err error) {

	// we established that the user joined a voice channel
	// until here

	// now we check if the user joined a channel that is
	// in the autovoice list
	avChannels, err := a.db.GetAutoVoice(e.GuildID)
	if err != nil || len(avChannels) == 0 {
		return errors.New("no autovoice channels found")
	}

	idString := strings.Join(avChannels, ";")

	if strings.Contains(e.ChannelID, idString) {
		// create a new channel for the user
		err = a.createAVChannel(s, e.GuildID, e.UserID, e.ChannelID)
		if err != nil {
			return err
		}
	}

	// check if the user joined a channel that is
	// a created autovoice channel
	if a.isAVChannel(e.ChannelID) {
		a.appendMember(e.ChannelID, e.UserID)
	}

	return nil
}

func (a *AutovoiceHandler) Leave(s *discordgo.Session, e *discordgo.VoiceStateUpdate) (err error) {
	//TODO implement me
	panic("implement me")
}

func (a *AutovoiceHandler) Move(s *discordgo.Session, e *discordgo.VoiceStateUpdate) (err error) {
	//TODO implement me
	panic("implement me")
}

// HELPERS

func (a *AutovoiceHandler) createAVChannel(s *discordgo.Session, guildID, ownerID, parentID string) (err error) {
	ownerMember, err := discordutils.GetMember(s, guildID, ownerID)
	if err != nil {
		return
	}
	pChannel, err := discordutils.GetChannel(s, parentID)
	if err != nil {
		return
	}

	ch, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:     channelName(ownerMember, pChannel.Name),
		Type:     discordgo.ChannelTypeGuildVoice,
		ParentID: parentID,
		Position: pChannel.Position + 1,
	})
	if err != nil {
		return
	}

	a.setAVChannel(ownerID, models.AVChannel{
		GuildID:          guildID,
		OwnerID:          ownerID,
		OriginChannelID:  parentID,
		CreatedChannelID: ch.ID,
		Members:          []string{ownerID},
	})

	err = s.GuildMemberMove(guildID, ownerID, &ch.ID)

	return err
}

func (a *AutovoiceHandler) deleteAVChannel(s *discordgo.Session, guildID, ownerID string) (err error) {
	channel := a.getAVChannel(ownerID)

	_, err = s.ChannelDelete(channel.CreatedChannelID)
	if err != nil {
		return
	}

	delete(a.channels, ownerID)

	return
}

// appendMember appends the given memberID to the AVChannel
// it searches for the AVChannel in the map and appends the memberID
func (a *AutovoiceHandler) appendMember(channelID, memberID string) {
	for _, channel := range a.channels {
		if channel.CreatedChannelID == channelID {
			channel.Members = append(channel.Members, memberID)
		}
	}
}

// removeMember removes the given memberID from the AVChannel
// it searches for the AVChannel in the map and removes the memberID
func (a *AutovoiceHandler) removeMember(channelID, memberID string) {
	for _, channel := range a.channels {
		if channel.CreatedChannelID == channelID {
			for _, member := range channel.Members {
				if member == memberID {
					channel.Members = arrayutils.RemoveLazy(channel.Members, memberID)
				}
			}
		}
	}
}

// isAVChannel returns true if the given channelID is an autovoice channel
// otherwise it returns false
func (a *AutovoiceHandler) isAVChannel(channelID string) bool {
	for _, channel := range a.channels {
		if channel.CreatedChannelID == channelID {
			return true
		}
	}

	return false
}

// getAVChannel returns the AVChannel for the given userID
// if it exists, otherwise it returns an empty AVChannel
func (a *AutovoiceHandler) getAVChannel(userID string) *models.AVChannel {
	if channel, ok := a.channels[userID]; ok {
		return &channel
	}

	return &models.AVChannel{}
}

// setAVChannel sets the AVChannel for the given userID
func (a *AutovoiceHandler) setAVChannel(userID string, channel models.AVChannel) {
	a.channels[userID] = channel
}

// channelName returns the name of the channel that should be created
// for the given user
func channelName(member *discordgo.Member, pChannelName string) string {
	if member.Nick != "" {
		return member.Nick + "'s " + pChannelName
	} else {
		return member.User.Username + "'s " + pChannelName
	}
}
