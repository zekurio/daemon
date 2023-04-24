package stringutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInteger(t *testing.T) {
	if !IsInteger("123") {
		t.Error("not detected as integer")
	}

	if !IsInteger("-123") {
		t.Error("not detected as integer")
	}

	if IsInteger("") {
		t.Error("detected as integer")
	}

	if IsInteger("1a3") {
		t.Error("detected as integer")
	}
}

func TestEnsureNotEmpty(t *testing.T) {
	const (
		val = "foo"
		def = "bar"
	)

	if EnsureNotEmpty("", def) != def {
		t.Error("did not return default string")
	}

	if EnsureNotEmpty(val, def) != val {
		t.Error("did not return value string")
	}

	if EnsureNotEmpty("", "An error occurred.") != "An error occurred." {
		t.Error("return value was not empty")
	}
}

func TestFromBool(t *testing.T) {
	const (
		tr = "true"
		fa = "false"
	)

	if FromBool(true, tr, fa) != tr {
		t.Error("true does not return true string")
	}

	if FromBool(false, tr, fa) != fa {
		t.Error("false does not return false string")
	}
}

func TestHasPrefixAny(t *testing.T) {
	if !HasPrefixAny("foo", "a", "fo") {
		t.Error("falsely detected has no prefix")
	}

	if !HasPrefixAny("foo", "a", "b", "f") {
		t.Error("falsely detected has no prefix")
	}

	if HasPrefixAny("bar", "a", "f") {
		t.Error("falsely detected has prefix")
	}

	if HasPrefixAny("foo") {
		t.Error("falsely detected has prefix")
	}

	if HasPrefixAny("", "a", "f") {
		t.Error("falsely detected has prefix")
	}
}

func TestHasSuffixAny(t *testing.T) {
	if !HasSuffixAny("foo", "a", "oo") {
		t.Error("falsely detected has no prefix")
	}

	if !HasSuffixAny("foo", "a", "b", "o") {
		t.Error("falsely detected has no prefix")
	}

	if HasSuffixAny("foo", "a", "b") {
		t.Error("falsely detected has prefix")
	}

	if HasSuffixAny("foo") {
		t.Error("falsely detected has prefix")
	}

	if HasSuffixAny("", "a", "b") {
		t.Error("falsely detected has prefix")
	}
}

func TestHasPatternAny(t *testing.T) {
	if !HasPatternAny("foo", "a", "fo") {
		t.Error("falsely detected has no pattern")
	}

	if !HasPatternAny("foo", "a", "b", "o") {
		t.Error("falsely detected has no pattern")
	}

	if HasPatternAny("foo", "a", "b") {
		t.Error("falsely detected has pattern")
	}

	if HasPatternAny("foo") {
		t.Error("falsely detected has pattern")
	}

	if HasPatternAny("", "a", "b") {
		t.Error("falsely detected has pattern")
	}
}

func TestCapitalize(t *testing.T) {
	assert.Equal(t, "", Capitalize("", false))
	assert.Equal(t, "", Capitalize("", true))

	assert.Equal(t, "F", Capitalize("f", false))
	assert.Equal(t, "F", Capitalize("f", true))

	assert.Equal(t, "Foo", Capitalize("foo", false))
	assert.Equal(t, "Foo", Capitalize("foo", true))

	assert.Equal(t, "Foo bar", Capitalize("foo bar", false))
	assert.Equal(t, "Foo Bar", Capitalize("foo bar", true))

}
