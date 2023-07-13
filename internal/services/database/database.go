package database

import (
	"github.com/zekurio/daemon/internal/models"
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

	// Votes

	GetVotes() (map[string]models.Vote, error)
	AddUpdateVote(vote models.Vote) error
	DeleteVote(voteID string) error

	// AutoVoice

	/* this is currently not useful, need a way to keep track of
	   	the channel map while the bot is shut down1
		GetGuildMap(guildID string) (models.ChannelMap, error)
		AddUpdateGuildMap(guildID string, channelMap models.ChannelMap) error
		GetAllGuildMaps() (map[string]models.ChannelMap, error)
		DeleteGuildMap(guildID string) error
	*/

	// Data management

	FlushGuildData(guildID string) error
}
