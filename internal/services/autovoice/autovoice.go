package autovoice

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/arrayutils"
	"github.com/zekurio/daemon/pkg/discordutils"
)

// AutovoiceHandler is the struct that handles the autovoice service
type AutovoiceHandler struct {
	db       database.Database
	channels map[string]models.AVChannel
}

var _ AutovoiceProvider = (*AutovoiceHandler)(nil)

func InitAutovoice(ctn di.Container) *AutovoiceHandler {
	return &AutovoiceHandler{
		db:       ctn.Get(static.DiDatabase).(database.Database),
		channels: make(map[string]models.AVChannel),
	}
}

func (a *AutovoiceHandler) Deconstruct() error {
	return nil
}

func (a *AutovoiceHandler) Join(s *discordgo.Session, vs *discordgo.VoiceState) (err error) {
	avChannels, err := a.db.GetAutoVoice(vs.GuildID)
	if err != nil || len(avChannels) == 0 {
		return errors.New("no autovoice channels found")
	}

	idString := strings.Join(avChannels, ";")

	if strings.Contains(idString, vs.ChannelID) {
		// create a new channel for the user
		if err = a.createAVChannel(s, vs.GuildID, vs.UserID, vs.ChannelID); err != nil {
			return err
		}
	}

	// check if the user joined a channel that is
	// a created autovoice channel
	if a.isAVChannel(vs.ChannelID) {
		a.appendMember(vs.ChannelID, vs.UserID)
	}

	return nil
}

func (a *AutovoiceHandler) Leave(s *discordgo.Session, vs *discordgo.VoiceState) (err error) {
	if a.isAVChannel(vs.ChannelID) {
		a.removeMember(vs.ChannelID, vs.UserID)

		if a.isOwner(vs.ChannelID, vs.UserID) {
			channel := a.getAVChannel(vs.ChannelID)
			if len(channel.Members) == 0 {
				return a.deleteAVChannel(s, vs.ChannelID)
			} else {
				return a.swapOwner(s, vs.ChannelID, channel.Members[0])
			}
		}
	}

	return nil
}

func (a *AutovoiceHandler) Move(s *discordgo.Session, vsNew, vsOld *discordgo.VoiceState) (err error) {
	// do nothing if the user gets moved to their own channel
	if a.isAVChannel(vsNew.ChannelID) && a.isOwner(vsNew.ChannelID, vsNew.UserID) {
		return nil
	}

	// check if the user moved from an autovoice channel
	// we can use leave here to handle the deletion of the channel
	if err = a.Leave(s, vsOld); err != nil {
		return err
	}

	// now we handle the join part of the move
	if err = a.Join(s, vsNew); err != nil {
		return err
	}

	return err
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

	ch, err := s.GuildChannelCreate(guildID, channelName(ownerMember, pChannel.Name), discordgo.ChannelTypeGuildVoice)
	if err != nil {
		return
	}

	ch, err = s.ChannelEdit(ch.ID, &discordgo.ChannelEdit{
		ParentID: pChannel.ID,
		Position: pChannel.Position + 1,
	})
	if err != nil {
		return
	}

	a.setAVChannel(ch.ID, models.AVChannel{
		GuildID:          guildID,
		OwnerID:          ownerID,
		OriginChannelID:  parentID,
		CreatedChannelID: ch.ID,
		Members:          []string{ownerID},
	})

	err = s.GuildMemberMove(guildID, ownerID, &ch.ID)

	return err
}

func (a *AutovoiceHandler) deleteAVChannel(s *discordgo.Session, channelID string) (err error) {
	channel := a.getAVChannel(channelID)

	_, err = s.ChannelDelete(channel.CreatedChannelID)
	if err != nil {
		return
	}

	delete(a.channels, channelID)

	return
}

// swapOwner swaps the owner of the AVChannel
func (a *AutovoiceHandler) swapOwner(s *discordgo.Session, channelID, newOwnerID string) (err error) {
	channel := a.getAVChannel(channelID)

	// first, we get our new owner member and parent channel
	ownerMember, err := discordutils.GetMember(s, channel.GuildID, newOwnerID)
	if err != nil {
		return
	}

	pChannel, err := discordutils.GetChannel(s, channel.OriginChannelID)
	if err != nil {
		return
	}

	// then we edit the channel
	_, err = s.ChannelEditComplex(channel.CreatedChannelID, &discordgo.ChannelEdit{
		Name:     channelName(ownerMember, pChannel.Name),
		ParentID: channel.OriginChannelID,
	})
	if err != nil {
		return
	}

	// then we set the new owner
	channel.OwnerID = newOwnerID

	// and finally we save the channel
	a.setAVChannel(channelID, *channel)

	return
}

// appendMember appends the given memberID to the AVChannel
// it searches for the AVChannel in the map and appends the memberID
func (a *AutovoiceHandler) appendMember(channelID, memberID string) {
	if channel, ok := a.channels[channelID]; ok {
		channel.Members = arrayutils.Add(channel.Members, memberID, -1)
		a.setAVChannel(channelID, channel)
	}
}

// removeMember removes the given memberID from the AVChannel
// it searches for the AVChannel in the map and removes the memberID
func (a *AutovoiceHandler) removeMember(channelID, memberID string) {
	if channel, ok := a.channels[channelID]; ok {
		channel.Members = arrayutils.RemoveLazy(channel.Members, memberID)
		a.setAVChannel(channelID, channel)
	}
}

// isAVChannel returns true if the given channelID is an autovoice channel
// otherwise it returns false
func (a *AutovoiceHandler) isAVChannel(channelID string) bool {
	return a.getAVChannel(channelID) != nil
}

// getAVChannel returns the AVChannel for the given channelID
func (a *AutovoiceHandler) getAVChannel(channelID string) *models.AVChannel {
	if channel, ok := a.channels[channelID]; ok {
		return &channel
	}

	return &models.AVChannel{}
}

// setAVChannel sets the AVChannel for the given channelID
func (a *AutovoiceHandler) setAVChannel(channelID string, channel models.AVChannel) {
	a.channels[channelID] = channel
}

func (a *AutovoiceHandler) isOwner(channelID, memberID string) bool {
	return a.getAVChannel(channelID).OwnerID == memberID
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
