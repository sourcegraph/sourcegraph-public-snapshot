package uploadstore

import (
	"runtime"

	"github.com/sourcegraph/conc/pool"
)

// ForEachString invokes the given callback once for each of the
// given string values. The callback function will receive the index as well
// as the string value as parameters. Callbacks will be invoked in a number
// of concurrent routines proportional to the maximum number of CPUs that
// can be executing simultaneously.
func ForEachString(values []string, f func(index int, value string) error) error {
	p := pool.New().
		WithErrors().
		WithMaxGoroutines(runtime.GOMAXPROCS(0))
	for i, value := range values {
		i, value := i, value
		p.Go(func() error {
			return f(i, value)
		})
	}
	return p.Wait()
}
