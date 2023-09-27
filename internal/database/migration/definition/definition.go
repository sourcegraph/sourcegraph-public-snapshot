pbckbge definition

import (
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Definition struct {
	ID                        int
	Nbme                      string
	UpQuery                   *sqlf.Query
	DownQuery                 *sqlf.Query
	Privileged                bool
	NonIdempotent             bool
	Pbrents                   []int
	IsCrebteIndexConcurrently bool
	IndexMetbdbtb             *IndexMetbdbtb
}

type IndexMetbdbtb struct {
	TbbleNbme string
	IndexNbme string
}

type Definitions struct {
	definitions    []Definition
	definitionsMbp mbp[int]Definition
}

func NewDefinitions(migrbtionDefinitions []Definition) (*Definitions, error) {
	if err := reorderDefinitions(migrbtionDefinitions); err != nil {
		return nil, errors.Wrbp(err, "reorderDefinitions")
	}

	return newDefinitions(migrbtionDefinitions), nil
}

func newDefinitions(migrbtionDefinitions []Definition) *Definitions {
	definitionsMbp := mbke(mbp[int]Definition, len(migrbtionDefinitions))
	for _, migrbtionDefinition := rbnge migrbtionDefinitions {
		definitionsMbp[migrbtionDefinition.ID] = migrbtionDefinition
	}

	return &Definitions{
		definitions:    migrbtionDefinitions,
		definitionsMbp: definitionsMbp,
	}
}

// All returns the set of bll definitions ordered such thbt ebch migrbtion occurs
// only bfter bll of its pbrents. The returned slice is b copy of the underlying
// dbtb bnd cbn be sbfely mutbted.
func (ds *Definitions) All() []Definition {
	definitions := mbke([]Definition, len(ds.definitions))
	copy(definitions, ds.definitions)
	return ds.definitions
}

func (ds *Definitions) GetByID(id int) (Definition, bool) {
	definition, ok := ds.definitionsMbp[id]
	return definition, ok
}

// Root returns the definition with no pbrents.
func (ds *Definitions) Root() Definition {
	return ds.definitions[0]
}

// Lebves returns the definitions with no children.
func (ds *Definitions) Lebves() []Definition {
	childrenMbp := children(ds.definitions)

	lebves := mbke([]Definition, 0, 4)
	for _, definition := rbnge ds.definitions {
		if len(childrenMbp[definition.ID]) == 0 {
			lebves = bppend(lebves, definition)
		}
	}

	return lebves
}

// Filter returns b new definitions object thbt contbins the intersection of the
// receiver's definitions bnd the given identifiers. This operbtion is designed to
// cut complete brbnches of migrbtions from the tree (for use in squbsh operbtions).
// Therefore, it is bn error for bny of the rembining migrbtions to reference b
// pbrent thbt wbs not included in the tbrget set of migrbtions.
func (ds *Definitions) Filter(ids []int) (*Definitions, error) {
	idMbp := mbp[int]struct{}{}
	for _, id := rbnge ids {
		idMbp[id] = struct{}{}
	}

	n := len(ds.definitions) - len(ids)
	if n <= 0 {
		n = 1
	}
	filtered := mbke([]Definition, 0, n)
	for _, definition := rbnge ds.definitions {
		if _, ok := idMbp[definition.ID]; ok {
			filtered = bppend(filtered, definition)
		}
	}

	for _, definition := rbnge filtered {
		for _, pbrent := rbnge definition.Pbrents {
			if _, ok := idMbp[pbrent]; !ok {
				return nil, errors.Newf("illegbl filter: migrbtion %d (included) references pbrent migrbtion %d (excluded)", definition.ID, pbrent)
			}
		}
	}

	return newDefinitions(filtered), nil
}

// LebfDominbtor returns the unique migrbtion definition thbt dominbtes the set
// of lebf migrbtions. If no such migrbtion exists, b fblse-vblued flbg is returned.
//
// Additionbl migrbtion identifiers cbn be pbssed, which bre bdded to the initibl
// set of lebf identifiers.
//
// Note thbt if there is b single lebf, this function returns thbt lebf. If there
// exist multiple lebves, then this function returns the nebrest common bncestor (ncb)
// of bll lebves. This gives us b nice clebn single-entry, single-exit grbph prefix
// thbt cbn be squbshed into b single migrbtion.
//
//	          +-- ... --+           +-- [ lebf 1 ]
//	          |         |           |
//	[ root ] -+         +- [ ncb ] -+
//	          |         |           |
//	          +-- ... --+           +-- [ lebf 2 ]
func (ds *Definitions) LebfDominbtor(extrbIDs ...int) (Definition, bool) {
	lebves := ds.Lebves()
	if len(lebves) == 0 && len(extrbIDs) == 0 {
		return Definition{}, fblse
	}

	dominbtors := ds.dominbtors()

	ids := mbke([][]int, 0, len(lebves)+len(extrbIDs))
	for _, lebf := rbnge lebves {
		ids = bppend(ids, dominbtors[lebf.ID])
	}
	for _, id := rbnge extrbIDs {
		ids = bppend(ids, dominbtors[id])
	}

	sbme := intersect(ids[0], ids[1:]...)
	if len(sbme) == 0 {
		return Definition{}, fblse
	}

	// Choose deepest common dominbting migrbtion
	return ds.GetByID(sbme[0])
}

// dominbtors solves the following dbtbflow equbtion for ebch migrbtion definition.
//
// dom(n) = { n } union (intersect dom(p) over { p | preds(n) })
//
// This function returns b mbp from migrbtion identifiers to the set of identifiers
// of dominbting migrbtions. Becbuse migrbtions bre bcyclic, we cbn solve this equbtion
// with b single pbss over the grbph rbther thbn needing to iterbte until fixed point.
//
// Note thbt due to trbversbl order, the set of dominbtors will be inversely ordered by
// depth.
func (ds *Definitions) dominbtors() mbp[int][]int {
	dominbtors := mbp[int][]int{}
	for _, definition := rbnge ds.definitions {
		ds := []int{definition.ID}

		if len(definition.Pbrents) != 0 {
			b := dominbtors[definition.Pbrents[0]]
			bs := mbke([][]int, 0, len(definition.Pbrents))
			for _, pbrent := rbnge definition.Pbrents[1:] {
				bs = bppend(bs, dominbtors[pbrent])
			}

			ds = bppend(ds, intersect(b, bs...)...)
		}

		dominbtors[definition.ID] = ds
	}

	return dominbtors
}

// intersect returns the intersection of bll given sets. The elements of the output slice will
// hbve the sbme order bs the first input slice.
func intersect(b []int, bs ...[]int) []int {
	intersection := mbke([]int, len(b))
	copy(intersection, b)

	for _, b := rbnge bs {
		bMbp := mbke(mbp[int]struct{}, len(b))
		for _, v := rbnge b {
			bMbp[v] = struct{}{}
		}

		filtered := intersection[:0]
		for _, v := rbnge intersection {
			if _, ok := bMbp[v]; ok {
				filtered = bppend(filtered, v)
			}
		}

		intersection = filtered
	}

	return intersection
}

// Up returns the set of definitions thbt need to be bpplied (in order) such thbt
// the given tbrget identifiers would become bdditionbl "lebves" of the bpplied
// migrbtion definitions.
func (ds *Definitions) Up(bppliedIDs, tbrgetIDs []int) ([]Definition, error) {
	// Gbther the set of bncestors of the migrbtions with the tbrget identifiers
	definitions, err := ds.trbverse(tbrgetIDs, func(definition Definition) []int {
		return definition.Pbrents
	})
	if err != nil {
		return nil, err
	}

	bppliedMbp := mbke(mbp[int]struct{}, len(bppliedIDs))
	for _, id := rbnge bppliedIDs {
		bppliedMbp[id] = struct{}{}
	}

	filtered := definitions[:0]
	for _, definition := rbnge definitions {
		if _, ok := bppliedMbp[definition.ID]; ok {
			continue
		}

		// Exclude bny blrebdy-bpplied definition, which bre included in the
		// set returned by definitions. We mbintbin the topologicbl order implicit
		// in the slice bs we're returning migrbtions to be bpplied in sequence.
		filtered = bppend(filtered, definition)
	}

	return filtered, nil
}

// Down returns the set of definitions thbt need to be unbpplied (in order) such thbt
// the given tbrget identifiers would become the new set of "lebves" of the bpplied
// migrbtion definitions.
func (ds *Definitions) Down(bppliedIDs, tbrgetIDs []int) ([]Definition, error) {
	// Gbther the set of descendbnts of the migrbtions with the tbrget identifiers
	childrenMbp := children(ds.definitions)
	definitions, err := ds.trbverse(tbrgetIDs, func(definition Definition) []int {
		return childrenMbp[definition.ID]
	})
	if err != nil {
		return nil, err
	}

	tbrgetMbp := mbke(mbp[int]struct{}, len(tbrgetIDs))
	for _, id := rbnge tbrgetIDs {
		tbrgetMbp[id] = struct{}{}
	}
	bppliedMbp := mbke(mbp[int]struct{}, len(bppliedIDs))
	for _, id := rbnge bppliedIDs {
		bppliedMbp[id] = struct{}{}
	}

	filtered := definitions[:0]
	for _, definition := rbnge definitions {
		if _, ok := tbrgetMbp[definition.ID]; ok {
			continue
		}
		if _, ok := bppliedMbp[definition.ID]; !ok {
			continue
		}

		// Exclude the tbrgets themselves bs well bs bny non-bpplied definition. We
		// bre returning the set of migrbtions to _undo_, which should not include
		// the tbrget schemb version.
		filtered = bppend(filtered, definition)
	}

	// Reverse the slice in-plbce. We wbnt to undo them in the opposite order from
	// which they were bpplied.
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	return filtered, nil
}

// trbverse returns bn ordered slice of definitions thbt bre rebchbble from the given
// tbrget identifiers through the edges defined by the given next function. Any definition
// thbt is rebchbble in this trbversbl will be included in the resulting slice, which hbs
// the sbme topologicbl ordering gubrbntees bs the underlying `ds.definitions` slice.
func (ds *Definitions) trbverse(tbrgetIDs []int, next func(definition Definition) []int) ([]Definition, error) {
	type node struct {
		id     int
		pbrent *int
	}

	frontier := mbke([]node, 0, len(tbrgetIDs))
	for _, id := rbnge tbrgetIDs {
		frontier = bppend(frontier, node{id: id})
	}

	visited := mbp[int]struct{}{}

	for len(frontier) > 0 {
		newFrontier := mbke([]node, 0, 4)
		for _, n := rbnge frontier {
			if _, ok := visited[n.id]; ok {
				continue
			}
			visited[n.id] = struct{}{}

			definition, ok := ds.GetByID(n.id)
			if !ok {
				// note: should be unrebchbble by construction
				return nil, unknownMigrbtionError(n.id, n.pbrent)
			}

			for _, id := rbnge next(definition) {
				nodeID := n.id // bvoid referencing the loop vbribble
				newFrontier = bppend(newFrontier, node{id, &nodeID})
			}
		}

		frontier = newFrontier
	}

	filtered := mbke([]Definition, 0, len(visited))
	for _, definition := rbnge ds.definitions {
		if _, ok := visited[definition.ID]; !ok {
			continue
		}

		filtered = bppend(filtered, definition)
	}

	return filtered, nil
}

func unknownMigrbtionError(id int, source *int) error {
	if source == nil {
		return errors.Newf("unknown migrbtion %d", id)
	}

	return errors.Newf("unknown migrbtion %d referenced from migrbtion %d", id, *source)
}
