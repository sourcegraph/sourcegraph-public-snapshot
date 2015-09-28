package auth

import (
	"fmt"
	"strconv"
)

// Resource represents a thing that requires permission to be accessed. E.g., a repository, person, or organization.
// Currently, only support for person and organization is implemented.
type Resource struct {
	ID   int
	URI  string // for repositories only
	Type ResourceType
}

func (r Resource) String() string {
	var ident string
	if r.ID != 0 {
		ident = fmt.Sprintf("(ID %d)", r.ID)
	} else if r.URI != "" {
		ident = fmt.Sprintf("(URI %q)", r.URI)
	}
	return fmt.Sprintf("%s resource %s", r.Type, ident)
}

type ResourceType int

const (
	ResourcePerson ResourceType = iota
	ResourceTeam
	ResourceRepo
	ResourceOrg
	ResourceSGInternal // special resource that should only be accessible to Sourcegraph employees
)

func (rt ResourceType) String() string {
	switch rt {
	case ResourcePerson:
		return "person"
	case ResourceTeam:
		return "team"
	case ResourceRepo:
		return "repo"
	case ResourceOrg:
		return "org"
	case ResourceSGInternal:
		return "sg-internal"
	default:
		panic("unknown resource type: " + strconv.Itoa(int(rt)))
	}
}
