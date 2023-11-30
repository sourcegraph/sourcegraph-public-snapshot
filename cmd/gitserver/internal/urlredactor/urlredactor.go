package urlredactor

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// URLRedactor redacts all sensitive strings from a message.
type URLRedactor struct {
	// sensitive are sensitive strings to be redacted.
	// The strings should not be empty.
	sensitive []string
}

// New returns a new urlRedactor that redacts credentials found in rawurl, and
// the rawurl itself.
func New(parsedURL *vcs.URL) *URLRedactor {
	var sensitive []string
	pw, _ := parsedURL.User.Password()
	u := parsedURL.User.Username()
	if pw != "" && u != "" {
		// Only block password if we have both as we can
		// assume that the username isn't sensitive in this case
		sensitive = append(sensitive, pw)
	} else {
		if pw != "" {
			sensitive = append(sensitive, pw)
		}
		if u != "" {
			sensitive = append(sensitive, u)
		}
	}
	sensitive = append(sensitive, parsedURL.String())
	return &URLRedactor{sensitive: sensitive}
}

// Redact returns a redacted version of message.
// Sensitive strings are replaced with "<redacted>".
func (r *URLRedactor) Redact(message string) string {
	for _, s := range r.sensitive {
		message = strings.ReplaceAll(message, s, "<redacted>")
	}
	return message
}
