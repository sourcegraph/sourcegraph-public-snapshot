package codeowners

import "github.com/sourcegraph/sourcegraph/internal/types"

type OwnerType string

const (
	OwnerTypePerson OwnerType = "person"
	OwnerTypeTeam             = "team"
)

type ResolvedOwner interface {
	Type() OwnerType
}

type Person struct {
	User   *types.User // todo extend to user for SCIM
	Handle string
	Email  string
}

func (p Person) Type() OwnerType {
	return OwnerTypePerson
}
