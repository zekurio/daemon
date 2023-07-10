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
	Options     []string
	Buttons     map[string]OptionButton
	CurrentVote map[string]CurrentVote
}

// OptionButton is a struct for a option button that
// is used to vote
type OptionButton struct {
	Button *discordgo.Button
	Option string
}

// CurrentVote is a struct for a current user vote
type CurrentVote struct {
	UserID string
	Option int // the number of the option in the vote
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
		if _, ok := totalVotes[cv.Option]; !ok {
			totalVotes[cv.Option] = 1
		} else {
			totalVotes[cv.Option]++
		}
	}

	description := v.Description + "\n\n"
	for i, p := range v.Options {
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

	optionButtons := map[string]*discordgo.Button{}
	for _, c := range v.Options {
		optionButtons[c] = &discordgo.Button{
			Label:    c,
			Style:    discordgo.PrimaryButton,
			CustomID: xid.New().String(),
		}
	}

	nCols := len(optionButtons) / 5
	if len(optionButtons)%5 != 0 {
		nCols++
	}

	optionButtonColumns := make([][]OptionButton, nCols)
	optionStrs := make([]string, len(optionButtons))
	i := 0
	for cStr, cBtn := range optionButtons {
		optionButtonColumns[i/5] = append(optionButtonColumns[i/5], OptionButton{
			Button: cBtn,
			Option: cStr,
		})
		optionStrs = append(optionStrs, cStr)
		i++
	}

	for _, cBtns := range optionButtonColumns {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			for _, cBtn := range cBtns {
				b.Add(cBtn.Button, AddVote(cBtn.Option, v))
			}
		})
	}

	_, err := cb.Build()

	return optionStrs, err

}

// Close closes a vote by editing the message
func (v *Vote) Close(s *discordgo.Session, voteState ...VoteState) error {
	currState := StateClosed
	if len(voteState) > 0 {
		currState = voteState[0]
	}

	emb, err := v.AsEmbed(s, currState)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Components: []discordgo.MessageComponent{},
		Embed:      emb,
		ID:         v.MsgID,
		Channel:    v.ChannelID,
	})

	return err
}

func (v *Vote) SetExpire(s *discordgo.Session, d time.Duration) error {
	v.Expires = time.Now().Add(d)

	emb, err := v.AsEmbed(s)
	if err != nil {
		return err
	}
	_, err = s.ChannelMessageEditEmbed(v.ChannelID, v.MsgID, emb)

	return err
}

func AddVote(option string, vote *Vote) func(ctx ken.ComponentContext) bool {
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
		newOption := option
		oldOption := vote.Options[vote.CurrentVote[userID].Option]

		// check if user has already voted
		if _, ok := vote.CurrentVote[ctx.User().ID]; ok {
			// check if user is changing their vote
			// or removing their vote
			if newOption == oldOption {
				delete(vote.CurrentVote, userID)
				err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Your vote for `%s` has been removed", oldOption),
				}).Send().DeleteAfter(5 * time.Second).Error
			} else {
				// change vote
				vote.CurrentVote[userID] = CurrentVote{
					Option: arrayutils.IndexOf(vote.Options, newOption),
					UserID: userID,
				}
				err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Your vote has been changed from `%s` to `%s`", oldOption, newOption),
				}).Send().DeleteAfter(5 * time.Second).Error
			}
		} else {
			// add vote
			vote.CurrentVote[userID] = CurrentVote{
				Option: arrayutils.IndexOf(vote.Options, newOption),
				UserID: userID,
			}
			err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
				Description: fmt.Sprintf("Your vote for `%s` has been added", newOption),
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
