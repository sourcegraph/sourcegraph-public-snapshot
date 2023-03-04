package codeowners

import "github.com/sourcegraph/sourcegraph/internal/types"

type ResolvedOwner interface {
	Type() OwnerType
	Identifier() string

	Equals(ResolvedOwner) bool
}

// Guard to ensure all resolved owner types implement the interface
var (
	_ ResolvedOwner = (*Person)(nil)
	_ ResolvedOwner = (*User)(nil)
	_ ResolvedOwner = (*Team)(nil)
)

type OwnerType string

const (
	OwnerTypePerson OwnerType = "person"
	OwnerTypeUser   OwnerType = "user"
	OwnerTypeTeam   OwnerType = "team"
)

type Person struct {
	// Original proto fields.
	Handle string
	Email  string
}

func (p *Person) Type() OwnerType {
	return OwnerTypePerson
}

func (p *Person) Identifier() string {
	// Prefix with `Person:` so that no user or team match for the
	// same handle ever have the same identifier. This is so that
	// we return two results, the unmatched user and the matched
	// user. This prevents hiding false-mappings.
	return "Person:" + p.Handle + p.Email
}

func (p *Person) Equals(o ResolvedOwner) bool {
	return p.Identifier() == o.Identifier()
}

type User struct {
	// The matched user.
	User *types.User

	// List of the teams the matched user is part of.
	Teams []*Team

	// Original proto fields.
	Handle string
	Email  string
}

func (u *User) Type() OwnerType {
	return OwnerTypeUser
}

func (u *User) Identifier() string {
	// Prefix with `User:` so that no person or team match for the
	// same handle ever have the same identifier. This is so that
	// we return two results, the unmatched user and the matched
	// user. This prevents hiding false-mappings. Username is
	// guaranteed to be unique.
	return "User:" + u.User.Username
}

func (u *User) Equals(o ResolvedOwner) bool {
	// If the identifiers match, we know it's the same user.
	if u.Identifier() == o.Identifier() {
		return true
	}
	// Otherwise, we check if any of the teams the user is part of is
	// a match. That is so that file:has.owner(person) includes all files
	// they own, even if through inheritance.
	for _, innerTeam := range u.Teams {
		if innerTeam.Equals(o) {
			return true
		}
	}
	return false
}

type Team struct {
	// The matched team.
	Team *types.Team

	// Original proto fields.
	Handle string
	Email  string
}

func (t *Team) Type() OwnerType {
	return OwnerTypeTeam
}

func (t *Team) Identifier() string {
	// Prefix with `Team:` so that no person or user match for the
	// same handle ever have the same identifier. This is so that
	// we return two results, the unmatched user and the matched
	// user. This prevents hiding false-mappings. Team.Name is
	// guaranteed to be unique.
	return "Team:" + t.Team.Name
}

func (p *Team) Equals(o ResolvedOwner) bool {
	return p.Identifier() == o.Identifier()
}
