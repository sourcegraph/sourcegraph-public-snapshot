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
	// queue all work
	queue := make(chan int, len(fns))
	for i := range fns {
		queue <- i
	}
	close(queue)

	// create map with all work marked as pending
	pending := map[int]bool{}
	for i := 0; i < len(fns); i++ {
		pending[i] = false
	}
	pendingMap := &pendingMap{pending: pending}

	// launch workers
	errs := make(chan errPair, len(fns))
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			for i := range queue {
				pendingMap.set(i)
				err := fns[i].Fn()
				errs <- errPair{i, err}
			}
		}()
	}

	return pentimento.PrintProgress(func(p *pentimento.Printer) error {
		for {
			n := pendingMap.size()
			if n == 0 {
				break
			}

			select {
			case pair := <-errs:
				if pair.err != nil {
					go func() {
						// Drain queue to stop additional work
						for range queue {
						}
					}()

					go func() {
						// Drain errors so we can close the channel safely
						// without closing it while one of the worker goroutines
						// above are still running.
						for i := 0; i < n; i++ {
							<-errs
						}
						close(errs)
					}()

					// Clear the screen
					_ = p.Reset()
					return pair.err
				}

				// Nil-valued error, remove it from the pending map
				pendingMap.remove(pair.i)

			case <-time.After(time.Millisecond * 250):
				// Time's up, fall through and update the screen
			}

			content := pentimento.NewContent()

			for count, i := range pendingMap.keys() {
				if count > MaxDisplayLines {
					content.AddLine("\n(additional pending tasks omitted)...")
					break
				}

				if pendingMap.get(i) {
					content.AddLine(fmt.Sprintf("%s %s", pentimento.Dots, fns[i].Description))
				} else {
					content.AddLine(fmt.Sprintf("%s %s", "   ", fns[i].Description))
				}
			}

			_ = p.WriteContent(content)
		}

		_ = p.Reset()
		return nil
	})
}

type pendingMap struct {
	sync.RWMutex
	pending map[int]bool
}

func (m *pendingMap) remove(i int) {
	m.Lock()
	defer m.Unlock()
	delete(m.pending, i)
}

func (m *pendingMap) keys() (keys []int) {
	m.RLock()
	defer m.RUnlock()

	for k := range m.pending {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (m *pendingMap) set(i int) {
	m.Lock()
	defer m.Unlock()
	m.pending[i] = true
}

func (m *pendingMap) get(i int) bool {
	m.RLock()
	defer m.RUnlock()
	return m.pending[i]
}

func (m *pendingMap) size() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.pending)
}
