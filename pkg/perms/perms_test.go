package perms

import "testing"

func TestMatchPermission(t *testing.T) {
	testMatch := func(neededPerm, perm string, exp int) {
		if res := matchPermission(neededPerm, perm); res != exp {
			t.Errorf("%s -> %s : result was %d (expected %d)",
				neededPerm, perm, res, exp)
		}
	}

	testMatch("foo.bar.baz", "foo.bar.baz", maxMatch)
	testMatch("foo.bar.baz", "foo.bar.*", 1)
	testMatch("foo.bar.baz", "foo.*", 0)

	testMatch("!foo.bar.baz", "foo.bar.baz", maxMatch)
	testMatch("!foo.bar.baz", "foo.bar.*", -1)
	testMatch("!foo.bar.baz", "foo.*", -1)

}

func TestCheckPermission(t *testing.T) {
	testCheck := func(neededPerm, perm string, expMatch int, expPass bool) {
		if match, pass := checkPermission(neededPerm, perm); match != expMatch || pass != expPass {
			t.Errorf("%s -> %s : result was %d / %t (expected %d / %t)",
				neededPerm, perm, match, pass, expMatch, expPass)
		}
	}

	testCheck("foo.bar.baz", "+foo.bar.baz", maxMatch, true)
	testCheck("foo.bar.baz", "-foo.bar.baz", maxMatch, false)

	testCheck("foo.bar.baz", "+foo.bar.*", 1, true)
	testCheck("foo.bar.baz", "-foo.bar.*", 1, false)

	testCheck("foo.bar.baz", "+foo.*", 0, true)
	testCheck("foo.bar.baz", "-foo.*", 0, false)

	testCheck("foo.bar.baz", "foo.bar.baz", -1, false)

}
