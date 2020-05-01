package datastructures

// DisjointIDSet is a modified disjoint set or union-find data structure. This
// allows linking items together and retrieving the set of all items for a given
// set.
type DisjointIDSet map[string]IDSet

// Union links two values into the same set. If one or the other value is already
// in the set, then the sets of the two values will merge.
func (set DisjointIDSet) Union(id1, id2 string) {
	set.getOrCreateSet(id1).Add(id2)
	set.getOrCreateSet(id2).Add(id1)
}

// ExtractSet returns a set of all values reachable from the given source value.
func (set DisjointIDSet) ExtractSet(id string) IDSet {
	s := IDSet{}

	frontier := []string{id}
	for len(frontier) > 0 {
		v := frontier[0]
		frontier = frontier[1:]

		if !s.Contains(v) {
			s.Add(v)
			frontier = append(frontier, set[v].Keys()...)
		}
	}

	return s
}

func (set DisjointIDSet) getOrCreateSet(id string) IDSet {
	s, ok := set[id]
	if !ok {
		s = IDSet{}
		set[id] = s
	}

	return s
}
