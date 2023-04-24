package database

import (
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/util/vote"
	"github.com/zekurio/daemon/pkg/perms"
)

// Database is the interface for our database service
// which is then implemented by either postgres or nutsdb
type Database interface {
	Close() error

	// Guild settings

	GetAutoRoles(guildID string) ([]string, error)
	SetAutoRoles(guildID string, roleIDs []string) error

	GetAutoVoice(guildID string) ([]string, error)
	SetAutoVoice(guildID string, channelIDs []string) error

	GetCreatedAV(guildID string) ([]string, error)
	SetCreatedAV(guildID string, channelIDs []string) error

	// Permissions

	GetPermissions(guildID string) (map[string]perms.PermsArray, error)
	SetPermissions(guildID, roleID string, perms perms.PermsArray) error

	// Role selection

	AddRoleSelections(v []models.RoleSelection) error
	GetRoleSelections() ([]models.RoleSelection, error)
	RemoveRoleSelections(guildID, channelID, messageID string) error

	// Quotes
	// TODO

	// Votes

	GetVotes() (map[string]vote.Vote, error)
	AddUpdateVote(vote vote.Vote) error
	DeleteVote(voteID string) error

	// Data management

	FlushGuildData(guildID string) error
}
