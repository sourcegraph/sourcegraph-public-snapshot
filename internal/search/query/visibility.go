package query

import "strings"

type repoVisibility string

const (
	Any     repoVisibility = "any"
	Private repoVisibility = "private"
	Public  repoVisibility = "public"
)

func ParseVisibility(s string) repoVisibility {
	switch strings.ToLower(s) {
	case "private":
		return Private
	case "public":
		return Public
	default:
		return Any
	}
}
