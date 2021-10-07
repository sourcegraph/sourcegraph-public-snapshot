package query

import "strings"

type RepoVisibility string

const (
	Any     RepoVisibility = "Any"
	Private RepoVisibility = "Private"
	Public  RepoVisibility = "Public"
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
