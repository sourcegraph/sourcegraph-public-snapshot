package store

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/alecthomas/binary"
	"github.com/smartystreets/mafsa"

	"strings"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

type defQueryIndex struct {
	mt    *mafsaTable
	f     DefFilter
	ready bool
	sync.RWMutex
}

var _ interface {
	Index
	persistedIndex
	defIndexBuilder
	defIndex
} = (*defQueryIndex)(nil)

var c_defQueryIndex_getByQuery = &counter{count: new(int64)}

func (x *defQueryIndex) String() string { return fmt.Sprintf("defQueryIndex(ready=%v)", x.ready) }

func (x *defQueryIndex) getByQuery(q string) (byteOffsets, bool) {
	vlog.Printf("defQueryIndex.getByQuery(%q)", q)
	c_defQueryIndex_getByQuery.increment()

	if x.mt == nil {
		panic("mafsaTable not built/read")
	}

	q = strings.ToLower(q)
	node, i := x.mt.t.IndexedTraverse([]rune(q))
	if node == nil {
		return nil, false
	}
	nn := node.Number
	if node.Final {
		i--
		nn++
	}
	var ofs byteOffsets
	for _, ofs0 := range x.mt.Values[i : i+nn] {
		ofs = append(ofs, ofs0...)
	}
	vlog.Printf("defQueryIndex.getByQuery(%q): found %d defs.", q, len(ofs))
	return ofs, true
}

// Covers implements defIndex.
func (x *defQueryIndex) Covers(filters interface{}) int {
	cov := 0
	for _, f := range storeFilters(filters) {
		if _, ok := f.(ByDefQueryFilter); ok {
			cov++
		}
	}
	return cov
}

// Defs implements defIndex.
func (x *defQueryIndex) Defs(f ...DefFilter) (byteOffsets, error) {
	x.RLock()
	defer x.RUnlock()
	for _, ff := range f {
		if pf, ok := ff.(ByDefQueryFilter); ok {
			ofs, found := x.getByQuery(pf.ByDefQuery())
			if !found {
				return nil, nil
			}
			return ofs, nil
		}
	}
	return nil, nil
}

type defLowerNameAndOffset struct {
	lowerName string
	ofs       int64
}

type defsByLowerName []*defLowerNameAndOffset

func (ds defsByLowerName) Len() int           { return len(ds) }
func (ds defsByLowerName) Swap(i, j int)      { ds[i], ds[j] = ds[j], ds[i] }
func (ds defsByLowerName) Less(i, j int) bool { return ds[i].lowerName < ds[j].lowerName }

// Build implements defIndexBuilder.
func (x *defQueryIndex) Build(defs []*graph.Def, ofs byteOffsets) (err error) {
	x.Lock()
	defer x.Unlock()
	vlog.Printf("defQueryIndex: building index... (%d defs)", len(defs))

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in defQueryIndex.Build (%d defs): %v", len(defs), err)
		}
	}()

	// Clone slice so we can sort it by whatever we want.
	dofs := make([]*defLowerNameAndOffset, 0, len(defs))
	for i, def := range defs {
		if x.f.SelectDef(def) && !hasNonASCIIChars(def.Name) {
			// See https://github.com/smartystreets/mafsa/issues/1 for
			// why we need to kick out non-ASCII.

			dofs = append(dofs, &defLowerNameAndOffset{strings.ToLower(def.Name), ofs[i]})
		}
	}
	if len(dofs) == 0 {
		x.mt = &mafsaTable{}
		x.ready = true
		return nil
	}
	sort.Sort(defsByLowerName(dofs))
	vlog.Printf("defQueryIndex: done sorting by def name (%d defs).", len(defs))

	bt := mafsa.New()
	x.mt = &mafsaTable{}
	x.mt.Values = make([]byteOffsets, 0, len(dofs))
	j := 0 // index of earliest def with same name
	for i, def := range dofs {
		if i > 0 && dofs[j].lowerName == def.lowerName {
			x.mt.Values[len(x.mt.Values)-1] = append(x.mt.Values[len(x.mt.Values)-1], def.ofs)
		} else {
			bt.Insert(def.lowerName)
			x.mt.Values = append(x.mt.Values, byteOffsets{def.ofs})
			j = i
		}
	}
	bt.Finish()
	vlog.Printf("defQueryIndex: done adding %d defs to MAFSA & table and minimizing.", len(defs))

	b, err := bt.MarshalBinary()
	if err != nil {
		return err
	}
	vlog.Printf("defQueryIndex: done serializing MAFSA & table to %d bytes.", len(b))

	x.mt.B = b
	x.mt.t, err = new(mafsa.Decoder).Decode(x.mt.B)
	if err != nil {
		return err
	}
	x.ready = true
	vlog.Printf("defQueryIndex: done building index (%d defs).", len(defs))
	return nil
}

// Write implements persistedIndex.
func (x *defQueryIndex) Write(w io.Writer) error {
	x.RLock()
	defer x.RUnlock()
	if x.mt == nil {
		panic("no mafsaTable to write")
	}
	b, err := binary.Marshal(x.mt)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// Read implements persistedIndex.
func (x *defQueryIndex) Read(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	x.Lock()
	defer x.Unlock()
	var mt mafsaTable
	err = binary.Unmarshal(b, &mt)
	x.mt = &mt
	if err == nil && len(x.mt.B) > 0 {
		x.mt.t, err = new(mafsa.Decoder).Decode(x.mt.B)
	}
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *defQueryIndex) Ready() bool {
	x.RLock()
	defer x.RUnlock()
	return x.ready
}

// Fprint prints a human-readable representation of the index.
func (x *defQueryIndex) Fprint(w io.Writer) error {
	x.RLock()
	defer x.RUnlock()
	if x.mt == nil {
		panic("mafsaTable not built/read")
	}

	allTerms := make([]string, 0, len(x.mt.Values))
	var getAllTerms func(term string, n *mafsa.MinTreeNode)
	getAllTerms = func(term string, n *mafsa.MinTreeNode) {
		if n.Final {
			allTerms = append(allTerms, term)
		}
		for _, c := range n.OrderedEdges() {
			getAllTerms(term+string([]rune{c}), n.Edges[c])
		}
	}

	getAllTerms("", x.mt.t.Root)
	fmt.Fprintln(w, "Terms")
	for i, term := range allTerms {
		fmt.Fprintf(w, "  %d - %q\n", i, term)
	}

	fmt.Fprintln(w)

	fmt.Fprintln(w, "Unit offsets")
	for i, ofs := range x.mt.Values {
		fmt.Fprintf(w, "Term %q (node %d)\n", allTerms[i], i)
		for _, ofs := range ofs {
			fmt.Fprintf(w, "\t\t%d\n", ofs)
		}
	}

	return nil
}

// A mafsaTable is a minimal perfect hashed MA-FSA with an associated
// table of values for each entry in the MA-FSA (indexed on the
// entry's hash value).
type mafsaTable struct {
	t      *mafsa.MinTree
	B      []byte        // bytes of the MinTree
	Values []byteOffsets // one value per entry in build or min
}

func hasNonASCIIChars(s string) bool {
	for _, c := range s {
		if c < 0 || c >= 128 {
			return true
		}
	}
	return false
}
