// Package session exports symbols from frontend/internal/session. See the
// parent package godoc for more information.
package session

import "github.com/sourcegraph/sourcegraph/internal/session"

var (
	ResetMockSessionStore   = session.ResetMockSessionStore
	SetActor                = session.SetActor
	SetActorFromUser        = session.SetActorFromUser
	SetData                 = session.SetData
	GetData                 = session.GetData
	InvalidateSessionsByIDs = session.InvalidateSessionsByIDs
)
