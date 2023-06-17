package vote

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
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

func (v *VotesHandler) CreateVote(ctx ken.SubCommandContext, body, imageURL string, choices []string, expire time.Time) (*models.Vote, error) {

	ivote := models.Vote{
		ID:          ctx.GetEvent().ID,
		CreatorID:   ctx.User().ID,
		GuildID:     ctx.GetEvent().GuildID,
		ChannelID:   ctx.GetEvent().ChannelID,
		Description: body,
		Choices:     choices,
		ImageURL:    imageURL,
		Expires:     expire,
		Buttons:     map[string]models.ChoiceButton{},
		CurrentVote: map[string]models.CurrentVote{},
	}

	err := v.db.AddUpdateVote(ivote)
	if err != nil {
		return nil, err
	}

	v.votes[ivote.ID] = ivote

	return &ivote, nil
}

func (v *VotesHandler) GetVote(id string) (*models.Vote, error) {
	if vote, ok := v.votes[id]; ok {
		return &vote, nil
	}

	return nil, errors.New("vote not found")
}

func (v *VotesHandler) GetEmbed(s *discordgo.Session, id string, state ...models.VoteState) (*discordgo.MessageEmbed, error) {
	vote := v.votes[id]

	currState := models.StateOpen
	if len(state) > 0 {
		currState = state[0]
	}

	creator, err := s.User(vote.CreatorID)
	if err != nil {
		return nil, err
	}
	title := "Open Vote"
	color := static.ColorDefault
	expires := fmt.Sprintf("Expires <t:%d:R>", vote.Expires.Unix())

	if (vote.Expires == time.Time{}) {
		expires = "Never expires"
	}

	switch currState {
	case models.StateClosed, models.StateClosedNC:
		title = "Vote closed"
		color = static.ColorOrange
		expires = "Closed"
	case models.StateExpired:
		title = "Vote expired"
		color = static.ColorViolet
		expires = fmt.Sprintf("Expired <t:%d:R>", vote.Expires.Unix())
	}

	totalVotes := map[int]int{}
	for _, cv := range vote.CurrentVote {
		if _, ok := totalVotes[cv.Choice]; !ok {
			totalVotes[cv.Choice] = 1
		} else {
			totalVotes[cv.Choice]++
		}
	}

	description := vote.Description + "\n\n"
	for i, p := range vote.Choices {
		description += fmt.Sprintf("**%d. %s** - `%d`\n", i+1, p, totalVotes[i])
	}

	emb := &discordgo.MessageEmbed{
		Color:       color,
		Title:       title,
		Description: description,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: creator.AvatarURL("16x16"),
			Name:    creator.Username + "#" + creator.Discriminator,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   expires,
				Value:  "",
				Inline: false,
			},
			{
				Name:   fmt.Sprintf("ID `%s`", vote.ID),
				Value:  "",
				Inline: false,
			},
		},
	}

	if vote.ImageURL != "" {
		emb.Image = &discordgo.MessageEmbedImage{
			URL: vote.ImageURL,
		}
	}

	return emb, nil

}

func (v *VotesHandler) AddVote(ctx ken.ComponentContext, id, choice string) error {
	//TODO implement me
	panic("implement me")
}

func (v *VotesHandler) CloseVote(id string, state models.VoteState) error {
	//TODO implement me
	panic("implement me")
}
