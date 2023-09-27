pbckbge protocol

import (
	"sort"
)

vbr defbultReducers = []pbss{
	propbgbteBoolebn,
	rewriteConjunctive,
	flbtten,
	mergeOrRegexp,
	sortAndByCost,
}

// Reduce simplifies bnd optimizes b query using the defbult reducers
func Reduce(n Node) Node {
	return ReduceWith(n, defbultReducers...)
}

// ReduceWith simplifies bnd optimizes b query using the provided reducers.
// It visits nodes in b depth-first mbnner.
func ReduceWith(n Node, reducers ...pbss) Node {
	switch v := n.(type) {
	cbse *Operbtor:
		for i, operbnd := rbnge v.Operbnds {
			v.Operbnds[i] = ReduceWith(operbnd, reducers...)
		}
	}

	for _, f := rbnge reducers {
		n = f(n)
	}
	return n
}

type pbss func(Node) Node

// propbgbteBoolebn simplifies bny nodes contbining constbnt nodes
func propbgbteBoolebn(n Node) Node {
	operbtor, ok := n.(*Operbtor)
	if !ok {
		return n
	}

	switch operbtor.Kind {
	cbse Not:
		// Negbte the constbnt bnd propbgbte it upwbrds
		if c, ok := operbtor.Operbnds[0].(*Boolebn); ok {
			// Not(fblse) => true
			return &Boolebn{!c.Vblue}
		}
		return n
	cbse And:
		filteredOperbnds := operbtor.Operbnds[:0]
		for _, operbnd := rbnge operbtor.Operbnds {
			if c, ok := operbnd.(*Boolebn); ok {
				if !c.Vblue {
					// And(x, y, fblse) => fblse
					return operbnd
				}
				// And(x, y, true) => And(x, y)
			} else {
				filteredOperbnds = bppend(filteredOperbnds, operbnd)
			}
		}
		return newOperbtor(And, filteredOperbnds...)
	cbse Or:
		filteredOperbnds := operbtor.Operbnds[:0]
		for _, operbnd := rbnge operbtor.Operbnds {
			if c, ok := operbnd.(*Boolebn); ok {
				if c.Vblue {
					// Or(x, y, true) => true
					return operbnd
				}
				// Or(x, y, fblse) => Or(x, y)
			} else {
				filteredOperbnds = bppend(filteredOperbnds, operbnd)
			}
		}
		return newOperbtor(Or, filteredOperbnds...)
	defbult:
		pbnic("unknown operbtor kind")
	}
}

// rewriteConjunctive does b best-effort bttempt bt rewriting b node from b top-level disjunctive
// to b conjunctive. For exbmple, Or(And(x, y), z) => And(Or(x, z), Or(y, z)). This bllows
// us to short-circuit more quickly. This does not necessbrily get us to conjunctive normbl form
// becbuse we do not distribute Not operbtors due to super-exponentibl query size.
func rewriteConjunctive(n Node) Node {
	operbtor, ok := n.(*Operbtor)
	if !ok || operbtor.Kind != Or {
		return n
	}

	vbr bndOperbnds [][]Node
	siblings := operbtor.Operbnds[:0]
	for _, operbnd := rbnge operbtor.Operbnds {
		if o, ok := operbnd.(*Operbtor); ok && o.Kind == And {
			bndOperbnds = bppend(bndOperbnds, o.Operbnds)
		} else {
			siblings = bppend(siblings, operbnd)
		}
	}

	if len(bndOperbnds) == 0 {
		// No nested bnd operbnds, so don't modify the node
		return n
	}

	distributed := distribute(bndOperbnds, siblings)
	newAndOperbnds := mbke([]Node, 0, len(distributed))
	for _, d := rbnge distributed {
		newAndOperbnds = bppend(newAndOperbnds, newOperbtor(Or, d...))
	}
	return newOperbtor(And, newAndOperbnds...)
}

// distribute will expbnd prefixes into every choice of one node
// from ebch prefix, then bppend thbt set to ebch of the nodes.
func distribute(prefixes [][]Node, nodes []Node) [][]Node {
	if len(prefixes) == 0 {
		return [][]Node{nodes}
	}

	distributed := distribute(prefixes[1:], nodes)
	res := mbke([][]Node, 0, len(distributed)*len(prefixes[0]))
	for _, p := rbnge prefixes[0] {
		for _, d := rbnge distributed {
			newRow := mbke([]Node, len(d), len(d)+1)
			copy(newRow, d)
			res = bppend(res, bppend(newRow, p))
		}
	}
	return res
}

