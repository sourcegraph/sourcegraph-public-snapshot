package datastructures

type mapState int

const (
	mapStateEmpty mapState = iota
	mapStateInline
	mapStateHeap
	ILLEGAL_MAPSTATE = "invariant violation: illegal map state!"
	// random sentinel key to identify runtime errors
	uninitSentinelKey = -0xc0c0
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
	inlineKey   int            // key for the Inline state
	inlineValue *IDSet         // value for the Inline state
	m           map[int]*IDSet // storage for 2 or more key-value pairs
}

// NewDefaultIDSetMap creates a new empty default identifier set map.
func NewDefaultIDSetMap() *DefaultIDSetMap {
	return &DefaultIDSetMap{}
}

func (sm *DefaultIDSetMap) state() mapState {
	if sm.inlineValue == nil {
		if sm.m == nil {
			return mapStateEmpty
		}
		return mapStateHeap
	}
	if sm.m != nil {
		panic("m field of DefaultIDSetMap should be nil when value is present inline")
	}
	return mapStateInline
}

// DefaultIDSetMapWith creates a default identifier set map with
// a copy of the given contents.
//
// map entries with nil or empty IDSets are ignored.
func DefaultIDSetMapWith(m map[int]*IDSet) *DefaultIDSetMap {
	tmp := NewDefaultIDSetMap()
	for k, v := range m {
		tmp.UnionIDSet(k, v)
	}
	return tmp
}

// Len returns the number of keys.
func (sm *DefaultIDSetMap) Len() int {
	switch sm.state() {
	case mapStateEmpty:
		return 0
	case mapStateInline:
		return 1
	case mapStateHeap:
		return len(sm.m)
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}

// UnorderedKeys returns a slice with a copy of all keys in an unspecified order.
func (sm *DefaultIDSetMap) UnorderedKeys() []int {
	switch sm.state() {
	case mapStateEmpty:
		return []int{}
	case mapStateInline:
		return []int{sm.inlineKey}
	case mapStateHeap:
		var out = make([]int, 0, sm.Len())
		for k := range sm.m {
			out = append(out, k)
		}
		return out
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}

// Get returns the identifier set at the given key or nil if it does not exist.
func (sm *DefaultIDSetMap) Get(key int) *IDSet {
	switch sm.state() {
	case mapStateEmpty:
		return nil
	case mapStateInline:
		if sm.inlineKey == key {
			return sm.inlineValue
		}
		return nil
	case mapStateHeap:
		return sm.m[key]
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}

// Pop returns the identifier set at the given key or nil if it does not exist and
// removes the key from the map.
func (sm *DefaultIDSetMap) Pop(key int) *IDSet {
	switch sm.state() {
	case mapStateEmpty:
		return nil
	case mapStateInline:
		if sm.inlineKey != key {
			return nil
		}
		v := sm.inlineValue
		sm.inlineKey = uninitSentinelKey
		sm.inlineValue = nil
		return v
	case mapStateHeap:
		v, ok := sm.m[key]
		if ok {
			sm.deleteFromMap(key)
		}
		return v
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}

// Delete removes the identifier set at the given key if it exists.
func (sm *DefaultIDSetMap) Delete(key int) {
	switch sm.state() {
	case mapStateEmpty:
		return
	case mapStateInline:
		if sm.inlineKey == key {
			sm.inlineKey = uninitSentinelKey
			sm.inlineValue = nil
		}
	case mapStateHeap:
		sm.deleteFromMap(key)
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}

func (sm *DefaultIDSetMap) deleteFromMap(key int) {
	delete(sm.m, key)
	if len(sm.m) == 1 {
		for k, v := range sm.m {
			sm.inlineKey = k
			sm.inlineValue = v
		}
		sm.m = nil
	}
}

// Each invokes the given function with each key and identifier set in the map.
//
// The order of iteration is not guaranteed to be deterministic.
func (sm *DefaultIDSetMap) Each(f func(key int, value *IDSet)) {
	switch sm.state() {
	case mapStateEmpty:
		return
	case mapStateInline:
		f(sm.inlineKey, sm.inlineValue)
	case mapStateHeap:
		for k, v := range sm.m {
			f(k, v)
		}
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}

// NumIDsForKey returns the number of identifiers in the identifier set at the given key.
func (sm *DefaultIDSetMap) NumIDsForKey(key int) int {
	switch sm.state() {
	case mapStateEmpty:
		return 0
	case mapStateInline:
		if sm.inlineKey == key {
			return sm.inlineValue.Len()
		}
	case mapStateHeap:
		if s, ok := sm.m[key]; ok {
			return s.Len()
		}
	default:
		panic(ILLEGAL_MAPSTATE)
	}
	return 0
}

// Contains determines if the given identifier belongs to the set at the given key.
func (sm *DefaultIDSetMap) Contains(key, id int) bool {
	switch sm.state() {
	case mapStateEmpty:
		return false
	case mapStateInline:
		return sm.inlineKey == key && sm.inlineValue.Contains(id)
	case mapStateHeap:
		if s, ok := sm.m[key]; ok {
			return s.Contains(id)
		}
	default:
		panic(ILLEGAL_MAPSTATE)
	}
	return false
}

// EachID invokes the given function with each identifier in the set at the given key.
//
// The order of iteration is not guaranteed to be deterministic.
func (sm *DefaultIDSetMap) EachID(key int, f func(id int)) {
	switch sm.state() {
	case mapStateEmpty:
		return
	case mapStateInline:
		if sm.inlineKey == key {
			sm.inlineValue.Each(f)
		}
	case mapStateHeap:
		if s, ok := sm.m[key]; ok {
			s.Each(f)
		}
	default:
		panic(ILLEGAL_MAPSTATE)
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
	switch sm.state() {
	case mapStateEmpty:
		sm.inlineKey = key
		sm.inlineValue = NewIDSet()
		return sm.inlineValue
	case mapStateInline:
		if sm.inlineKey == key {
			return sm.inlineValue
		}
		newValue := NewIDSet()
		sm.m = map[int]*IDSet{sm.inlineKey: sm.inlineValue, key: newValue}
		sm.inlineValue = nil
		sm.inlineKey = uninitSentinelKey
		return newValue
	case mapStateHeap:
		if s, ok := sm.m[key]; ok {
			return s
		}
		newValue := NewIDSet()
		sm.m[key] = newValue
		return newValue
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}

// compareDefaultIDSetMaps returns true if the given identifier default identifier set maps
// have equivalent keys and each key contains equivalent elements.
func compareDefaultIDSetMaps(x, y *DefaultIDSetMap) bool {
	if x == nil && y == nil {
		return true
	}

	if x.state() != y.state() {
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
	switch s.state() {
	case mapStateEmpty:
		return nil
	case mapStateInline:
		return map[int]*IDSet{s.inlineKey: s.inlineValue}
	case mapStateHeap:
		m := map[int]*IDSet{}
		for k, v := range s.m {
			m[k] = v
		}
		return m
	default:
		panic(ILLEGAL_MAPSTATE)
	}
}
