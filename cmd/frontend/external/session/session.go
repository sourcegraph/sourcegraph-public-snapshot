package session

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"

var (
	ResetMockSessionStore = session.ResetMockSessionStore
	SetActor              = session.SetActor
	SetData               = session.SetData
	GetData               = session.GetData
)
