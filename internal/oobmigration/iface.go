package oobmigration

import "context"

// storeIface is an interface of the Store methods used by Runner.
type storeIface interface {
	List(ctx context.Context) ([]Migration, error)
	UpdateProgress(ctx context.Context, id int, progress float64) error
	AddError(ctx context.Context, id int, message string) error
}
