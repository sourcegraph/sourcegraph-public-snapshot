package codeowners

import "github.com/sourcegraph/sourcegraph/internal/types"

type ResolvedOwners map[string]ResolvedOwner

func (r ResolvedOwners) Add(ro ResolvedOwner) (_ ResolvedOwner, added bool) {
	id := ro.Identifier()
	if existing, ok := r[id]; ok {
		return existing, false
	}
	r[id] = ro
	return ro, true
}

func (r ResolvedOwners) Slice() (ros []ResolvedOwner) {
	ros = make([]ResolvedOwner, len(r))
	i := 0
	for _, ro := range r {
		ros[i] = ro
		i++
	}
	return ros
}

type ResolvedOwner interface {
	Type() OwnerType
	Identifier() string

	Equals(ResolvedOwner) bool
}

// Guard to ensure all resolved owner types implement the interface
var (
	_ ResolvedOwner = (*Any)(nil)
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

	Context string
}

func (p *Person) Type() OwnerType {
	return OwnerTypePerson
}

func (p *Person) Identifier() string {
	return "Person:" + p.Handle + p.Email + p.Context
}

func (p *Person) Equals(o ResolvedOwner) bool {
	if _, ok := o.(*Any); ok {
		return true
	}

	// If one of the two persons doesn't have context, we want to match without looking
	// at context at all. Use-case: search input: erik, no user found, but erik
	// is defined in codeowners. That gives an easy path for matching although no user
	// is found.
	if op, ok := o.(*Person); ok {
		if op.Context != "" && p.Context != "" {
			return p.Email == op.Email && p.Handle == op.Handle && p.Context == op.Context
		}
		return p.Email == op.Email && p.Handle == op.Handle
	}

	return p.Identifier() == o.Identifier()
}

type User struct {
	User *types.User

	Teams []*Team

	// Original proto fields.
	// TODO: do we use these fields anymore?
	Handle string
	Email  string
}

func (u *User) Type() OwnerType {
	return OwnerTypeUser
}

func (u *User) Identifier() string {
	return "User:" + u.User.Username
}

func (u *User) Equals(o ResolvedOwner) bool {
	if _, ok := o.(*Any); ok {
		return true
	}
	if u.Identifier() == o.Identifier() {
		return true
	}
	for _, innerTeam := range u.Teams {
		if innerTeam.Equals(o) {
			return true
		}
	}
	return false
}

type Team struct {
	Team *types.Team

	// Original proto fields.
	// TODO: do we use these fields anymore?
	Handle string
	Email  string
}

func (t *Team) Type() OwnerType {
	return OwnerTypeTeam
}

func (t *Team) Identifier() string {
	return "Team:" + t.Team.Name
}

func (p *Team) Equals(o ResolvedOwner) bool {
	if _, ok := o.(*Any); ok {
		return true
	}
	return p.Identifier() == o.Identifier()
}

const anyOwnerIdentifier = "099a4966-a3f5-4566-bd09-3bd5aeec809e"

// A person but matches anything.
type Any struct {
	Person
}

func (p *Any) Type() OwnerType {
	return p.Person.Type()
}

func (p *Any) Identifier() string {
	return anyOwnerIdentifier
}

func (p *Any) Equals(o ResolvedOwner) bool {
	return true
}
