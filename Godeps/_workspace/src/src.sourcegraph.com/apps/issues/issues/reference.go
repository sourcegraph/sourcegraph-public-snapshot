package issues

import (
	"html/template"
)

// Reference represents a reference to code.
type Reference struct {
	Repo      RepoSpec
	Path      string // Path is a relative, '/'-separated path to a file within a repo.
	CommitID  string
	StartLine uint32
	EndLine   uint32
	Contents  template.HTML
}
