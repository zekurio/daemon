package autovoice

import "github.com/bwmarrin/discordgo"

type AutovoiceProvider interface {
	// Join is called when a user joins a voice channel
	// it handles all the logic for creating a new channel
	// or handling people joining a created auto voice channel
	Join(s *discordgo.Session, vs *discordgo.VoiceState) (err error)

	// Leave is called when a user leaves a voice channel
	// it handles all the logic for deleting a channel
	// or handling people leaving a created auto voice channel
	Leave(s *discordgo.Session, vs *discordgo.VoiceState) (err error)

	// Move is called when a user moves from one voice channel to another
	// it handles all the logic for deleting a channel
	// or handling people leaving a created auto voice channel
	Move(s *discordgo.Session, vsNew, vsOld *discordgo.VoiceState) (err error)

	// Deconstruct saves all the autovoice channels to the database when daemon is shut down.
	Deconstruct() error
}
