// Package textutil has utility functions for working with text.
package textutil

import (
	"fmt"
	"html"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"

	"github.com/microcosm-cc/bluemonday"
)

// FirstNameOrLogin returns returns the user's first name if the user has a name. Otherwise, it returns the user's login.
func FirstNameOrLogin(user *sourcegraph.User) string {
	if firstName := FirstName(user); firstName != "" {
		return firstName
	}
	return user.Login
}

// FirstName returns the user's first name or the empty string if the user has no name.
func FirstName(user *sourcegraph.User) string {
	fields := strings.Fields(user.Name)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// FirstSentence returns the first sentence from s, including the trailing
// period (if present).
func FirstSentence(s string) string {
	trunc := false

	i0 := strings.Index(s, ". ")
	i1 := strings.Index(s, ".\n")
	i := -1
	if i0 == -1 {
		i = i1
	} else if i1 == -1 {
		i = i0
	} else if i0 < i1 {
		i = i0
	} else {
		i = i1
	}

	maxlen := 300
	if i == -1 {
		if len(s) > maxlen {
			rs := []rune(s[:maxlen-1])
			i = len(string(rs))
			trunc = true
		} else {
			i = len(s) - 1
		}
	}
	if i > maxlen {
		i = maxlen - 1
	}
	fs := s[:i+1]
	if trunc {
		fs += "..."
	}
	return fs
}

func FirstChars(s string) string {
	rs := []rune(s)
	maxlen := 500
	if len(rs) < maxlen {
		return string(rs)
	}
	return fmt.Sprintf("%s...", string(rs[0:maxlen]))
}

// TextFromHTML returns all of the textual content in htmlStr. HTML entities are
// unescaped in the returned string.
func TextFromHTML(htmlStr string) string {
	return html.UnescapeString(bluemonday.StrictPolicy().Sanitize(htmlStr))
}

func Truncate(n int, s string) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n]) + "…"
}

// ShortCommitMessage truncates the given git commit message and suffixes it
// with "…" (single rune) such that the returned string is at max N runes.
//
// Only the first line of the string is considered.
func ShortCommitMessage(n int, msg string) string {
	if strings.Contains(msg, "\n") {
		msg = strings.Split(msg, "\n")[0]
	}
	return Truncate(n-1, msg)
}
