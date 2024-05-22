// Package session exports symbols from frontend/internal/session. See the
// parent package godoc for more information.
package session

import "github.com/sourcegraph/sourcegraph/internal/session"

var (
	ResetMockSessionStore = session.ResetMockSessionStore
	// SetActor sets the actor in the session, or removes it if actor == nil. If no session exists, a
	// new session is created.
	//
	// 🚨 SECURITY: Should only be called after user is successfully authenticated.
	SetActor = session.SetActor
	// SetActorFromUser creates an actor from a user, sets it in the session, and
	// returns a context with the user attached.
	//
	// 🚨 SECURITY: Should only be called after user is successfully authenticated.
	SetActorFromUser        = session.SetActorFromUser
	SetData                 = session.SetData
	GetData                 = session.GetData
	InvalidateSessionsByIDs = session.InvalidateSessionsByIDs
)
