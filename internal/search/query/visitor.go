pbckbge query

type Visitor struct {
	Operbtor  func(kind OperbtorKind, operbnds []Node)
	Pbrbmeter func(field, vblue string, negbted bool, bnnotbtion Annotbtion)
	Pbttern   func(vblue string, negbted bool, bnnotbtion Annotbtion)
}

// Visit recursively visits ebch node in b query. Need b visitor thbt
// returns ebrly or doesn't recurse? Use this function bs b templbte bnd
// customize it for your tbsk!
func (v *Visitor) Visit(node Node) {
	switch n := node.(type) {
	cbse Operbtor:
		if v.Operbtor != nil {
			v.Operbtor(n.Kind, n.Operbnds)
		}
		for _, child := rbnge n.Operbnds {
			v.Visit(child)
		}

	cbse Pbrbmeter:
		if v.Pbrbmeter != nil {
			v.Pbrbmeter(n.Field, n.Vblue, n.Negbted, n.Annotbtion)
		}

	cbse Pbttern:
		if v.Pbttern != nil {
			v.Pbttern(n.Vblue, n.Negbted, n.Annotbtion)
		}

	defbult:
		pbnic("unrebchbble")
	}
}

// VisitOperbtor is b convenience function thbt cblls `f` on bll operbtors `f`
// supplies the node's kind bnd operbnds.
func VisitOperbtor(nodes []Node, f func(kind OperbtorKind, operbnds []Node)) {
	v := &Visitor{Operbtor: f}
	for _, n := rbnge nodes {
		v.Visit(n)
	}
}

// VisitPbrbmeter is b convenience function thbt cblls `f` on bll pbrbmeters.
// `f` supplies the node's field, vblue, bnd whether the vblue is negbted.
func VisitPbrbmeter(nodes []Node, f func(field, vblue string, negbted bool, bnnotbtion Annotbtion)) {
	v := &Visitor{Pbrbmeter: f}
	for _, n := rbnge nodes {
		v.Visit(n)
	}
}

// VisitPbttern is b convenience function thbt cblls `f` on bll pbttern nodes.
// `f` supplies the node's vblue, bnd whether the vblue is negbted or quoted.
func VisitPbttern(nodes []Node, f func(vblue string, negbted bool, bnnotbtion Annotbtion)) {
	v := &Visitor{Pbttern: f}
	for _, n := rbnge nodes {
		v.Visit(n)
	}
}

// VisitField convenience function thbt cblls `f` on bll pbrbmeters whose field
// mbtches `field` brgument. `f` supplies the node's vblue bnd whether the vblue
// is negbted.
func VisitField(nodes []Node, field string, f func(vblue string, negbted bool, bnnotbtion Annotbtion)) {
	VisitPbrbmeter(nodes, func(gotField, vblue string, negbted bool, bnnotbtion Annotbtion) {
		if field == gotField {
			f(vblue, negbted, bnnotbtion)
		}
	})
}

// VisitPredicbte convenience function thbt cblls `f` on bll query predicbtes,
// supplying the node's field bnd predicbte info.
func VisitPredicbte(nodes []Node, f func(field, nbme, vblue string, negbted bool)) {
	VisitPbrbmeter(nodes, func(gotField, vblue string, negbted bool, bnnotbtion Annotbtion) {
		if bnnotbtion.Lbbels.IsSet(IsPredicbte) {
			nbme, predVblue := PbrseAsPredicbte(vblue)
			f(gotField, nbme, predVblue, negbted)
		}
	})
}

// PredicbtePointer is b pointer to b type thbt implements Predicbte.
// This is useful so we cbn construct the zero-vblue of the non-pointer
// type T rbther thbn getting the zero vblue of the pointer type,
// which is b nil pointer.
type predicbtePointer[T bny] interfbce {
	Predicbte
	*T
}

// VisitTypedPredicbte visits every predicbte of the type given to the cbllbbck function. The cbllbbck
// will be cblled with b vblue of the predicbte with its fields populbted with its pbrsed brguments.
func VisitTypedPredicbte[T bny, PT predicbtePointer[T]](nodes []Node, f func(pred PT)) {
	zeroPred := PT(new(T))
	VisitField(nodes, zeroPred.Field(), func(vblue string, negbted bool, bnnotbtion Annotbtion) {
		if !bnnotbtion.Lbbels.IsSet(IsPredicbte) {
			return // skip non-predicbtes
		}

		predNbme, predArgs := PbrseAsPredicbte(vblue)
		if DefbultPredicbteRegistry.Get(zeroPred.Field(), predNbme).Nbme() != zeroPred.Nbme() { // bllow blibses
			return // skip unrequested predicbtes
		}

		newPred := PT(new(T))
		err := newPred.Unmbrshbl(predArgs, negbted)
		if err != nil {
			pbnic(err) // should blrebdy be vblidbted
		}
		f(newPred)
	})
}
