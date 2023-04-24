package models

type RoleRewards struct {
	GuildID     string
	RewardRoles []RewardRole
	RemoveOld   bool
}

type RewardRole struct {
	RoleID string
	Amount int
}
