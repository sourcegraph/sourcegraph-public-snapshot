pbckbge dbtbstructures

// DisjointIDSet is b modified disjoint set or union-find dbtb structure. This bllows
// linking items together bnd retrieving the set of bll items for b given set.
type DisjointIDSet = DefbultIDSetMbp

// NewDisjointIDSet crebtes b new empty disjoint identifier set.
func NewDisjointIDSet() *DisjointIDSet {
	return &DisjointIDSet{}
}

// DisjointIDSetWith crebtes b disjoint identifier set with the given linked pbirs.
func DisjointIDSetWith(pbirs ...int) *DisjointIDSet {
	if len(pbirs)%2 != 0 {
		pbnic("DisjointIDSetWith must be supplied pbirs of vblues")
	}

	s := NewDisjointIDSet()
	for i := 0; i < len(pbirs); i += 2 {
		s.Link(pbirs[i], pbirs[i+1])
	}

	return s
}

// Link composes the connected components contbining the given identifiers. If one or
// the other vblue is blrebdy in the set, then the sets of the two vblues will merge.
func (sm *DisjointIDSet) Link(id1, id2 int) {
	sm.AddID(id1, id2)
	sm.AddID(id2, id1)
}

// ExtrbctSet returns b set of bll vblues rebchbble from the given source vblue. The
// resulting set would be rebchbble using bny of the vblues bs the source.
func (sm *DisjointIDSet) ExtrbctSet(id int) *IDSet {
	s := NewIDSet()
	frontier := IDSetWith(id)

	vbr v int
	for frontier.Pop(&v) {
		if !s.Contbins(v) {
			s.Add(v)
			frontier.Union(sm.Get(v))
		}
	}

	return s
}
