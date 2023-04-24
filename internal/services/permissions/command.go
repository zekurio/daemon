package permissions

type CommandPerms interface {
	// Perm is the permission required to run the command.
	Perm() string
	// SubPerms are the permissions to run the sub commands.
	SubPerms() []SubCommandPerms
}

type SubCommandPerms struct {
	Perm        string
	Explicit    bool
	Description string
}
