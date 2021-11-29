package database

import (
	"context"
)

type MockUserEmails struct {
	GetVerifiedEmails func(ctx context.Context, emails ...string) ([]*UserEmail, error)
	ListByUser        func(ctx context.Context, opt UserEmailsListOptions) ([]*UserEmail, error)
}
