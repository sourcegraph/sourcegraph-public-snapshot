package codeowners

import "github.com/sourcegraph/sourcegraph/internal/types"

type ResolvedOwner interface {
	Type() OwnerType
	Identifier() string

	SetOwnerData(handle, email string)
}

// Guard to ensure all resolved owner types implement the interface
var (
	_ ResolvedOwner = (*Person)(nil)
	_ ResolvedOwner = (*Team)(nil)
)

type OwnerType string

const (
	OwnerTypePerson OwnerType = "person"
	OwnerTypeTeam   OwnerType = "team"
)

type Person struct {
	User         *types.User // If this is nil we've been unable to identify a user from the owner proto. Matches Own API.
	PrimaryEmail *string

	// Original proto fields.
	Handle string
	Email  string
}

func (p *Person) Type() OwnerType {
	return OwnerTypePerson
}

func (p *Person) Identifier() string {
	return p.Handle + p.Email
}

func (p *Person) GetEmail() string {
	if p.PrimaryEmail != nil {
		return *p.PrimaryEmail
	}
	return p.Email
}

func (p *Person) SetOwnerData(handle, email string) {
	p.Handle = handle
	p.Email = email
}

type Team struct {
	Team *types.Team

	// Original proto fields.
	Handle string
	Email  string
}

func (t *Team) Type() OwnerType {
	return OwnerTypeTeam
}

func (t *Team) Identifier() string {
	return t.Handle + t.Email
}

func (t *Team) SetOwnerData(handle, email string) {
	t.Handle = handle
	t.Email = email
}
