package grafana

import (
	"strings"
	"unicode"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ValidateUID checks if the given string is a valid UID for entry into a Grafana dashboard. This is
// primarily used in the URL, e.g. /-/debug/grafana/d/syntect-server/<UID> and allows us to have
// static URLs we can document like:
//
//	Go to https://sourcegraph.example.com/-/debug/grafana/d/syntect-server/syntect-server
//
// Instead of having to describe all the steps to navigate there because the UID is random.
func ValidateUID(s string) error {
	const lengthLimit = 40
	if len(s) > lengthLimit {
		return errors.Newf("UID must be less than %d characters", lengthLimit)
	}
	if s != strings.ToLower(s) {
		return errors.New("UID must be all lowercase characters")
	}
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-') {
			return errors.New("UID contains illegal character")
		}
	}
	return nil
}
