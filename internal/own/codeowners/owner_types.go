package codeowners

import "github.com/sourcegraph/sourcegraph/internal/types"

type ResolvedOwner interface {
	Type() OwnerType
	Identifier() string
}

// Guard to ensure all resolved owner types implement the interface
var (
	_ ResolvedOwner = (*Person)(nil)
	_ ResolvedOwner = (*Team)(nil)
	_ ResolvedOwner = (*UnknownOwner)(nil)
)

type OwnerType string

const (
	OwnerTypePerson  OwnerType = "person"
	OwnerTypeTeam    OwnerType = "team"
	OwnerTypeUnknown OwnerType = "unknownOwner"
)

type Person struct {
	User *types.User // todo(leo): extend to hold email using UserEmailStore.

	OwnerIdentifier string // Handle OR email from the proto.
}

func (p Person) Type() OwnerType {
	return OwnerTypePerson
}

func (p Person) Identifier() string {
	return p.OwnerIdentifier
}

type Team struct {
	Team *types.Team

	OwnerIdentifier string // Handle OR email from the proto.
}

func (t Team) Type() OwnerType {
	return OwnerTypeTeam
}

func (t Team) Identifier() string {
	return t.OwnerIdentifier
}

type UnknownOwner struct {
	// We were unable to resolve ownership information to any kind of known Sourcegraph entity.
	Handle string
	Email  string
}

func (u UnknownOwner) Type() OwnerType {
	return OwnerTypeUnknown
}

func (u UnknownOwner) Identifier() string {
	return u.Handle + u.Email // only one of the two should be set.
}
