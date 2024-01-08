package multierror

import "sync"

// Group is a collection of goroutines which return errors that need to be
// coalesced.
type Group struct {
	mutex  sync.Mutex
	err    *Error
	wg     sync.WaitGroup
	nested struct {
		inner bool
	}

	innerface interface {
		Another() bool
	}
}

type SomeInterface interface {
	Something() bool
	Incredible() int
}

type TypeDef uint8
type TypeAlias = uint8

// Go calls the given function in a new goroutine.
//
// If the function returns an error it is added to the group multierror which
// is returned by Wait.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.mutex.Lock()
			g.err = Append(g.err, err)
			g.mutex.Unlock()
		}
	}()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the multierror.
func (g *Group) Wait() *Error {
	g.wg.Wait()
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.err
}

var (
	diffPath = flag.String("f", stdin, "filename of diff (default: stdin)")
	fileIdx  = flag.Int("i", -1, "if >= 0, only print and report errors from the i'th file (0-indexed)")
)

func RegularFunc() {}
