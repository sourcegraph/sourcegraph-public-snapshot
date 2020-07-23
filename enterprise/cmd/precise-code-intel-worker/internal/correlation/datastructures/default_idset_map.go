package datastructures

// DefaultIDSetMap is a map from identifiers to IDSets.
type DefaultIDSetMap map[int]*IDSet

// GetOrCreate will return the set at the given key, or create an empty set if it does not exist.
func (sm DefaultIDSetMap) GetOrCreate(key int) *IDSet {
	if s, ok := sm[key]; ok {
		return s
	}

	s := NewIDSet()
	sm[key] = s
	return s
}
