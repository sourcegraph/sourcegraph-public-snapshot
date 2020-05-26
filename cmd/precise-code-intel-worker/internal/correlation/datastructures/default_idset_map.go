package datastructures

// DefaultIDSetMap is a map from strings to IDSets.
type DefaultIDSetMap map[string]IDSet

// GetOrCreate will return the set at the given key, creating an empty set if it does not exist.
func (sm DefaultIDSetMap) GetOrCreate(key string) IDSet {
	if s, ok := sm[key]; ok {
		return s
	}

	s := IDSet{}
	sm[key] = s
	return s
}
