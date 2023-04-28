package discordutils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zekurio/daemon/internal/util/static"

	"github.com/bwmarrin/discordgo"
)

// GetInviteLink returns invite link for the app
func GetInviteLink(session *discordgo.Session) string {
	return fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&scope=%s&permissions=%d",
		session.State.User.ID, static.OAuthScopes, static.InvitePermission)
}

// GetDiscordSnowflakeCreationTime returns time when snowflake was created
func GetDiscordSnowflakeCreationTime(snowflake string) (time.Time, error) {
	sfI, err := strconv.ParseInt(snowflake, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	timestamp := (sfI >> 22) + 1420070400000
	return time.Unix(timestamp/1000, timestamp), nil
}

// SendMessageDM sends message to user
func SendMessageDM(session *discordgo.Session, userID, message string) (msg *discordgo.Message, err error) {
	ch, err := session.UserChannelCreate(userID)
	if err != nil {
		return
	}
	msg, err = session.ChannelMessageSend(ch.ID, message)
	return
}

// SendEmbedMessageDM sends embed message to user
func SendEmbedMessageDM(session *discordgo.Session, userID string, embed *discordgo.MessageEmbed) (msg *discordgo.Message, err error) {
	ch, err := session.UserChannelCreate(userID)
	if err != nil {
		return
	}
	msg, err = session.ChannelMessageSendEmbed(ch.ID, embed)
	return
}

// GetMember returns member from guild
func GetMember(session *discordgo.Session, guildID, userID string) (*discordgo.Member, error) {
	member, err := session.State.Member(guildID, userID)
	if err == nil {
		return member, nil
	}

	member, err = session.GuildMember(guildID, userID)
	return member, err
}

// GetGuild returns guild from session
func GetGuild(session *discordgo.Session, id string) (*discordgo.Guild, error) {
	guild, err := session.State.Guild(id)
	if err == nil {
		return guild, nil
	}

	guild, err = session.Guild(id)
	return guild, err
}

// GetChannel returns channel from session
func GetChannel(session *discordgo.Session, id string) (*discordgo.Channel, error) {
	channel, err := session.State.Channel(id)
	if err == nil {
		return channel, nil
	}

	channel, err = session.Channel(id)
	return channel, err
}

// GetUser returns user from session
func GetUser(session *discordgo.Session, id string) (*discordgo.User, error) {
	user, err := session.User(id)
	return user, err
}

// UsersInGuildVoice returns users in guild voice channels
func UsersInGuildVoice(session *discordgo.Session, guildID string) ([]string, error) {
	g, err := session.State.Guild(guildID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(g.VoiceStates))
	for _, vs := range g.VoiceStates {
		if vs.UserID != session.State.User.ID {
			userIDs = append(userIDs, vs.UserID)
		}
	}

	return userIDs, nil
}

// FindUserVS returns a users voice states
func FindUserVS(session *discordgo.Session, userID string) (discordgo.VoiceState, bool) {
	for _, g := range session.State.Guilds {
		for _, vs := range g.VoiceStates {
			for vs.UserID == userID {
				return *vs, true
			}
		}
	}
	return discordgo.VoiceState{}, false
}

// FindGuildTextChannel returns the first text channel in a guild.
func FindGuildTextChannel(session *discordgo.Session, guildID string) *discordgo.Channel {
	g, err := GetGuild(session, guildID)
	if err != nil {
		return nil
	}

	for _, c := range g.Channels {
		if c.Type == discordgo.ChannelTypeGuildText {
			return c
		}
	}

	return nil
}

// IsAdmin checks if the given member has the admin permission set.
func IsAdmin(guild *discordgo.Guild, member *discordgo.Member) bool {
	if member == nil || guild == nil {
		return false
	}

	for _, r := range guild.Roles {
		if r.Permissions&0x8 != 0 {
			for _, mrID := range member.Roles {
				if r.ID == mrID {
					return true
				}
			}
		}
	}

	return false
}

// IsAFKChannel checks if the given channel is the guilds AFK channel.
func IsAFKChannel(session *discordgo.Session, guildID, channelID string) bool {
	g, err := GetGuild(session, guildID)
	if err != nil {
		return false
	}

	return g.AfkChannelID == channelID || err != nil
}

// GetVoiceMembers returns all members in a voice channel.
func GetVoiceMembers(session *discordgo.Session, guildID, channelID string) ([]*discordgo.Member, error) {

	g, err := session.State.Guild(guildID)
	if err != nil {
		return nil, err
	}

	var members []*discordgo.Member

	for _, vs := range g.VoiceStates {
		if vs.ChannelID == channelID {
			m, err := GetMember(session, guildID, vs.UserID)
			if err != nil {
				continue
			}

			members = append(members, m)
		}
	}

	return members, nil

}

// GetMessages returns messages from channel
func GetMessages(session *discordgo.Session, channelID string, limit int) ([]*discordgo.Message, error) {
	var messages []*discordgo.Message

	for {
		ms, err := session.ChannelMessages(channelID, limit, "", "", "")
		if err != nil {
			return nil, err
		}

		messages = append(messages, ms...)

		if len(ms) < limit {
			break
		}
	}

	return messages, nil
}

// GetMessage returns a message by its ID
func GetMessage(session *discordgo.Session, channelID, messageID string) (*discordgo.Message, error) {

	m, err := session.State.Message(channelID, messageID)
	if err == nil {
		return m, nil
	}

	m, err = session.ChannelMessage(channelID, messageID)
	return m, err
}

// GetMessageLink returns a message link
func GetMessageLink(msg *discordgo.Message, guildID string) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, msg.ChannelID, msg.ID)
}
