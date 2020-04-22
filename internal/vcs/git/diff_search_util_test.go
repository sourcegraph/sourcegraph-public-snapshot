package git

import (
	"strings"
	"testing"
)

func TestRegexpToGlobBestEffort(t *testing.T) {
	tests := map[string]struct {
		glob  string
		equiv bool
	}{
		"":          {"*", true},
		"foo":       {"*foo*", true},
		"^foo":      {"foo*", true},
		`foo\.js`:   {"*foo.js*", true},
		"foo.js":    {"*foo?js*", true},
		"foo.*js":   {"*foo*js*", true},
		"^fo.o":     {"fo?o*", true},
		"foo$":      {"*foo", true},
		"fo.o$":     {"*fo?o", true},
		"^foo$":     {"foo", true},
		"^foo|bar$": {"", false},
		".*js":      {"*js*", true},
		"^.*js":     {"*js*", true},
		"foo.*":     {"*foo*", true},
		"foo.*$":    {"*foo*", true},

		// need to escape *?[]\
		// Note: We could just prefix with :(literal)
		`foo\*bar`: {`*foo\*bar*`, true},
		`foo\?bar`: {`*foo\?bar*`, true},
		`foo\[bar`: {`*foo\[bar*`, true},
		`foo\]bar`: {`*foo\]bar*`, true},
		`foo\\bar`: {`*foo\\bar*`, true},
		`\*bar`:    {`*\*bar*`, true},
		`\?bar`:    {`*\?bar*`, true},
		`\[bar`:    {`*\[bar*`, true},
		`\]bar`:    {`*\]bar*`, true},
		`\\bar`:    {`*\\bar*`, true},
		`foo\*`:    {`*foo\**`, true},
		`foo\?`:    {`*foo\?*`, true},
		`foo\[`:    {`*foo\[*`, true},
		`foo\]`:    {`*foo\]*`, true},
		`foo\\`:    {`*foo\\*`, true},

		`^foo\*bar`:  {`foo\*bar*`, true},
		`^foo\*bar$`: {`foo\*bar`, true},
		`foo\*bar$`:  {`*foo\*bar`, true},

		// leading : has special meaning, lets just ignore them.
		"^:": {"", false},
		":":  {"*:*", true},
		":$": {"*:", true},

		// : anywhere else is fine
		"foo:": {"*foo:*", true},

		// everything upto the last "/" is regarded as a path prefix and is
		// not passed to fnmatch. What isn't documented is that glob chars are
		// part of what helps decide the path prefix. So our normal wildcard
		// logic works.
		"foo/bar/baz":   {"*foo/bar/baz*", true},
		"^foo/bar/baz":  {"foo/bar/baz*", true},
		"^foo/bar/baz$": {"foo/bar/baz", true},
		"foo/bar/baz$":  {"*foo/bar/baz", true},
	}
	for pat, want := range tests {
		t.Run(pat, func(t *testing.T) {
			glob, equiv := regexpToGlobBestEffort(pat)
			if glob != want.glob {
				t.Errorf("got glob %q, want %q", glob, want.glob)
			}
			// If functionaltiy to handle leading : globs is added, be sure
			// to update the code that handles matching paths case-insensitively.
			// Update it by removing the code that concatenates :(icase) to the resulting glob.
			if strings.HasPrefix(glob, ":") {
				t.Errorf("got glob %q with ':' as a prefix. Callers expect regexpToGlobBestEffort to not return ':' as a prefix", glob)
			}
			if equiv != want.equiv {
				t.Errorf("got equiv %v, want %v", equiv, want.equiv)
			}
		})
	}
}
