package goroutine

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

func TestRunWorkersN(t *testing.T) {
	n := 25

	ch := make(chan int, n)
	var expectedErrors []string
	for i := 0; i < 25; i++ {
		if i%3 == 0 {
			expectedErrors = append(expectedErrors, fmt.Sprintf("err: %d", i))
		}

		ch <- i
	}
	close(ch)
	sort.Strings(expectedErrors)

	err := RunWorkersN(n, SimplePoolWorker(func() error {
		if v := <-ch; v%3 == 0 {
			return fmt.Errorf("err: %d", v)
		}
		return nil
	}))

	merr, ok := err.(*multierror.Error)
	if err == nil || !ok {
		t.Errorf("unexpected error wrapper: %v", err)
	}

	var errStrings []string
	for _, err := range merr.WrappedErrors() {
		errStrings = append(errStrings, err.Error())
	}
	sort.Strings(errStrings)

	if diff := cmp.Diff(expectedErrors, errStrings); diff != "" {
		t.Errorf("unexpected errors (-want +got):\n%s", diff)
	}
}
