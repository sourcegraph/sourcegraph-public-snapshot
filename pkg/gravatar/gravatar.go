// Package gravatar generates Gravatar avatar URLs.
package gravatar

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
)

// URL returns the URL to the Gravatar avatar image for email. If size
// is 0, the default is used.
func URL(email string, size uint16) string {
	if size == 0 {
		size = 128
	}
	email = strings.TrimSpace(email) // Trim leading and trailing whitespace from an email address.
	email = strings.ToLower(email)   // Force all characters to lower-case.
	h := md5.New()
	io.WriteString(h, email) // md5 hash the final string.
	return fmt.Sprintf("https://secure.gravatar.com/avatar/%x?s=%d&d=mm", h.Sum(nil), size)
}
