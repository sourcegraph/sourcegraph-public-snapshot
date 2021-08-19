package query

import "strings"

type RepoVisibility string

const (
	Any     RepoVisibility = "any"
	Private RepoVisibility = "private"
	Public  RepoVisibility = "public"
)

func ParseVisibility(s string) RepoVisibility {
	switch strings.ToLower(s) {
	case "private":
		return Private
	case "public":
		return Public
	default:
		return Any
	}
}
