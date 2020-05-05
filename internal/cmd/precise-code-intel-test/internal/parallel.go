package internal

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/efritz/pentimento"
)

// MaxDisplayLines is the number of lines that will be displayed before truncation.
const MaxDisplayLines = 50

// FnPair groups an error-returning function with a description that can be displayed
// by RunParallel.
type FnPair struct {
	Fn          func() error
	Description string
}

type errPair struct {
	i   int
	err error
}

// RunParallel runs each function in parallel. Returns the first error to occur. The
// number of invocations is limited by maxConcurrency.
func RunParallel(maxConcurrency int, fns []FnPair) error {
	return pentimento.PrintProgress(func(p *pentimento.Printer) error {
		queue := make(chan int, len(fns))
		for i := range fns {
			queue <- i
		}

		var m sync.RWMutex
		pending := map[int]bool{}
		for i := 0; i < len(fns); i++ {
			pending[i] = false
		}

		errs := make(chan errPair, len(fns))
		for i := 0; i < maxConcurrency; i++ {
			go func() {
				for i := range queue {
					m.Lock()
					pending[i] = true
					m.Unlock()

					err := fns[i].Fn()
					errs <- errPair{i, err}
				}
			}()
		}

		for {
			m.RLock()
			n := len(pending)
			m.RUnlock()
			if n == 0 {
				break
			}

			content := pentimento.NewContent()

			select {
			case pair := <-errs:
				if pair.err != nil {
					go func() {
						for i := 0; i < n; i++ {
							<-errs
						}
						close(errs)
					}()

					_ = p.Reset()
					return pair.err
				}

				m.Lock()
				temp := map[int]bool{}
				for i, processing := range pending {
					if i != pair.i {
						temp[i] = processing
					}
				}
				pending = temp
				m.Unlock()

			case <-time.After(time.Millisecond * 250):
			}

			m.RLock()
			var keys []int
			for k := range pending {
				keys = append(keys, k)
			}
			sort.Ints(keys)

			for count, i := range keys {
				if count > MaxDisplayLines {
					break
				}

				if pending[i] {
					content.AddLine(fmt.Sprintf("%s %s", pentimento.Dots, fns[i].Description))
				} else {
					content.AddLine(fmt.Sprintf("%s %s", "   ", fns[i].Description))
				}
			}
			m.RUnlock()

			if len(keys) > MaxDisplayLines {
				content.AddLine("\n(additional pending tasks omitted)...")
			}

			_ = p.WriteContent(content)
		}

		_ = p.Reset()
		return nil
	})
}
