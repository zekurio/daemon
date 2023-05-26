package autovoice

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/pkg/discordutils"
)

// AutovoiceHandler is the struct that handles the autovoice service
type AutovoiceHandler struct {
	db     database.Database
	s      *discordgo.Session
	guilds map[string]*models.GuildMap
}

var _ AutovoiceProvider = (*AutovoiceHandler)(nil)

// AddGuild adds a guild to the guild map to be used later on
func (h *AutovoiceHandler) AddGuild(guildID string) error {
	h.guilds[guildID] = &models.GuildMap{}

	return nil
}

// CreateChannel creates a new autovoice channel and adds it to the guild map
func (h *AutovoiceHandler) CreateChannel(guildID, ownerID, parentID string) (a *models.AVChannel, err error) {
	var (
		chName string
		pCh    *discordgo.Channel
	)

	pCh, err = h.s.Channel(parentID)
	if err != nil {
		return
	}

	oUser, err := discordutils.GetMember(h.s, guildID, ownerID)
	if err != nil {
		return
	}

	if oUser.Nick != "" {
		chName = oUser.Nick + "'s " + pCh.Name
	} else {
		chName = oUser.User.Username + "'s " + pCh.Name
	}

	ch, err := h.s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:     chName,
		Type:     discordgo.ChannelTypeGuildVoice,
		ParentID: pCh.ParentID,
		Position: pCh.Position + 1,
	})
	if err != nil {
		return
	}

	a = &models.AVChannel{
		GuildID:          guildID,
		OwnerID:          ownerID,
		OriginChannelID:  parentID,
		CreatedChannelID: ch.ID,
	}

	(*h.guilds[guildID])[ch.ID] = a

	if err := h.s.GuildMemberMove(guildID, ownerID, &ch.ID); err != nil {
		return nil, err
	}

	// TODO add to database

	return
}

// DeleteChannel deletes an autovoice channel and removes it from the guild map, it does
// not delete the channel if there are still people in it
func (h *AutovoiceHandler) DeleteChannel(guildID, channelID string) (err error) {
	// check if there are still people in the channel
	members, err := discordutils.GetVoiceMembers(h.s, guildID, channelID)
	if err != nil {
		return
	}

	if len(members) > 0 {
		return h.SwapOwner(members[0].GuildID, members[0].User.ID, channelID)
	}

	_, err = h.s.ChannelDelete(channelID)
	if err != nil {
		return
	}

	delete(*h.guilds[guildID], channelID)

	// TODO delete from database

	return
}

// SwapOwner swaps the owner of an autovoice channel
func (h *AutovoiceHandler) SwapOwner(guildID, newOwner, channelID string) (err error) {
	var (
		chName string
	)

	(*h.guilds[guildID])[channelID].OwnerID = newOwner

	// rename channel
	oUser, err := discordutils.GetMember(h.s, guildID, newOwner)
	if err != nil {
		return
	}

	pCh, err := discordutils.GetChannel(h.s, (*h.guilds[guildID])[channelID].OriginChannelID)
	if err != nil {
		return
	}

	if oUser.Nick != "" {
		chName = oUser.Nick + "'s " + pCh.Name
	} else {
		chName = oUser.User.Username + "'s " + pCh.Name
	}

	_, err = h.s.ChannelEdit(channelID, &discordgo.ChannelEdit{
		Name: chName,
	})
	if err != nil {
		return
	}

	return
}

// GetChannelFromOrigin returns the AVChannel struct from the guild map based on the origin channel ID
func (h *AutovoiceHandler) GetChannelFromOrigin(guildID, originID string) (*models.AVChannel, error) {
	for _, v := range *h.guilds[guildID] {
		if v.OriginChannelID == originID {
			return v, nil
		}
	}

	return nil, errors.New("channel not found")
}

// GetChannelFromOwner returns the AVChannel struct from the guild map based on the owner ID
func (h *AutovoiceHandler) GetChannelFromOwner(guildID, ownerID string) (*models.AVChannel, error) {
	for _, v := range *h.guilds[guildID] {
		if v.OwnerID == ownerID {
			return v, nil
		}
	}

	return nil, errors.New("channel not found")
}
