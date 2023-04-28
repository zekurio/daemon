package stringutils

import (
	"regexp"
	"strings"
)

var (
	rxNumber = regexp.MustCompile(`^-?\d+$`)
	rxBool   = regexp.MustCompile(`^(true|false)$`)
)

// IsInteger returns true if the passed string is
// a valid number.
func IsInteger(str string) bool {
	return rxNumber.MatchString(str)
}

// IsBool returns true if the passed string is
// a valid bool.
func IsBool(str string) bool {
	return rxBool.MatchString(str)
}

// EnsureNotEmpty returns def if str is empty.
func EnsureNotEmpty(str, def string) string {
	if str == "" {
		return def
	}
	return str
}

// FromBool returns ifTrue if cond is true
// else returns ifFalse.
func FromBool(cond bool, ifTrue, ifFalse string) string {
	if cond {
		return ifTrue
	}
	return ifFalse
}

// HasPrefixAny returns true if the given str has
// any of the given prefixes.
func HasPrefixAny(str string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(str, p) {
			return true
		}
	}

	return false
}

// HasSuffixAny returns true if the given str has
// any of the given suffixes.
func HasSuffixAny(str string, suffixes ...string) bool {
	for _, s := range suffixes {
		if strings.HasSuffix(str, s) {
			return true
		}
	}

	return false
}

// HasPatternAny returns true if the given str has
// any of the given patterns.
func HasPatternAny(str string, patterns ...string) bool {
	for _, p := range patterns {
		if strings.Contains(str, p) {
			return true
		}
	}

	return false
}

// Capitalize uppercase's the first character of the
// given string.
//
// If all is true, all starting characters of all
// words in the string are capitalized.
func Capitalize(v string, all bool) string {
	if v == "" {
		return ""
	}
	if all {
		split := strings.Split(v, " ")
		for i, v := range split {
			split[i] = Capitalize(v, false)
		}
		return strings.Join(split, " ")
	} else {
		return strings.ToUpper(string(v[0])) + v[1:]
	}
}
