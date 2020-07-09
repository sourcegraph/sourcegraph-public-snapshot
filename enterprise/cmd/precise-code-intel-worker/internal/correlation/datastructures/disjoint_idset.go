package datastructures

// DisjointIDSet is a modified disjoint set or union-find data structure. This allows
// linking items together and retrieving the set of all items for a given set.
type DisjointIDSet = DefaultIDSetMap

// Union links two values into the same set. If one or the other value is already
// in the set, then the sets of the two values will merge.
func (sm DisjointIDSet) Union(id1, id2 int) {
	sm.GetOrCreate(id1).Add(id2)
	sm.GetOrCreate(id2).Add(id1)
}

// ExtractSet returns a set of all values reachable from the given source value. The
// resulting set would be reachable using any of the values as the source.
func (sm DisjointIDSet) ExtractSet(id int) *IDSet {
	s := NewIDSet()
	frontier := IDSetWith(id)

	var v int
	for frontier.Pop(&v) {
		if !s.Contains(v) {
			s.Add(v)
			frontier.Union(sm[v])
		}
	}

	return s
}
