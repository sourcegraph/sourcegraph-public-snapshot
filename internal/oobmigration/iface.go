package oobmigration

import "context"

// storeIface is an interface of the Store methods used by Runner.
type storeIface interface {
	Transact(ctx context.Context) (storeIface, error)
	Done(err error) error

	SynchronizeMetadata(ctx context.Context) error
	List(ctx context.Context) ([]Migration, error)
	UpdateDirection(ctx context.Context, id int, applyReverse bool) error
	UpdateProgress(ctx context.Context, id int, progress float64) error
	AddError(ctx context.Context, id int, message string) error
}

type storeShim struct {
	*Store
}

func (s *storeShim) Transact(ctx context.Context) (storeIface, error) {
	tx, err := s.Store.Transact(ctx)
	return &storeShim{tx}, err
}
