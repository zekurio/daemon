package vote

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/rs/xid"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/pkg/arrayutils"
	"github.com/zekurio/daemon/pkg/hashutils"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
)

// Vote is a struct for a vote
type Vote struct {
	ID          string
	MsgID       string
	CreatorID   string
	GuildID     string
	ChannelID   string
	Description string
	ImageURL    string
	Expires     time.Time
	Choices     []string
	Buttons     map[string]ChoiceButton
	CurrentVote map[string]CurrentVote
}

// ChoiceButton is a struct for a choice button that
// is used to vote
type ChoiceButton struct {
	Button *discordgo.Button
	Choice string
}

// CurrentVote is a struct for a current user vote
type CurrentVote struct {
	UserID string
	Choice int // the number of the choice in the vote
}

// State VoteState is a type for the state of a vote
type State int

const (
	StateOpen State = iota
	StateClosed
	StateClosedNC
	StateExpired
)

// VotesRunning is a map of all running votes
var VotesRunning = map[string]Vote{}

// Unmarshal decodes a vote from a string
func Unmarshal(data string) (v Vote, err error) {
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
func Marshal(v Vote) (data string, err error) {
	var buffer bytes.Buffer
	gobenc := gob.NewEncoder(&buffer)

	err = gobenc.Encode(v)
	if err != nil {
		return
	}

	data = base64.StdEncoding.EncodeToString(buffer.Bytes())

	return
}

// AsEmbed returns a vote as a discordgo.MessageEmbed
func (v *Vote) AsEmbed(s *discordgo.Session, voteState ...State) (*discordgo.MessageEmbed, error) {
	state := StateOpen
	if len(voteState) > 0 {
		state = voteState[0]
	}

	creator, err := s.User(v.CreatorID)
	if err != nil {
		return nil, err
	}
	title := "Open Vote"
	color := static.ColorDefault
	expires := fmt.Sprintf("Expires <t:%d:R>", v.Expires.Unix())

	if (v.Expires == time.Time{}) {
		expires = "Never expires"
	}

	switch state {
	case StateClosed, StateClosedNC:
		title = "Vote closed"
		color = static.ColorOrange
		expires = "Closed"
	case StateExpired:
		title = "Vote expired"
		color = static.ColorViolet
		expires = fmt.Sprintf("Expired <t:%d:R>", v.Expires.Unix())
	}

	totalVotes := map[int]int{}
	for _, cv := range v.CurrentVote {
		if _, ok := totalVotes[cv.Choice]; !ok {
			totalVotes[cv.Choice] = 1
		} else {
			totalVotes[cv.Choice]++
		}
	}

	description := v.Description + "\n\n"
	for i, p := range v.Choices {
		// enumerate possibilities
		description += fmt.Sprintf("%d. %s - %d\n", i+1, p, totalVotes[i])
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
				Name:   fmt.Sprintf("ID `%s`", v.ID),
				Value:  "",
				Inline: false,
			},
		},
	}

	if len(totalVotes) > 0 && (state == StateClosed || state == StateExpired) {

		values := make([]chart.Value, len(v.Choices))

		for i, p := range v.Choices {
			values[i] = chart.Value{
				Label: p,
				Value: float64(totalVotes[i]),
			}
		}

		pie := chart.PieChart{
			Width:  512,
			Height: 512,
			Values: values,
			Background: chart.Style{
				FillColor: drawing.ColorTransparent,
			},
		}

		imgData := []byte{}
		buff := bytes.NewBuffer(imgData)
		err = pie.Render(chart.PNG, buff)
		if err != nil {
			return nil, err
		}

		_, err := s.ChannelMessageSendComplex(v.ChannelID, &discordgo.MessageSend{
			File: &discordgo.File{
				Name:   fmt.Sprintf("vote_chart_%s.png", v.ID),
				Reader: buff,
			},
			Reference: &discordgo.MessageReference{
				MessageID: v.MsgID,
				ChannelID: v.ChannelID,
				GuildID:   v.GuildID,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if v.ImageURL != "" {
		emb.Image = &discordgo.MessageEmbedImage{
			URL: v.ImageURL,
		}
	}

	return emb, nil
}

// AsField returns a vote as a discordgo.MessageEmbedField
func (v *Vote) AsField() *discordgo.MessageEmbedField {
	shortenedDescription := v.Description
	if len(shortenedDescription) > 200 {
		shortenedDescription = shortenedDescription[200:] + "..."
	}

	expiresTxt := "never"
	if (v.Expires != time.Time{}) {
		expiresTxt = fmt.Sprintf("**Expires <t:%d:R>**", v.Expires.Unix())
	}

	return &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("ID `%s`", v.ID),
		Value: fmt.Sprintf("**Description:** %s\n%s\n`%d votes`\n[*Jump to message*](%s)",
			shortenedDescription, expiresTxt, len(v.CurrentVote), discordutils.GetMessageLink(&discordgo.Message{
				ID:        v.MsgID,
				ChannelID: v.ChannelID,
			}, v.GuildID)),
	}
}

// AddButtons adds the buttons to the vote and returns an array of the choice names
func (v *Vote) AddButtons(cb *ken.ComponentBuilder) ([]string, error) {

	choiceButtons := map[string]*discordgo.Button{}
	for _, c := range v.Choices {
		choiceButtons[c] = &discordgo.Button{
			Label:    c,
			Style:    discordgo.PrimaryButton,
			CustomID: xid.New().String(),
		}
	}

	nCols := len(choiceButtons) / 5
	if len(choiceButtons)%5 != 0 {
		nCols++
	}

	choiceButtonColumns := make([][]ChoiceButton, nCols)
	choiceStrs := make([]string, len(choiceButtons))
	i := 0
	for cStr, cBtn := range choiceButtons {
		choiceButtonColumns[i/5] = append(choiceButtonColumns[i/5], ChoiceButton{
			Button: cBtn,
			Choice: cStr,
		})
		choiceStrs = append(choiceStrs, cStr)
		i++
	}

	for _, cBtns := range choiceButtonColumns {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			for _, cBtn := range cBtns {
				b.Add(cBtn.Button, OnChoiceSelect(cBtn.Choice, v))
			}
			closeBtn := &discordgo.Button{
				Label:    "Close",
				Style:    discordgo.DangerButton,
				CustomID: xid.New().String(),
			}

			closeNCBtn := &discordgo.Button{
				Label:    "Close (no chart)",
				Style:    discordgo.DangerButton,
				CustomID: xid.New().String(),
			}

			b.Add(closeBtn, OnChoiceSelect("close", v))
			b.Add(closeNCBtn, OnChoiceSelect("closenc", v))

		})
	}

	_, err := cb.Build()

	return choiceStrs, err

}

