package vote

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/models"
)

type VotesProvider interface {
	// CreateVote creates a new vote, saves it to the database and
	// returns the vote object.
	CreateVote(ctx ken.SubCommandContext, body, imageURL string, choices []string, expire time.Time) (*models.Vote, error)

	// GetVote returns a vote by its id.
	GetVote(id string) (*models.Vote, error)

	// GetEmbed returns the embed of a vote.
	GetEmbed(s *discordgo.Session, id string, state ...models.VoteState) (*discordgo.MessageEmbed, error)

	// AddVote adds a vote to a given vote.
	AddVote(ctx ken.ComponentContext, id, choice string) error

	// CloseVote closes a vote.
	CloseVote(id string, state models.VoteState) error
}
