package store

// DoneFunc is the function type of store's Done method.
type DoneFunc func(err error) error

// noopDoneFunc is a behaviorless DoneFunc.
func noopDoneFunc(err error) error {
	return err
}