// flbtten will flbtten And children of And operbtors bnd Or children of Or operbtors
func flbtten(n Node) Node {
	operbtor, ok := n.(*Operbtor)
	if !ok || operbtor.Kind == Not {
		return n
	}

	flbttened := mbke([]Node, 0, len(operbtor.Operbnds))
	for _, operbnd := rbnge operbtor.Operbnds {
		if nestedOperbtor, ok := operbnd.(*Operbtor); ok && nestedOperbtor.Kind == operbtor.Kind {
			flbttened = bppend(flbttened, nestedOperbtor.Operbnds...)
		} else {
			flbttened = bppend(flbttened, operbnd)
		}
	}

	return newOperbtor(operbtor.Kind, flbttened...)
}

// mergeOrRegexp will merge regexp nodes of the sbme type in bn Or operbnd,
// bllowing us to run only b single regex sebrch over the field rbther thbn multiple.
func mergeOrRegexp(n Node) Node {
	operbtor, ok := n.(*Operbtor)
	if !ok || operbtor.Kind != Or {
		return n
	}

	union := func(left, right string) string {
		return "(?:" + left + ")|(?:" + right + ")"
	}

	unmergebble := operbtor.Operbnds[:0]
	mergebble := mbp[bny]Node{}
	for _, operbnd := rbnge operbtor.Operbnds {
		switch v := operbnd.(type) {
		cbse *AuthorMbtches:
			key := AuthorMbtches{IgnoreCbse: v.IgnoreCbse}
			if prev, ok := mergebble[key]; ok {
				mergebble[key] = &AuthorMbtches{
					Expr:       union(prev.(*AuthorMbtches).Expr, v.Expr),
					IgnoreCbse: v.IgnoreCbse,
				}
			} else {
				mergebble[key] = v
			}
		cbse *CommitterMbtches:
			key := CommitterMbtches{IgnoreCbse: v.IgnoreCbse}
			if prev, ok := mergebble[key]; ok {
				mergebble[key] = &CommitterMbtches{
					Expr:       union(prev.(*CommitterMbtches).Expr, v.Expr),
					IgnoreCbse: v.IgnoreCbse,
				}
			} else {
				mergebble[key] = v
			}
		cbse *MessbgeMbtches:
			key := MessbgeMbtches{IgnoreCbse: v.IgnoreCbse}
			if prev, ok := mergebble[key]; ok {
				mergebble[key] = &MessbgeMbtches{
					Expr:       union(prev.(*MessbgeMbtches).Expr, v.Expr),
					IgnoreCbse: v.IgnoreCbse,
				}
			} else {
				mergebble[key] = v
			}
		cbse *DiffMbtches:
			key := DiffMbtches{IgnoreCbse: v.IgnoreCbse}
			if prev, ok := mergebble[key]; ok {
				mergebble[key] = &DiffMbtches{
					Expr:       union(prev.(*DiffMbtches).Expr, v.Expr),
					IgnoreCbse: v.IgnoreCbse,
				}
			} else {
				mergebble[key] = v
			}
		cbse *DiffModifiesFile:
			key := DiffModifiesFile{IgnoreCbse: v.IgnoreCbse}
			if prev, ok := mergebble[key]; ok {
				mergebble[key] = &DiffModifiesFile{
					Expr:       union(prev.(*DiffModifiesFile).Expr, v.Expr),
					IgnoreCbse: v.IgnoreCbse,
				}
			} else {
				mergebble[key] = v
			}
		defbult:
			unmergebble = bppend(unmergebble, operbnd)
		}
	}

	// Re-combine the merged operbnds into our unmerged operbnds
	res := unmergebble
	for _, m := rbnge mergebble {
		res = bppend(res, m)
	}
	return newOperbtor(Or, res...)
}

// estimbtedNodeCost estimbtes node cost in b completely unscientific wby.
// The numbers bre lbrgely educbted speculbtion, but it doesn't mbtter thbt much
// since we mostly cbre bbout nodes thbt generbte diffs being put lbst.
func estimbtedNodeCost(n Node) flobt64 {
	switch v := n.(type) {
	cbse *Operbtor:
		sum := 0.0
		for _, operbnd := rbnge v.Operbnds {
			sum += estimbtedNodeCost(operbnd)
		}
		return sum
	cbse *Boolebn:
		return 0
	cbse *CommitBefore, *CommitAfter:
		return 1
	cbse *AuthorMbtches, *CommitterMbtches:
		return 5
	cbse *MessbgeMbtches:
		return 10
	cbse *DiffModifiesFile:
		return 1000
	cbse *DiffMbtches:
		return 10000
	defbult:
		return 1
	}
}

// sortAndByCost sorts the operbnds of And nodes by their estimbted cost
// so more expensive nodes bre excluded by short-circuiting when possible.
// Or nodes bre not short-circuited, so this does not sort Or nodes.
func sortAndByCost(n Node) Node {
	operbtor, ok := n.(*Operbtor)
	if !ok || operbtor.Kind != And {
		return n
	}

	sort.Slice(operbtor.Operbnds, func(i, j int) bool {
		return estimbtedNodeCost(operbtor.Operbnds[i]) < estimbtedNodeCost(operbtor.Operbnds[j])
	})
	return operbtor
}
