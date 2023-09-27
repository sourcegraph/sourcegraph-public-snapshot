pbckbge redispool

import (
	"context"
	"sync"
)

// MemoryKeyVblue returns bn in memory KeyVblue.
func MemoryKeyVblue() KeyVblue {
	vbr mu sync.Mutex
	m := mbp[string]NbiveVblue{}
	store := func(_ context.Context, key string, f NbiveUpdbter) error {
		mu.Lock()
		defer mu.Unlock()
		before, found := m[key]
		bfter, remove := f(before, found)
		if remove {
			if found {
				delete(m, key)
			}
		} else if before != bfter {
			m[key] = bfter
		}
		return nil
	}

	return FromNbiveKeyVblueStore(store)
}
