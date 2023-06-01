package uploadstore

import (
	"runtime"

	"github.com/sourcegraph/conc/pool"
)

// RunWorkersOverStrings invokes the given worker once for each of the
// given string values. The worker function will receive the index as well
// as the string value as parameters. Workers will be invoked in a number
// of concurrent routines proportional to the maximum number of CPUs that
// can be executing simultaneously.
func RunWorkersOverStrings(values []string, worker func(index int, value string) error) error {
	p := pool.New().
		WithErrors().
		WithMaxGoroutines(runtime.GOMAXPROCS(0))
	for i, value := range values {
		i, value := i, value
		p.Go(func() error {
			return worker(i, value)
		})
	}
	return p.Wait()
}
