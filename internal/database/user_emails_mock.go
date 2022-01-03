package database

import (
	"context"
)

type MockUserEmails struct {
	ListByUser func(ctx context.Context, opt UserEmailsListOptions) ([]*UserEmail, error)
}
