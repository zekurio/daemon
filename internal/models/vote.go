package models

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/xid"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/arrayutils"
	"github.com/zekurio/daemon/pkg/discordutils"
	"github.com/zekurio/daemon/pkg/hashutils"
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

// VoteState is a type for the state of a vote
type VoteState int

const (
	StateOpen VoteState = iota
	StateClosed
	StateExpired
)

func (v *Vote) AsEmbed(s *discordgo.Session, state ...VoteState) (*discordgo.MessageEmbed, error) {
	currState := StateOpen
	if len(state) > 0 {
		currState = state[0]
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

	switch currState {
	case StateClosed:
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
				Name:   fmt.Sprintf("ID `%s`", v.ID),
				Value:  "",
				Inline: false,
			},
		},
	}

	if v.ImageURL != "" {
		emb.Image = &discordgo.MessageEmbedImage{
			URL: v.ImageURL,
		}
	}

	return emb, nil
}

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
				b.Add(cBtn.Button, AddVote(cBtn.Choice, v))
			}
		})
	}

	_, err := cb.Build()

	return choiceStrs, err

}

func AddVote(choice string, vote *Vote) func(ctx ken.ComponentContext) bool {
	return func(ctx ken.ComponentContext) bool {
		ctx.SetEphemeral(true)
		err := ctx.Defer()
		if err != nil {
			return false
		}

		userID := ctx.User().ID
		if userID, err = hashutils.HashSnowflake(userID, []byte(vote.ID)); err != nil {
			return false
		}
		newChoice := choice
		oldChoice := vote.Choices[vote.CurrentVote[userID].Choice]

		// check if user has already voted
		if _, ok := vote.CurrentVote[ctx.User().ID]; ok {
			// check if user is changing their vote
			// or removing their vote
			if newChoice == oldChoice {
				delete(vote.CurrentVote, userID)
				err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Your vote for `%s` has been removed", oldChoice),
				}).Send().DeleteAfter(5 * time.Second).Error
			} else {
				// change vote
				vote.CurrentVote[userID] = CurrentVote{
					Choice: arrayutils.IndexOf(vote.Choices, newChoice),
					UserID: userID,
				}
				err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Your vote has been changed from `%s` to `%s`", oldChoice, newChoice),
				}).Send().DeleteAfter(5 * time.Second).Error
			}
		} else {
			// add vote
			vote.CurrentVote[userID] = CurrentVote{
				Choice: arrayutils.IndexOf(vote.Choices, newChoice),
				UserID: userID,
			}
			err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
				Description: fmt.Sprintf("Your vote for `%s` has been added", newChoice),
			}).Send().DeleteAfter(5 * time.Second).Error
		}

		emb, err := vote.AsEmbed(ctx.GetSession())
		if err != nil {
			return false
		}

		_, err = ctx.GetSession().ChannelMessageEditEmbed(vote.ChannelID, vote.MsgID, emb)

		return err == nil

	}
}
