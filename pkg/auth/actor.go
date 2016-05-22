package auth

import (
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// Actor represents an agent that accesses resources. It can represent
// an anonymous user or a logged-in user.
type Actor struct {
	// TODO: Make UID an int32.
	UID int `json:",omitempty"`

	// Login is the login of the currently authenticated user, if
	// any. It is provided as a convenience and is not guaranteed to
	// be correct (e.g., the user's login can change during the course
	// of a request if the user renames their account). It is also not
	// guaranteed to be populated (many request paths do not populate
	// it, as an optimization to avoid incurring the Users.Get call).
	Login string `json:",omitempty"`

	// Scope is a set of authorized scopes that the actor has
	// access to on the given server.
	Scope map[string]bool `json:",omitempty"`

	// Write indicates if the actor has global write access.
	Write bool

	// Admin indicates if the actor has global write access.
	Admin bool
}

func (a Actor) String() string {
	return fmt.Sprintf("Actor UID %d (clientID=%v scope=%v)", a.UID, a.Scope)
}

// IsAuthenticated returns true if the Actor is derived from an authenticated user.
func (a Actor) IsAuthenticated() bool {
	return a.UID != 0
}

// HasScope returns a boolean indicating whether this actor has the
// given scope.
func (a Actor) HasScope(s string) bool {
	hasScope, ok := a.Scope[s]
	return ok && hasScope
}

// HasWriteAccess checks if the actor has write or admin access.
func (a Actor) HasWriteAccess() bool {
	return a.IsAuthenticated() && (a.Write || a.Admin)
}

// HasAdminAccess checks if the actor has admin access.
func (a Actor) HasAdminAccess() bool {
	return a.IsAuthenticated() && (a.Admin)
}

func (a Actor) UserSpec() sourcegraph.UserSpec {
	return sourcegraph.UserSpec{
		UID:   int32(a.UID),
		Login: a.Login,
	}
}

func UnmarshalScope(scope []string) map[string]bool {
	scopeMap := make(map[string]bool)
	for _, s := range scope {
		scopeMap[s] = true
	}
	return scopeMap
}

func MarshalScope(scopeMap map[string]bool) []string {
	scope := make([]string, 0)
	for s, ok := range scopeMap {
		if !ok {
			continue
		}
		scope = append(scope, s)
	}
	return scope
}
