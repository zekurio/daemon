package perms

// Array defines the permissions a given user or role has.PermsArray
// A permission looks something like that:
// "+admin.kick"
// "+admin.ban"
// "-user.skip"
// Which gives the user explicit permissions to commands that
// that only require the kick. ban permission
// and disallows the usage of the skip command for music related commands
// while "+admin.*" gives the user permission to any admin action
type Array []string

// Update updates the permission array with the given permission, it respects the prefix of the permission
// and will add or remove the permission from the array accordingly. Returns the new permission array
// and a boolean indicating if the permission was overridden or not.
func (p Array) Update(newPerm string, override bool) (newArray Array, overridden bool) {
	newArray = make(Array, len(p)+1)

	i := 0
	add := true
	for _, perm := range p {
		//
		if len(perm) > 1 && newPerm[1:] == perm[1:] {
			add = false

			if override {
				newArray[i] = newPerm
				i++
				continue
			}

			if perm[0] != newPerm[0] {
				continue
			}
		}

		newArray[i] = perm
		i++
	}

	if add {
		newArray[i] = newPerm
		i++
	}

	newArray = newArray[:i]

	overridden = !p.Equals(newArray)

	return
}

// Equals checks if the permission array is equal to the given permission array
func (p Array) Equals(other Array) bool {
	if len(p) != len(other) {
		return false
	}

	for i, perm := range p {
		if perm != other[i] {
			return false
		}
	}

	return true
}

// Merge merges the given permission array with the current permission array
func (p Array) Merge(newPerms Array, override bool) Array {
	for _, cp := range newPerms {
		p, _ = p.Update(cp, override)
	}
	return p
}

// Has checks if the permission array has the given permission
func (p Array) Has(neededPerm string) bool {
	match := -1
	allow := false

	for _, perm := range p {
		m, a := checkPermission(neededPerm, perm)
		if m > match {
			allow, match = a, m
		}
	}

	return allow

}
