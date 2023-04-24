package permissions

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"github.com/zekrotja/ken"

	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/discordutils"
	"github.com/zekurio/daemon/pkg/perms"
	"github.com/zekurio/daemon/pkg/roleutils"
)

type Permissions struct {
	db  database.Database
	cfg models.Config
	s   *discordgo.Session
}

var _ PermsProvider = (*Permissions)(nil)

func InitPermissions(ctn di.Container) *Permissions {
	return &Permissions{
		db:  ctn.Get(static.DiDatabase).(database.Database),
		cfg: ctn.Get(static.DiConfig).(models.Config),
		s:   ctn.Get(static.DiDiscord).(*discordgo.Session),
	}
}

func (p *Permissions) Before(ctx *ken.Ctx) (next bool, err error) {
	cmd, ok := ctx.Command.(CommandPerms)
	if !ok {
		next = true
		return
	}

	if ctx.User() == nil {
		return
	}

	ok, err = p.HasPerms(ctx.GetSession(), ctx.GetEvent().GuildID, ctx.User().ID, cmd.Perm())

	if err != nil {
		return false, err
	}

	if !ok {
		err = ctx.RespondError("You are not permitted to use this command!", "Missing Permission")
		return
	}

	next = true
	return
}

func (p *Permissions) HasPerms(session *discordgo.Session, guildID, userID, dn string) (ok bool, err error) {
	perms, err := p.GetPerms(session, guildID, userID)
	if err != nil {
		return false, err
	}

	return perms.Has(dn), nil
}

func (p *Permissions) GetPerms(session *discordgo.Session, guildID, userID string) (perm perms.PermsArray, err error) {

	if guildID != "" {
		guild, err := discordutils.GetGuild(session, guildID)
		if err != nil {
			return perms.PermsArray{}, nil
		}

		member, err := discordutils.GetMember(session, guildID, userID)
		if err != nil {
			return perms.PermsArray{}, nil
		}

		if userID == guild.OwnerID || (member != nil && discordutils.IsAdmin(guild, member)) {
			var defAdminRoles []string
			defAdminRoles = p.cfg.Permissions.AdminRules
			if defAdminRoles == nil {
				defAdminRoles = static.DefaultAdminRules
			}

			perm = perm.Merge(defAdminRoles, false)
		}

		memberPerms, err := p.GetMemberPerms(session, guildID, userID)
		if err == nil {
			perm = perm.Merge(memberPerms, true)
		}
	}

	var defUserRoles []string
	defUserRoles = p.cfg.Permissions.UserRules
	if defUserRoles == nil {
		defUserRoles = static.DefaultUserRules
	}

	if userID == p.cfg.Discord.OwnerID {
		perm = perms.PermsArray{"+dm.*"}
	}

	perm = perm.Merge(defUserRoles, false)

	return perm, nil
}

func (p *Permissions) GetMemberPerms(session *discordgo.Session, guildID string, memberID string) (perms.PermsArray, error) {
	guildPerms, err := p.db.GetPermissions(guildID)
	if err != nil {
		return nil, err
	}
	membRoles, err := roleutils.GetSortedMemberRoles(session, guildID, memberID, false, true)
	if err != nil {
		return nil, err
	}

	var res perms.PermsArray
	for _, r := range membRoles {
		if p, ok := guildPerms[r.ID]; ok {
			if res == nil {
				res = p
			} else {
				res = res.Merge(p, true)
			}
		}
	}

	return res, nil
}
