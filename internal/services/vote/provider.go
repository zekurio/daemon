package vote

import (
	"time"

	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/models"
)

type VotesProvider interface {
	// CreateVote creates a new vote, saves it to the database and
	// returns the vote object.
	CreateVote(ctx ken.SubCommandContext, body, imageURL string, options []string, expire time.Time) (*models.Vote, error)

	// GetVote returns a vote by its id.
	GetVote(voteID string) (*models.Vote, error)
}
