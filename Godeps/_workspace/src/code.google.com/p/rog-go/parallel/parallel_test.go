package parallel_test

import (
	"code.google.com/p/rog-go/parallel"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestParallelMaxPar(t *testing.T) {
	const (
		totalDo = 10
		maxPar  = 3
	)
	var mu sync.Mutex
	max := 0
	n := 0
	tot := 0
	r := parallel.NewRun(maxPar)
	for i := 0; i < totalDo; i++ {
		r.Do(func() error {
			mu.Lock()
			tot++
			n++
			if n > max {
				max = n
			}
			mu.Unlock()
			time.Sleep(0.1e9)
			mu.Lock()
			n--
			mu.Unlock()
			return nil
		})
	}
	err := r.Wait()
	if n != 0 {
		t.Errorf("%d functions still running", n)
	}
	if tot != totalDo {
		t.Errorf("all functions not executed; want %d got %d", totalDo, tot)
	}
	if err != nil {
		t.Errorf("wrong error; want nil got %v", err)
	}
	if max != maxPar {
		t.Errorf("wrong number of do's ran at once; want %d got %d", maxPar, max)
	}
}

type intError int

func (intError) Error() string {
	return "error"
}

func TestParallelError(t *testing.T) {
	const (
		totalDo = 10
		errDo   = 5
	)
	r := parallel.NewRun(6)
	for i := 0; i < totalDo; i++ {
		i := i
		if i >= errDo {
			r.Do(func() error {
				return intError(i)
			})
		} else {
			r.Do(func() error {
				return nil
			})
		}
	}
	err := r.Wait()
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	errs := err.(parallel.Errors)
	if len(errs) != totalDo-errDo {
		t.Fatalf("wrong error count; want %d got %d", len(errs), totalDo-errDo)
	}
	ints := make([]int, len(errs))
	for i, err := range errs {
		ints[i] = int(err.(intError))
	}
	sort.Ints(ints)
	for i, n := range ints {
		if n != i+errDo {
			t.Errorf("unexpected error value; want %d got %d", i+errDo, n)
		}
	}
}
