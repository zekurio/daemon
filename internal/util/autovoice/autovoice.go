package autovoice

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/pkg/discordutils"
)

type AVChannel struct {
	GuildID          string
	OwnerID          string
	OriginChannelID  string
	CreatedChannelID string
}

// ActiveChannels is a map of all active autovoice channels, the key is the created channel ID
var ActiveChannels = map[string]AVChannel{}

func Unmarshal(data string) (a AVChannel, err error) {
	rawData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(rawData)
	gobdec := gob.NewDecoder(buffer)

	err = gobdec.Decode(&a)
	if err != nil {
		return
	}

	return
}

func Marshal(a AVChannel) (data string, err error) {
	var buffer bytes.Buffer
	gobenc := gob.NewEncoder(&buffer)

	err = gobenc.Encode(a)
	if err != nil {
		return
	}

	data = base64.StdEncoding.EncodeToString(buffer.Bytes())

	return
}

// Create creates a new autovoice channel and adds it to the guild map
func Create(s *discordgo.Session, gID, oID, cID string) (a AVChannel, err error) {
	var (
		chName string
		pCh    *discordgo.Channel
	)

	pCh, err = discordutils.GetChannel(s, cID)
	if err != nil {
		return
	}

	member, err := discordutils.GetMember(s, gID, oID)
	if err != nil {
		return
	}

	if member.Nick == "" {
		chName = member.User.Username + "'s " + pCh.Name
	} else {
		chName = member.Nick + "'s " + pCh.Name
	}

	createdCh, err := s.GuildChannelCreate(gID, chName, discordgo.ChannelTypeGuildVoice)
	if err != nil {
		return
	}

	_, err = s.ChannelEdit(createdCh.ID, &discordgo.ChannelEdit{
		ParentID: pCh.ParentID,
		Position: pCh.Position + 1,
	})

	a = AVChannel{
		GuildID:          gID,
		OwnerID:          oID,
		OriginChannelID:  cID,
		CreatedChannelID: createdCh.ID,
	}

	ActiveChannels[oID] = a

	if err := s.GuildMemberMove(gID, oID, &createdCh.ID); err != nil {
		return a, err
	}

	return
}

// Get gets an autovoice channel from the map by the owner ID
func Get(oID string) (a AVChannel, ok bool) {
	a, ok = ActiveChannels[oID]
	return
}

// Delete handles the deletion of a autovoice channel, including switching the owner if necessary
func (a *AVChannel) Delete(s *discordgo.Session) (err error) {
	members, err := discordutils.GetVoiceMembers(s, a.GuildID, a.CreatedChannelID)
	if err != nil {
		return
	}

	if len(members) > 0 {
		return a.SwitchOwner(s, members)
	}

	_, err = s.ChannelDelete(a.CreatedChannelID)
	if err != nil {
		return
	}

	delete(ActiveChannels, a.OwnerID)

	return
}

// SwitchOwner switches the owner of a autovoice channel to the next member in the list,
// in case Delete was called while the channel was not empty
func (a *AVChannel) SwitchOwner(s *discordgo.Session, members []*discordgo.Member) (err error) {
	var (
		newOwner = members[0]
		chName   string
	)

	pCh, err := discordutils.GetChannel(s, a.OriginChannelID)
	if err != nil {
		return
	}

	if newOwner.Nick == "" {
		chName = newOwner.User.Username + "'s " + pCh.Name
	} else {
		chName = newOwner.Nick + "'s " + pCh.Name
	}

	_, err = s.ChannelEdit(a.CreatedChannelID, &discordgo.ChannelEdit{
		Name:     chName,
		ParentID: pCh.ParentID,
		Position: pCh.Position + 1,
	})
	if err != nil {
		return
	}

	// delete old owner from map
	delete(ActiveChannels, a.OwnerID)

	// add new owner to map
	a.OwnerID = newOwner.User.ID
	ActiveChannels[a.OwnerID] = *a

	return
}
