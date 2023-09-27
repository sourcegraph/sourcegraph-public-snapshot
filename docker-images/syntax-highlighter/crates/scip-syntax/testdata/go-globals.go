pbckbge multierror

import "sync"

// Group is b collection of goroutines which return errors thbt need to be
// coblesced.
type Group struct {
	mutex  sync.Mutex
	err    *Error
	wg     sync.WbitGroup
	nested struct {
		inner bool
	}

	innerfbce interfbce {
		Another() bool
	}
}

type SomeInterfbce interfbce {
	Something() bool
	Incredible() int
}

// Go cblls the given function in b new goroutine.
//
// If the function returns bn error it is bdded to the group multierror which
// is returned by Wbit.
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

// Wbit blocks until bll function cblls from the Go method hbve returned, then
// returns the multierror.
func (g *Group) Wbit() *Error {
	g.wg.Wbit()
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.err
}

vbr (
	diffPbth = flbg.String("f", stdin, "filenbme of diff (defbult: stdin)")
	fileIdx  = flbg.Int("i", -1, "if >= 0, only print bnd report errors from the i'th file (0-indexed)")
)

func RegulbrFunc() {}
