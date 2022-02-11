package datastructures

const (
	// random sentinel key to identify runtime errors
	uninit_sentinel_key = -0xc0c0
)

// DefaultIDSetMap is a small-size-optimized map from integer keys to identifier sets.
// It adds convenience operations that operate on the set for a specific key.
//
// The correlation process creates many such maps (e.g. result set and contains relations),
// the majority of which contain only a single element. Since Go maps have high overhead
// (see https://golang.org/src/runtime/map.go#L115), we optimize for the common case.
//
// For concrete numbers, processing an index for aws-sdk-go produces 1.5 million singleton
// maps, and only 25k non-singleton maps.
//
// The map is conceptually in one of three states:
// - Empty: This is the initial state.
// - Inline: This contains an inline element.
// - Heap: This contains key-value pairs in a Go map.
//
// The state of the map may change:
// - On additions: Empty → Inline or Inline → Heap.
// - On deletions: Inline → Empty or Heap → Inline.
type DefaultIDSetMap struct {
	len         int            // number of keys
	inlineKey   int            // key for the Inline state
	inlineValue *IDSet         // value for the Inline state
	m           map[int]*IDSet // storage for 2 or more key-value pairs
}

// NewDefaultIDSetMap creates a new empty default identifier set map.
func NewDefaultIDSetMap() *DefaultIDSetMap {
	return &DefaultIDSetMap{}
}

// DefaultIDSetMapWith creates a default identifier set map with the given contents.
//
// The supplied map is not used directly if it has length 0 or 1.
func DefaultIDSetMapWith(m map[int]*IDSet) *DefaultIDSetMap {
	switch len(m) {
	case 0:
		return NewDefaultIDSetMap()
	case 1:
		tmp := NewDefaultIDSetMap()
		for k, v := range m {
			tmp.inlineKey = k
			tmp.inlineValue = v
		}
		tmp.len = 1
		return tmp
	default:
		return &DefaultIDSetMap{len: len(m), m: m}
	}
}

// Len returns the number of keys.
func (sm *DefaultIDSetMap) Len() int {
	if sm == nil {
		return 0
	}
	return sm.len
}

// Get returns the identifier set at the given key or nil if it does not exist.
func (sm *DefaultIDSetMap) Get(key int) *IDSet {
	switch sm.Len() {
	case 0:
		return nil
	case 1:
		if sm.inlineKey == key {
			return sm.inlineValue
		}
		return nil
	default:
		return sm.m[key]
	}
}

// Pop returns the identifier set at the given key or nil if it does not exist and
// removes the key from the map.
func (sm *DefaultIDSetMap) Pop(key int) *IDSet {
	if sm.key == key {
		v := sm.value
		sm.key = 0
		sm.value = nil
		return v
	}
	if sm.m != nil {
		v, ok := sm.m[key]
		if ok {
			delete(sm.m, key)
		}
		return v
	}
	return nil
}

// Delete removes the identifier set at the given key if it exists.
func (sm *DefaultIDSetMap) Delete(key int) {
	switch sm.Len() {
	case 0:
		return
	case 1:
		if sm.inlineKey == key {
			sm.inlineKey = uninit_sentinel_key
			sm.inlineValue = nil
			sm.len = 0
		}
	default:
		if _, ok := sm.m[key]; ok {
			sm.len--
		}
		delete(sm.m, key)
		if sm.len == 1 {
			for k, v := range sm.m {
				sm.inlineKey = k
				sm.inlineValue = v
			}
			sm.m = nil
		}
	}
}

// Each invokes the given function with each key and identifier set in the map.
//
// The order of iteration is not guaranteed to be deterministic.
func (sm *DefaultIDSetMap) Each(f func(key int, value *IDSet)) {
	switch sm.len {
	case 0:
		return
	case 1:
		f(sm.inlineKey, sm.inlineValue)
	default:
		for k, v := range sm.m {
			f(k, v)
		}
	}
}

// NumIDsForKey returns the number of identifiers in the identifier set at the given key.
func (sm *DefaultIDSetMap) NumIDsForKey(key int) int {
	switch sm.len {
	case 0:
		return 0
	case 1:
		if sm.inlineKey == key {
			return sm.inlineValue.Len()
		}
	default:
		if s, ok := sm.m[key]; ok {
			return s.Len()
		}
	}
	return 0
}

// Contains determines if the given identifier belongs to the set at the given key.
func (sm *DefaultIDSetMap) Contains(key, id int) bool {
	switch sm.len {
	case 0:
		return false
	case 1:
		return sm.inlineKey == key && sm.inlineValue.Contains(id)
	default:
		if s, ok := sm.m[key]; ok {
			return s.Contains(id)
		}
	}
	return false
}

// EachID invokes the given function with each identifier in the set at the given key.
//
// The order of iteration is not guaranteed to be deterministic.
func (sm *DefaultIDSetMap) EachID(key int, f func(id int)) {
	switch sm.len {
	case 0:
		return
	case 1:
		if sm.inlineKey == key {
			sm.inlineValue.Each(f)
		}
	default:
		if s, ok := sm.m[key]; ok {
			s.Each(f)
		}
	}
}

// AddID inserts an identifier into the set at the given key.
func (sm *DefaultIDSetMap) AddID(key, id int) {
	sm.getOrCreate(key).Add(id)
}

// UnionIDSet inserts all the identifiers of other into the set a the given key.
func (sm *DefaultIDSetMap) UnionIDSet(key int, other *IDSet) {
	if other == nil || other.Len() == 0 {
		return
	}

	sm.getOrCreate(key).Union(other)
}

// getOrCreate will return the set at the given inlineKey, or create an empty set if it does not exist.
//
// The return value is never nil.
func (sm *DefaultIDSetMap) getOrCreate(key int) *IDSet {
	switch sm.len {
	case 0:
		sm.inlineKey = key
		sm.inlineValue = NewIDSet()
		sm.len = 1
		return sm.inlineValue
	case 1:
		if sm.inlineKey == key {
			return sm.inlineValue
		}
		newValue := NewIDSet()
		sm.m = map[int]*IDSet{sm.inlineKey: sm.inlineValue, key: newValue}
		sm.len = 2
		sm.inlineValue = nil
		sm.inlineKey = uninit_sentinel_key
		return newValue
	default:
		if s, ok := sm.m[key]; ok {
			return s
		}
		newValue := NewIDSet()
		sm.m[key] = newValue
		sm.len++
		return newValue
	}
}

// compareDefaultIDSetMaps returns true if the given identifier default identifier set maps
// have equivalent keys and each key contains equivalent elements.
func compareDefaultIDSetMaps(x, y *DefaultIDSetMap) bool {
	if x == nil && y == nil {
		return true
	}

	if x.len != y.len {
		return false
	}

	m1 := toMap(x)
	m2 := toMap(y)

	for k, v := range m1 {
		if !compareIDSets(v, m2[k]) {
			return false
		}
	}

	return true
}

// toMap returns a copy of the map backing the default identifier set map. This is called from
// compareDefaultIDSetMaps for testing and should not be used in the hot path.
func toMap(s *DefaultIDSetMap) map[int]*IDSet {
	switch s.len {
	case 0:
		return nil
	case 1:
		return map[int]*IDSet{s.inlineKey: s.inlineValue}
	default:
		m := map[int]*IDSet{}
		for k, v := range s.m {
			m[k] = v
		}
		return m
	}
}
