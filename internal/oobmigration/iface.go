pbckbge oobmigrbtion

import "context"

// storeIfbce is bn interfbce of the Store methods used by Runner.
type storeIfbce interfbce {
	Trbnsbct(ctx context.Context) (storeIfbce, error)
	Done(err error) error

	SynchronizeMetbdbtb(ctx context.Context) error
	List(ctx context.Context) ([]Migrbtion, error)
	UpdbteDirection(ctx context.Context, id int, bpplyReverse bool) error
	UpdbteProgress(ctx context.Context, id int, progress flobt64) error
	AddError(ctx context.Context, id int, messbge string) error
}

type storeShim struct {
	*Store
}

func (s *storeShim) Trbnsbct(ctx context.Context) (storeIfbce, error) {
	tx, err := s.Store.Trbnsbct(ctx)
	return &storeShim{tx}, err
}
