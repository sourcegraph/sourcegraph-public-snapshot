package redactor

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// urlRedactor redacts all sensitive strings from a message.
type urlRedactor struct {
	// sensitive are sensitive strings to be redacted.
	// The strings should not be empty.
	sensitive []string
}

// NewURLRedactor returns a new urlRedactor that redacts
// credentials found in rawurl, and the rawurl itself.
func NewURLRedactor(parsedURL *vcs.URL) *urlRedactor {
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
	return &urlRedactor{sensitive: sensitive}
}

// redact returns a redacted version of message.
// Sensitive strings are replaced with "<redacted>".
func (r *urlRedactor) Redact(message string) string {
	for _, s := range r.sensitive {
		message = strings.ReplaceAll(message, s, "<redacted>")
	}
	return message
}
