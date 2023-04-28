package permissions

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"

	"github.com/zekurio/daemon/pkg/perms"
)

type PermsProvider interface {
	ken.MiddlewareBefore

	// GetPerms collects the permissions of a user from their roles.
	GetPerms(session *discordgo.Session, guildID, userID string) (perm perms.Array, override bool, err error)

	// GetMemberPerms collects the permissions of a member from their roles.
	GetMemberPerms(session *discordgo.Session, guildID string, memberID string) (perms.Array, error)

	// HasPerms checks if a user has the given permission.
	HasPerms(session *discordgo.Session, guildID, userID, perm string) (ok, override bool, err error)

	// HasSubCmdPerms checks if a user has the given permission for a subcommand.
	HasSubCmdPerms(ctx ken.Context, subPM string, explicit bool, message ...string) (ok bool, err error)
}
