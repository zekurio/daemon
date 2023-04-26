package vote

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/zekrotja/ken"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
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
	ButtonPresses map[string]*ButtonPress
}

// Button is a struct for a vote button
type Button struct {
	Button      *discordgo.Button
	Possibility string
}

// ButtonPress is a struct for a button press
type ButtonPress struct {
	UserID   string
	ButtonID string
}

// Tick is a struct for a tick
type Tick struct {
	UserID string
	Tick   int
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

	// TODO make this use the button presses

	description := v.Description + "\n\n"
	for i, p := range v.Possibilities {
		// enumerate possibilities
		// TODO include press count
		description += fmt.Sprintf("%d. %s\n", i+1, p)
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

	/* TODO make this use the button presses
	if len(totalTicks) > 0 && (state == StateClosed || state == StateExpired) {

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
	*/

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
			shortenedDescription, expiresTxt, len(v.ButtonPresses), discordutils.GetMessageLink(&discordgo.Message{
				ID:        v.MsgID,
				ChannelID: v.ChannelID,
			}, v.GuildID)),
	}
}

// AddButtons adds the buttons to the vote
func (v *Vote) AddButtons(cb *ken.ComponentBuilder) {
	buttons := make([]Button, len(v.Possibilities))
	for i, p := range v.Possibilities {
		customID := fmt.Sprintf("vote_%s_option_%d", v.ID, i)
		button := Button{
			Button: &discordgo.Button{
				Label:    p,
				Style:    discordgo.PrimaryButton,
				CustomID: customID,
			},
			Possibility: p,
		}
		buttons[i] = button
	}

	numRows := len(buttons) / 5
	if len(buttons)%5 > 0 {
		numRows++
	}

	buttonRows := make([][]Button, numRows)
	for i := range buttonRows {
		start := i * 5
		end := start + 5
		if end > len(buttons) {
			end = len(buttons)
		}
		buttonRows[i] = buttons[start:end]
	}

	for _, row := range buttonRows {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			for _, btn := range row {
				b.Add(btn.Button, OnButtonPress(btn.Button.CustomID, v))
			}
		})
	}
}

func OnButtonPress(customID string, v *Vote) func(ctx ken.ComponentContext) bool {
	return func(ctx ken.ComponentContext) bool {
		ctx.SetEphemeral(true)
		err := ctx.Defer()
		if err != nil {
			return false
		}

		userID := ctx.User().ID

		// Check if the user has already voted
		changedVote := false
		for _, buttonPress := range v.ButtonPresses {
			if buttonPress.UserID == userID {
				// Update the user's vote
				buttonPress.ButtonID = customID
				changedVote = true
				break
			}
		}

		// If the user has not voted yet, add their vote
		if !changedVote {
			v.ButtonPresses[customID] = &ButtonPress{
				UserID:   userID,
				ButtonID: customID,
			}
		}

		// Update the embed message with the new vote count
		s := ctx.GetSession()
		emb, err := v.AsEmbed(s)
		if err != nil {
			err = ctx.FollowUpError("Failed to update the vote count.", "").Send().DeleteAfter(10 * time.Second).Error
			return err == nil
		}
		_, err = s.ChannelMessageEditEmbed(v.ChannelID, v.MsgID, emb)
		if err != nil {
			err = ctx.FollowUpError("Failed to update the vote count.", "").Send().DeleteAfter(10 * time.Second).Error
			return err == nil
		}

		return true
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

// Close closes the vote and removes it from the running votes
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
