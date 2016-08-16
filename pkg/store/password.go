package store

import "context"

// Password defines the interface for checking, changing and resetting
// user passwords. It must be implemented by the Sourcegraph mothership.
type Password interface {
	CheckUIDPassword(ctx context.Context, UID int32, password string) error
	SetPassword(ctx context.Context, UID int32, password string) error
}
