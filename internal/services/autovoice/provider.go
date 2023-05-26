package autovoice

import "github.com/zekurio/daemon/internal/models"

type AutovoiceProvider interface {
	// AddGuild handles adding a guild to the guild map
	// and creating the database entry for it
	AddGuild(guildID string) error

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

	// GetChannelFromOrigin returns the AVChannel struct
	// from the guild map based on the origin channel ID
	GetChannelFromOrigin(guildID, originID string) (*models.AVChannel, error)

	// GetChannelFromOwner returns the AVChannel struct
	// from the guild map based on the owner ID
	GetChannelFromOwner(guildID, ownerID string) (*models.AVChannel, error)
}