func OnChoiceSelect(choice string, v *Vote) func(ctx ken.ComponentContext) bool {
	return func(ctx ken.ComponentContext) bool {

		if choice == "close" {
			err := v.Close(ctx.GetSession(), StateClosed)
			if err != nil {
				return false
			}
		} else if choice == "closenc" {
			err := v.Close(ctx.GetSession(), StateClosedNC)
			if err != nil {
				return false
			}
		}

		ctx.SetEphemeral(true)
		err := ctx.Defer()
		if err != nil {
			return false
		}

		userID := ctx.User().ID
		if userID, err = hashutils.HashSnowflake(userID, []byte(v.ID)); err != nil {
			return false
		}
		newChoice := choice
		oldChoice := v.Choices[v.CurrentVote[userID].Choice]

		// check if user has already voted
		if _, ok := v.CurrentVote[ctx.User().ID]; ok {
			// check if user is changing their vote
			// or removing their vote
			if newChoice == oldChoice {
				delete(v.CurrentVote, userID)
				err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Your vote for `%s` has been removed", oldChoice),
				}).Send().DeleteAfter(5 * time.Second).Error
			} else {
				// change vote
				v.CurrentVote[userID] = CurrentVote{
					Choice: arrayutils.IndexOf[string](v.Choices, newChoice),
					UserID: userID,
				}
				err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Your vote has been changed from `%s` to `%s`", oldChoice, newChoice),
				}).Send().DeleteAfter(5 * time.Second).Error
			}
		} else {
			// add vote
			v.CurrentVote[userID] = CurrentVote{
				Choice: arrayutils.IndexOf[string](v.Choices, newChoice),
				UserID: userID,
			}
			err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
				Description: fmt.Sprintf("Your vote for `%s` has been added", newChoice),
			}).Send().DeleteAfter(5 * time.Second).Error
		}

		emb, err := v.AsEmbed(ctx.GetSession())
		if err != nil {
			return false
		}

		_, err = ctx.GetSession().ChannelMessageEditEmbed(v.ChannelID, v.MsgID, emb)

		return err == nil

	}

}

// SetExpire sets the expiration time of the vote and updates the message
func (v *Vote) SetExpire(s *discordgo.Session, d time.Duration) error {
	v.Expires = time.Now().Add(d)

	emb, err := v.AsEmbed(s)
	if err != nil {
		return err
	}
	_, err = s.ChannelMessageEditEmbed(v.ChannelID, v.MsgID, emb)

	return err
}

// Close closes the vote, removes the buttons and updates the message
func (v *Vote) Close(s *discordgo.Session, voteState State) error {
	delete(VotesRunning, v.ID)
	emb, err := v.AsEmbed(s, voteState)
	if err != nil {
		return err
	}
	_, err = s.ChannelMessageEditEmbed(v.ChannelID, v.MsgID, emb)
	if err != nil {
		return err
	}
	err = s.MessageReactionsRemoveAll(v.ChannelID, v.MsgID)
	return err
}
