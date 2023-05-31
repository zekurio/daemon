package autovoice

import "github.com/zekurio/daemon/internal/models"

type AutovoiceProvider interface {
	// SetGuilds overwrites the complete guilds map with a new one
	SetGuilds(guildMap map[string]models.GuildMap)

	// CreateChannel handles creating a new autovoice channel
	// and adding it to the guild map, it also handles
	// saving to the database
	CreateChannel(guildID, ownerID, parentID string) (*models.AVChannel, error)

	// DeleteChannel handles deleting an autovoice channel
	// and removing it from the guild map, it also handles
	// removing it from the database
	DeleteChannel(guildID, channelID string) error

	// SwapOwner handles swapping the owner of an autovoice channel
	// in case the owner leaves the channel
	SwapOwner(guildID, newOwner, channelID string) error

	// Deconstruct deconstructs the autovoice service,
	// saves the guild map to the database
	Deconstruct() error

	// GetChannelFromOwner returns the AVChannel struct
	// from the guild map based on the owner ID
	GetChannelFromOwner(guildID, ownerID string) (*models.AVChannel, error)

	// IsCreatedChannel returns true if the channel ID is a created autovoice channel
	IsCreatedChannel(guildID, channelID string) bool

	// CurrentChannels returns the currently active autovoice channels in a guild
	CurrentChannels(guildID string) (channels []*models.AVChannel, err error)

	// AddMember adds a member to an autovoice channel
	AddMember(guildID, userID, channelID string) (err error)

	// RemoveMember removes a member from an autovoice channel
	RemoveMember(guildID, userID, channelID string) (err error)
}
