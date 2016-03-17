package mutexmap_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mutexmap"
)

func TestMutexMap(t *testing.T) {
	doneMu := sync.Mutex{}
	done := map[string]bool{}
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("key-%d", i)
		done[k] = false
	}

	m := mutexmap.New()
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("key-%d", i)
		m.Lock(k)
		go func() {
			time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
			doneMu.Lock()
			done[k] = true
			doneMu.Unlock()
			m.Unlock(k)
		}()
	}
	for i := 99; i >= 0; i-- {
		k := fmt.Sprintf("key-%d", i)
		m.Lock(k)
		doneMu.Lock()
		if !done[k] {
			t.Errorf("mutexmap lock on %s failed", k)
		}
		doneMu.Unlock()
		m.Unlock(k)
	}
}
