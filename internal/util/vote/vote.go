package vote

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
	"github.com/zekurio/daemon/pkg/hashutils"
)

// Vote is a struct for a vote
type Vote struct {
	ID            string
	MsgID         string
	CreatorID     string
	GuildID       string
	ChannelID     string
	Description   string
	ImageURL      string
	Expires       time.Time
	Possibilities []string
	Ticks         map[string]*Tick
}

// Tick is a struct for a tick
type Tick struct {
	UserID string
	Tick   int
}

// VoteState is a type for the state of a vote
type VoteState int

const (
	VoteStateOpen VoteState = iota
	VoteStateClosed
	VoteStateClosedNC
	VoteStateExpired
)

// VotesRunning is a map of all running votes
var VotesRunning = map[string]Vote{}

var VoteEmotes = strings.Fields("\u0031\u20E3 \u0032\u20E3 \u0033\u20E3 \u0034\u20E3 \u0035\u20E3 \u0036\u20E3 \u0037\u20E3 \u0038\u20E3 \u0039\u20E3 \u0030\u20E3")

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

// AsEmbed returns a vode as a discordgo.MessageEmbed
func (v *Vote) AsEmbed(s *discordgo.Session, voteState ...VoteState) (*discordgo.MessageEmbed, error) {
	state := VoteStateOpen
	if len(voteState) > 0 {
		state = voteState[0]
	}

	creator, err := s.User(v.CreatorID)
	if err != nil {
		return nil, err
	}
	title := "Open Vote"
	color := static.ColorDefault

	switch state {
	case VoteStateClosed, VoteStateClosedNC:
		title = "Vote closed"
		color = static.ColorOrange
	case VoteStateExpired:
		title = "Vote expired"
		color = static.ColorViolet
	}

	totalTicks := make(map[int]int)
	for _, t := range v.Ticks {
		if _, ok := totalTicks[t.Tick]; !ok {
			totalTicks[t.Tick] = 1
		} else {
			totalTicks[t.Tick]++
		}
	}

	description := v.Description + "\n\n"
	for i, p := range v.Possibilities {
		description += fmt.Sprintf("%s    %s  -  `%d`\n", VoteEmotes[i], p, totalTicks[i])
	}

	footerText := fmt.Sprintf("ID: %s", v.ID)
	if (v.Expires != time.Time{} && state == VoteStateOpen) {
		footerText = fmt.Sprintf("%s | Expires: %s", footerText, v.Expires.Format("01/02 15:04 MST"))
	}

	emb := &discordgo.MessageEmbed{
		Color:       color,
		Title:       title,
		Description: description,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: creator.AvatarURL("16x16"),
			Name:    creator.Username + "#" + creator.Discriminator,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}

	if len(totalTicks) > 0 && (state == VoteStateClosed || state == VoteStateExpired) {

		values := make([]chart.Value, len(v.Possibilities))

		for i, p := range v.Possibilities {
			values[i] = chart.Value{
				Value: float64(totalTicks[i]),
				Label: p,
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

// AsField returns a vode as a discordgo.MessageEmbedField
func (v *Vote) AsField() *discordgo.MessageEmbedField {
	shortenedDescription := v.Description
	if len(shortenedDescription) > 200 {
		shortenedDescription = shortenedDescription[200:] + "..."
	}

	expiresTxt := "never"
	if (v.Expires != time.Time{}) {
		expiresTxt = v.Expires.Format("01/02 15:04 MST")
	}

	return &discordgo.MessageEmbedField{
		Name: "VID: " + v.ID,
		Value: fmt.Sprintf("**Description:** %s\n**Expires:** %s\n`%d votes`\n[*jump to msg*](%s)",
			shortenedDescription, expiresTxt, len(v.Ticks), discordutils.GetMessageLink(&discordgo.Message{
				ID:        v.MsgID,
				ChannelID: v.ChannelID,
			}, v.GuildID)),
	}
}

// AddReactions adds the vote reactions to the vote message
func (v *Vote) AddReactions(s *discordgo.Session) error {
	for i := 0; i < len(v.Possibilities); i++ {
		err := s.MessageReactionAdd(v.ChannelID, v.MsgID, VoteEmotes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// Tick maps the specificed tick from a user to a vote
func (v *Vote) Tick(s *discordgo.Session, userID string, tick int) (err error) {
	if userID, err = hashutils.HashUserID(userID, []byte(v.ID)); err != nil {
		return
	}

	if t, ok := v.Ticks[userID]; ok {
		t.Tick = tick
	} else {
		v.Ticks[userID] = &Tick{
			UserID: userID,
			Tick:   tick,
		}
	}

	emb, err := v.AsEmbed(s)
	if err != nil {
		return
	}

	_, err = s.ChannelMessageEditEmbed(v.ChannelID, v.MsgID, emb)
	return
}

// SetExpires sets the expiration time of the vote and updates the message
func (v *Vote) SetExpire(s *discordgo.Session, d time.Duration) error {
	v.Expires = time.Now().Add(d)

	emb, err := v.AsEmbed(s)
	if err != nil {
		return err
	}
	_, err = s.ChannelMessageEditEmbed(v.ChannelID, v.MsgID, emb)

	return err
}

// Close closes the vote and removes it from the running votes
func (v *Vote) Close(s *discordgo.Session, voteState VoteState) error {
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
