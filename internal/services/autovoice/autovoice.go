package autovoice

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/internal/services/database"
)

// AutovoiceHandler is the struct that handles the autovoice service
type AutovoiceHandler struct {
	db     *database.Database
	s      *discordgo.Session
	guilds map[string]*GuildMap
}

type GuildMap map[string]*AVChannel

type AVChannel struct {
	GuildID          string
	OwnerID          string
	OriginChannelID  string
	CreatedChannelID string
}

// Unmarshal decodes a string into a GuildMap
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

// Marshal encodes a GuildMap into a string
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

func (h *AutovoiceHandler) AddGuild(gID string) {

}
