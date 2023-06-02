package database

import (
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/util/vote"
	"github.com/zekurio/daemon/pkg/perms"
)

// Database is the interface for our database service
// which is then implemented by postgres
type Database interface {
	Close() error

	// Guild settings

	GetAutoRoles(guildID string) ([]string, error)
	SetAutoRoles(guildID string, roleIDs []string) error

	GetAutoVoice(guildID string) ([]string, error)
	SetAutoVoice(guildID string, channelIDs []string) error

	// Permissions

	GetPermissions(guildID string) (map[string]perms.Array, error)
	SetPermissions(guildID, roleID string, perms perms.Array) error

	// TODO Votes

	GetVotes() (map[string]vote.Vote, error)
	AddUpdateVote(vote vote.Vote) error
	DeleteVote(voteID string) error

	// AutoVoice

	GetGuildMap(guildID string) (models.ChannelMap, error)
	AddUpdateGuildMap(guildID string, channelMap models.ChannelMap) error
	GetAllGuildMaps() (map[string]models.ChannelMap, error)
	DeleteGuildMap(guildID string) error

	// Data management

	FlushGuildData(guildID string) error
}
