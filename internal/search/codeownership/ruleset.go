package codeownership

import (
	"os"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

type Ruleset struct {
	file *codeownerspb.File
}

// TODO support sections
func (r *Ruleset) Match(path string) ([]*codeownerspb.Owner, error) {
	for _, r := range r.file.GetRule() {
		if p := pattern(r.GetPattern()); p.match(path) {
			return r.GetOwner(), nil
		}
	}
	return nil, nil
}

type pattern string

func (p pattern) String() string {
	return string(p)
}

func (p pattern) match(path string) bool {
	// left anchored
	if !strings.ContainsAny(p.String(), `*?\`) && p.String()[0] == os.PathSeparator {
		prefix := p.String()

		// Strip the leading slash as we're anchored to the root already
		if prefix[0] == os.PathSeparator {
			prefix = prefix[1:]
		}

		// If the pattern ends with a slash we can do a simple prefix match
		if prefix[len(prefix)-1] == os.PathSeparator {
			return strings.HasPrefix(path, prefix)
		}

		// If the strings are the same length, check for an exact match
		if len(path) == len(prefix) {
			return path == prefix
		}

		// Otherwise check if the test path is a subdirectory of the pattern
		if len(path) > len(prefix) && path[len(prefix)] == os.PathSeparator {
			return path[:len(prefix)] == prefix
		}
		return false
	}
	re, err := p.regex()
	if err != nil {
		return false
	}
	return re.MatchString(path)
}

func (p pattern) regex() (*regexp.Regexp, error) {
	// Handle specific edge cases first
	switch {
	case strings.Contains(p.String(), "***"):
		return nil, errors.Errorf("pattern cannot contain three consecutive asterisks")
	case p.String() == "":
		return nil, errors.Errorf("empty pattern")
	case p.String() == "/":
		// "/" doesn't match anything
		return regexp.Compile(`\A\z`)
	}

	segs := strings.Split(p.String(), "/")

	if segs[0] == "" {
		// Leading slash: match is relative to root
		segs = segs[1:]
	} else {
		// No leading slash - check for a single segment pattern, which matches
		// relative to any descendent path (equivalent to a leading **/)
		if len(segs) == 1 || (len(segs) == 2 && segs[1] == "") {
			if segs[0] != "**" {
				segs = append([]string{"**"}, segs...)
			}
		}
	}

	if len(segs) > 1 && segs[len(segs)-1] == "" {
		// Trailing slash is equivalent to "/**"
		segs[len(segs)-1] = "**"
	}

	sep := string(os.PathSeparator)

	lastSegIndex := len(segs) - 1
	needSlash := false
	var re strings.Builder
	re.WriteString(`\A`)
	for i, seg := range segs {
		switch seg {
		case "**":
			switch {
			case i == 0 && i == lastSegIndex:
				// If the pattern is just "**" we match everything
				re.WriteString(`.+`)
			case i == 0:
				// If the pattern starts with "**" we match any leading path segment
				re.WriteString(`(?:.+` + sep + `)?`)
				needSlash = false
			case i == lastSegIndex:
				// If the pattern ends with "**" we match any trailing path segment
				re.WriteString(sep + `.*`)
			default:
				// If the pattern contains "**" we match zero or more path segments
				re.WriteString(`(?:` + sep + `.+)?`)
				needSlash = true
			}

		case "*":
			if needSlash {
				re.WriteString(sep)
			}

			// Regular wildcard - match any characters except the separator
			re.WriteString(`[^` + sep + `]+`)
			needSlash = true

		default:
			if needSlash {
				re.WriteString(sep)
			}

			escape := false
			for _, ch := range seg {
				if escape {
					escape = false
					re.WriteString(regexp.QuoteMeta(string(ch)))
					continue
				}

				// Other pathspec implementations handle character classes here (e.g.
				// [AaBb]), but CODEOWNERS doesn't support that so we don't need to
				switch ch {
				case '\\':
					escape = true
				case '*':
					// Multi-character wildcard
					re.WriteString(`[^` + sep + `]*`)
				case '?':
					// Single-character wildcard
					re.WriteString(`[^` + sep + `]`)
				default:
					// Regular character
					re.WriteString(regexp.QuoteMeta(string(ch)))
				}
			}

			if i == lastSegIndex {
				// As there's no trailing slash (that'd hit the '**' case), we
				// need to match descendent paths
				re.WriteString(`(?:` + sep + `.*)?`)
			}

			needSlash = true
		}
	}
	re.WriteString(`\z`)
	return regexp.Compile(re.String())
}
