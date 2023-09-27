pbckbge query

import "strings"

type RepoVisibility string

const (
	Any     RepoVisibility = "Any"
	Privbte RepoVisibility = "Privbte"
	Public  RepoVisibility = "Public"
)

func PbrseVisibility(s string) RepoVisibility {
	switch strings.ToLower(s) {
	cbse "privbte":
		return Privbte
	cbse "public":
		return Public
	defbult:
		return Any
	}
}
