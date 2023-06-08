package vote

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
)

type VotesHandler struct {
	db     database.Database
	votes  map[string]models.Vote         // voteID -> vote
	fumsgs map[string]ken.FollowUpMessage // voteID -> followUpMessage
}

var _ VotesProvider = (*VotesHandler)(nil)

func InitVotesHandler(db database.Database) *VotesHandler {
	return &VotesHandler{
		db:     db,
		votes:  make(map[string]models.Vote),
		fumsgs: make(map[string]ken.FollowUpMessage),
	}
}

// Unmarshal decodes a vote from a string
func Unmarshal(data string) (v models.Vote, err error) {
	rawData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(rawData)
	gobdec := gob.NewDecoder(buffer)

	err = gobdec.Decode(&v)
	if err != nil {
		return
	}

	return
}

// Marshal encodes a vote to a string
func Marshal(v models.Vote) (data string, err error) {
	var buffer bytes.Buffer
	gobenc := gob.NewEncoder(&buffer)

	err = gobenc.Encode(v)
	if err != nil {
		return
	}

	data = base64.StdEncoding.EncodeToString(buffer.Bytes())

	return
}

func (v *VotesHandler) CreateVote(ctx ken.SubCommandContext, body string, choices []string, expires string, imageURL string) (*models.Vote, error) {
	//TODO implement me
	panic("implement me")
}

func (v *VotesHandler) GetVote(id string) (*models.Vote, error) {
	//TODO implement me
	panic("implement me")
}

func (v *VotesHandler) GetEmbed(id string) (*discordgo.MessageEmbed, error) {
	//TODO implement me
	panic("implement me")
}

func (v *VotesHandler) AddVote(ctx ken.ComponentContext, id, choice string) error {
	//TODO implement me
	panic("implement me")
}

func (v *VotesHandler) CloseVote(id string, state models.VoteState) error {
	//TODO implement me
	panic("implement me")
}
