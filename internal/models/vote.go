package models

import (
	"github.com/bwmarrin/discordgo"
	"time"
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
	StateClosedNC
	StateExpired
)
