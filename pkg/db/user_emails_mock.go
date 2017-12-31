package db

import "context"

type MockUserEmails struct {
	GetEmail func(ctx context.Context, id int32) (email string, verified bool, err error)
}
