package util

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/xid"
	"github.com/zekrotja/ken"

	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/arrayutils"
)

func AttachRoleSelectButtons(b *ken.ComponentBuilder, roles []*discordgo.Role) ([]string, error) {
	type roleButton struct {
		Button *discordgo.Button
		RoleID string
	}

	roleButtons := map[string]*discordgo.Button{}
	for _, role := range roles {
		roleButtons[role.ID] = &discordgo.Button{
			Label:    role.Name,
			Style:    discordgo.PrimaryButton,
			CustomID: xid.New().String(),
		}

	}

	nCols := len(roleButtons) / 5
	if len(roleButtons)%5 > 0 {
		nCols++
	}

	roleButtonsColumns := make([][]roleButton, nCols)
	roleIDs := make([]string, 0, len(roleButtons))
	i := 0
	for id, b := range roleButtons {
		roleButtonsColumns[i/5] = append(roleButtonsColumns[i/5], roleButton{
			Button: b,
			RoleID: id,
		})
		roleIDs = append(roleIDs, id)
		i++
	}

	for _, rbs := range roleButtonsColumns {
		b.AddActionsRow(func(b ken.ComponentAssembler) {
			for _, rb := range rbs {
				b.Add(rb.Button, OnRoleSelect(rb.RoleID))
			}
		})
	}

	_, err := b.Build()

	return roleIDs, err
}

func OnRoleSelect(roleID string) func(ctx ken.ComponentContext) bool {
	return func(ctx ken.ComponentContext) bool {
		ctx.SetEphemeral(true)
		ctx.Defer()

		if arrayutils.Contains(ctx.GetEvent().Member.Roles, roleID) {
			err := ctx.GetSession().GuildMemberRoleRemove(ctx.GetEvent().GuildID, ctx.User().ID, roleID)
			if err != nil {
				err = ctx.FollowUpError("Failed removing role.", "").
					Send().
					DeleteAfter(10 * time.Second).Error
				return err == nil
			}
			err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
				Color:       static.ColorGreen,
				Description: fmt.Sprintf("Role <@&%s> has been removed.", roleID),
			}).Send().DeleteAfter(10 * time.Second).Error
			return err == nil
		}

		err := ctx.GetSession().GuildMemberRoleAdd(ctx.GetEvent().GuildID, ctx.User().ID, roleID)
		if err != nil {
			err = ctx.FollowUpError("Failed adding role.", "").
				Send().
				DeleteAfter(10 * time.Second).Error
			return err == nil
		}

		err = ctx.FollowUpEmbed(&discordgo.MessageEmbed{
			Color:       static.ColorGreen,
			Description: fmt.Sprintf("Role <@&%s> has been added.", roleID),
		}).Send().DeleteAfter(10 * time.Second).Error

		return err == nil
	}
}
