package vote

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/scheduler"
	"github.com/zekurio/daemon/internal/util/static"
)

type VotesHandler struct {
	db     database.Database
	sched  scheduler.Provider
	votes  map[string]models.Vote         // voteID -> vote
	fumsgs map[string]ken.FollowUpMessage // voteID -> followUpMessage
}

var _ VotesProvider = (*VotesHandler)(nil)

func InitVotesHandler(ctn di.Container) *VotesHandler {
	return &VotesHandler{
		db:     ctn.Get(static.DiDatabase).(database.Database),
		sched:  ctn.Get(static.DiScheduler).(scheduler.Provider),
		votes:  make(map[string]models.Vote),
		fumsgs: make(map[string]ken.FollowUpMessage),
	}
}

func (v *VotesHandler) Deconstruct() {
	// TODO implement
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

func (v *VotesHandler) CreateVote(ctx ken.SubCommandContext, body, imageURL string, options []string, expire time.Time) (*models.Vote, error) {
	vote := models.Vote{
		ID:          ctx.GetEvent().ID,
		CreatorID:   ctx.User().ID,
		GuildID:     ctx.GetEvent().GuildID,
		ChannelID:   ctx.GetEvent().ChannelID,
		Description: body,
		Options:     options,
		ImageURL:    imageURL,
		Expires:     expire,
		Buttons:     map[string]models.OptionButton{},
		CurrentVote: map[string]models.CurrentVote{},
	}

	err := v.db.AddUpdateVote(vote)
	if err != nil {
		return nil, err
	}

	v.votes[vote.ID] = vote

	return &vote, nil
}

func (v *VotesHandler) GetVote(voteID string) (*models.Vote, error) {
	if vote, ok := v.votes[voteID]; ok {
		return &vote, nil
	}

	return nil, errors.New("vote not found")
}

func (v *VotesHandler) GetVotes() (map[string]models.Vote, error) {
	return v.votes, nil
}

func (v *VotesHandler) DeleteVote(s *discordgo.Session, voteID string, voteState ...models.VoteState) error {
	vote, err := v.GetVote(voteID)
	if err != nil {
		return err
	}

	currState := models.StateClosed
	if len(voteState) > 0 {
		currState = voteState[0]
	}

	err = vote.Close(s, currState)
	if err != nil {
		return err
	}

	if fumsg, ok := v.fumsgs[voteID]; ok {
		err = fumsg.UnregisterComponentHandlers()
		if err != nil {
			return err
		}
	}

	delete(v.votes, voteID)

	return v.db.DeleteVote(voteID)
}

func (v *VotesHandler) AddFollowUpMessage(voteID string, fumsg ken.FollowUpMessage) {
	v.fumsgs[voteID] = fumsg
}
