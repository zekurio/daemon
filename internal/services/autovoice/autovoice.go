package autovoice

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/pkg/discordutils"
)

// AutovoiceHandler is the struct that handles the autovoice service
type AutovoiceHandler struct {
	db     database.Database
	s      *discordgo.Session
	guilds map[string]*GuildMap
}

// NewAutovoiceHandler creates a new autovoice handler
func NewAutovoiceHandler(db database.Database, s *discordgo.Session) *AutovoiceHandler {
	return &AutovoiceHandler{
		db:     db,
		s:      s,
		guilds: make(map[string]*GuildMap),
	}
}

type GuildMap map[string]*AVChannel

type AVChannel struct {
	GuildID          string
	OwnerID          string
	OriginChannelID  string
	CreatedChannelID string
}

// Unmarshal decodes a string into a GuildMap, this is used to get the guild map from the database
func Unmarshal(data string) (g GuildMap, err error) {
	rawData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(rawData)
	gobdec := gob.NewDecoder(buffer)

	err = gobdec.Decode(&g)
	if err != nil {
		return
	}

	return
}

// Marshal encodes a GuildMap into a string, this is used to store the guild map in the database
func Marshal(g GuildMap) (data string, err error) {
	var buffer bytes.Buffer
	gobenc := gob.NewEncoder(&buffer)

	err = gobenc.Encode(g)
	if err != nil {
		return
	}

	data = base64.StdEncoding.EncodeToString(buffer.Bytes())

	return
}

// AddGuild adds a guild to the guild map to be used later on
func (h *AutovoiceHandler) AddGuild(guildID string) {
	h.guilds[guildID] = &GuildMap{}
}

// CreateChannel creates a new autovoice channel and adds it to the guild map
func (h *AutovoiceHandler) CreateChannel(guildID, ownerID, parentID string) (a *AVChannel, err error) {
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

	a = &AVChannel{
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
