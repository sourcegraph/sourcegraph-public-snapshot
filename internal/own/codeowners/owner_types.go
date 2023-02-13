package codeowners

import "github.com/sourcegraph/sourcegraph/internal/types"

type ResolvedOwner interface {
	Type() OwnerType
}

// Guard to ensure all resolved owner types implement the interface
var (
	_ ResolvedOwner = (*Person)(nil)
	_ ResolvedOwner = (*Team)(nil)
	_ ResolvedOwner = (*Unknown)(nil)
)

type OwnerType string

const (
	OwnerTypePerson  OwnerType = "person"
	OwnerTypeTeam              = "team"
	OwnerTypeUnknown           = "unknown"
)

type Person struct {
	User *types.User // todo(leo): extend to types.UserForSCIM. right now types.User does not hold email information.

	// Common proto fields
	Handle string
	Email  string
}

func (p Person) Type() OwnerType {
	return OwnerTypePerson
}

type Team struct {
	Team *types.Team

	// Common proto fields
	Handle string
	Email  string
}

func (t Team) Type() OwnerType {
	return OwnerTypeTeam
}

type Unknown struct {
	// We were unable to resolve ownership information to any kind of known Sourcegraph entity.

	// Common proto fields
	Handle string
	Email  string
}

func (u Unknown) Type() OwnerType {
	return OwnerTypeUnknown
}
