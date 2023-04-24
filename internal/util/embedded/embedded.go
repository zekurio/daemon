package embedded

import (
	"embed"
	"strings"
)

var (

	//go:embed Version.txt
	AppVersion string
	//go:embed Commit.txt
	AppCommit string
	//go:embed Release.txt
	Release string
	//go:embed migrations
	Migrations embed.FS
)

func IsRelease() bool {
	return strings.ToLower(Release) == "true"
}
