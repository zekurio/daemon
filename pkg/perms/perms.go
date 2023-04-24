package perms

import (
	"strings"
)

var maxMatch = 2 ^ 10

// matchPermission checks if the given permission matches the given permission
func matchPermission(neededPerm, perm string) int {

	if neededPerm == "" {
		return -1
	}

	var explicit bool

	if neededPerm[0] == '!' {
		explicit = true
		neededPerm = neededPerm[1:]
	}

	if neededPerm == perm {
		return maxMatch
	}

	if explicit {
		return -1
	}

	// Check for wildcards now
	toCheckSegments := strings.Split(neededPerm, ".")
	assembledPerm := ""
	for i, segment := range toCheckSegments {
		if assembledPerm == "" {
			assembledPerm = segment
		} else {
			assembledPerm += "." + segment
		}

		// check if the perm matches with the assembled perm
		// including the wildcard
		if perm == assembledPerm+".*" {
			return i
		}

	}

	return -1

}

// checkPermission checks how well the given permission matches the given permission
// and returns the match and if the permission is allowed
func checkPermission(neededPerm, perm string) (int, bool) {
	if perm == "" {
		return -1, false
	}

	if perm[0] != '+' && perm[0] != '-' {
		return -1, false
	}

	match := matchPermission(neededPerm, perm[1:])
	if match < 0 {
		return match, false
	}

	return match, !strings.HasPrefix(perm, "-")
}
