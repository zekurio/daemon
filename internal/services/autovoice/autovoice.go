package autovoice

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/pkg/arrayutils"
	"github.com/zekurio/daemon/pkg/discordutils"
)

// AutovoiceHandler is the struct that handles the autovoice service
type AutovoiceHandler struct {
	db     database.Database
	guilds map[string]models.GuildMap
}

var _ AutovoiceProvider = (*AutovoiceHandler)(nil)

func NewAutovoiceHandler(db database.Database) *AutovoiceHandler {
	return &AutovoiceHandler{
		db:     db,
		guilds: make(map[string]models.GuildMap),
	}
}

// SetGuilds overwrites the complete guilds map with a new one
func (h *AutovoiceHandler) SetGuilds(guildMap map[string]models.GuildMap) {
	h.guilds = guildMap
}

// CreateChannel creates a new autovoice channel and adds it to the guild map
func (h *AutovoiceHandler) CreateChannel(s *discordgo.Session, guildID, ownerID, parentID string) (a *models.AVChannel, err error) {
	var (
		chName string
		pCh    *discordgo.Channel
	)

	pCh, err = s.Channel(parentID)
	if err != nil {
		return
	}

	oUser, err := discordutils.GetMember(s, guildID, ownerID)
	if err != nil {
		return
	}

	if oUser.Nick != "" {
		chName = oUser.Nick + "'s " + pCh.Name
	} else {
		chName = oUser.User.Username + "'s " + pCh.Name
	}

	ch, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
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
		Members:          []string{ownerID},
	}

	// Check if the guild map is nil
	if h.guilds[guildID] == nil {
		h.guilds[guildID] = make(models.GuildMap)
	}

	h.guilds[guildID][ch.ID] = a

	if err := s.GuildMemberMove(guildID, ownerID, &ch.ID); err != nil {
		return nil, err
	}

	return a, h.db.AddUpdateGuildMap(guildID, h.guilds[guildID])
}

// DeleteChannel deletes an autovoice channel and removes it from the guild map, it does
// not delete the channel if there are still people in it
func (h *AutovoiceHandler) DeleteChannel(s *discordgo.Session, guildID, channelID string) (err error) {
	if len(h.guilds[guildID][channelID].Members) > 1 {
		newOwner := h.guilds[guildID][channelID].Members[1]

		// swap the owner
		if err = h.SwapOwner(s, guildID, newOwner, channelID); err != nil {
			return err
		}
	}

	_, err = s.ChannelDelete(channelID)
	if err != nil {
		return
	}

	delete(h.guilds[guildID], channelID)

	return h.db.AddUpdateGuildMap(guildID, h.guilds[guildID])
}

// SwapOwner swaps the owner of an autovoice channel
func (h *AutovoiceHandler) SwapOwner(s *discordgo.Session, guildID, newOwner, channelID string) (err error) {
	var (
		chName string
	)

	h.guilds[guildID][channelID].OwnerID = newOwner

	// rename channel
	oUser, err := discordutils.GetMember(s, guildID, newOwner)
	if err != nil {
		return
	}

	pCh, err := discordutils.GetChannel(s, h.guilds[guildID][channelID].OriginChannelID)
	if err != nil {
		return
	}

	if oUser.Nick != "" {
		chName = oUser.Nick + "'s " + pCh.Name
	} else {
		chName = oUser.User.Username + "'s " + pCh.Name
	}

	_, err = s.ChannelEdit(channelID, &discordgo.ChannelEdit{
		Name: chName,
	})
	if err != nil {
		return
	}

	return h.db.AddUpdateGuildMap(guildID, h.guilds[guildID])
}

// Deconstruct deconstructs the autovoice service, saves the guild map to the database
func (h *AutovoiceHandler) Deconstruct() error {
	for k, v := range h.guilds {
		if err := h.db.AddUpdateGuildMap(k, v); err != nil {
			return err
		}
	}

	return nil
}

// GetChannelFromOwner returns the AVChannel struct from the guild map based on the owner ID
func (h *AutovoiceHandler) GetChannelFromOwner(guildID, ownerID string) (*models.AVChannel, error) {
	for _, v := range h.guilds[guildID] {
		if v.OwnerID == ownerID {
			return v, nil
		}
	}

	return nil, errors.New("channel not found")
}

// IsCreatedChannel returns true if the channel ID is a created autovoice channel
func (h *AutovoiceHandler) IsCreatedChannel(guildID, channelID string) bool {
	for _, v := range h.guilds[guildID] {
		if v.CreatedChannelID == channelID {
			return true
		}
	}

	return false
}

// IsOwner returns true if the user ID is the owner of an autovoice channel
func (h *AutovoiceHandler) IsOwner(guildID, userID, channelID string) bool {
	return h.guilds[guildID][channelID].OwnerID == userID
}

// CurrentChannels returns the currently active autovoice channels in a guild
func (h *AutovoiceHandler) CurrentChannels(guildID string) (channels []*models.AVChannel, err error) {
	for _, v := range h.guilds[guildID] {
		channels = append(channels, v)
	}

	return
}

// AddMember adds a member to an autovoice channel
func (h *AutovoiceHandler) AddMember(guildID, userID, channelID string) (err error) {
	h.guilds[guildID][channelID].Members = arrayutils.Add(h.guilds[guildID][channelID].Members, userID, -1)

	return h.db.AddUpdateGuildMap(guildID, h.guilds[guildID])
}

// RemoveMember removes a member from an autovoice channel
func (h *AutovoiceHandler) RemoveMember(guildID, userID, channelID string) (err error) {
	h.guilds[guildID][channelID].Members = arrayutils.RemoveLazy(h.guilds[guildID][channelID].Members, userID)

	return h.db.AddUpdateGuildMap(guildID, h.guilds[guildID])
}
