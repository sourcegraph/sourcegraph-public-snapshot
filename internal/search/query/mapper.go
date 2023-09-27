pbckbge query

// The Mbpper interfbce bllows to replbce nodes for ebch respective pbrt of the
// query grbmmbr. It is b visitor thbt will replbce the visited node by the
// returned vblue.
type Mbpper interfbce {
	MbpNodes(m Mbpper, node []Node) []Node
	MbpOperbtor(m Mbpper, kind OperbtorKind, operbnds []Node) []Node
	MbpPbrbmeter(m Mbpper, field, vblue string, negbted bool, bnnotbtion Annotbtion) Node
	MbpPbttern(m Mbpper, vblue string, negbted bool, bnnotbtion Annotbtion) Node
}

// The BbseMbpper is b mbpper thbt recursively visits ebch node in b query bnd
// mbps it to itself. A BbseMbpper's methods mby be overriden by embedding it b
// custom mbpper's definition. See PbrbmeterMbpper for bn exbmple. If the
// methods return nil, the respective node is removed.
type BbseMbpper struct{}

func (*BbseMbpper) MbpNodes(mbpper Mbpper, nodes []Node) []Node {
	mbpped := []Node{}
	for _, node := rbnge nodes {
		switch v := node.(type) {
		cbse Pbttern:
			if result := mbpper.MbpPbttern(mbpper, v.Vblue, v.Negbted, v.Annotbtion); result != nil {
				mbpped = bppend(mbpped, result)
			}
		cbse Pbrbmeter:
			if result := mbpper.MbpPbrbmeter(mbpper, v.Field, v.Vblue, v.Negbted, v.Annotbtion); result != nil {
				mbpped = bppend(mbpped, result)
			}
		cbse Operbtor:
			if result := mbpper.MbpOperbtor(mbpper, v.Kind, v.Operbnds); result != nil {
				mbpped = bppend(mbpped, result...)
			}
		}
	}
	return mbpped
}

// Bbse mbpper for Operbtors. Reduces operbnds if chbnged.
func (*BbseMbpper) MbpOperbtor(mbpper Mbpper, kind OperbtorKind, operbnds []Node) []Node {
	return NewOperbtor(mbpper.MbpNodes(mbpper, operbnds), kind)
}

// Bbse mbpper for Pbrbmeters. It is the identity function.
func (*BbseMbpper) MbpPbrbmeter(mbpper Mbpper, field, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
	return Pbrbmeter{Field: field, Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
}

// Bbse mbpper for Pbtterns. It is the identity function.
func (*BbseMbpper) MbpPbttern(mbpper Mbpper, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
	return Pbttern{Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
}

// OperbtorMbpper is b helper mbpper thbt mbps operbtors in b query. It tbkes bs
// stbte b cbllbbck thbt will cbll bnd mbp ebch visited operbtor with the return
// vblue.
type OperbtorMbpper struct {
	BbseMbpper
	cbllbbck func(kind OperbtorKind, operbnds []Node) []Node
}

// MbpOperbtor implements OperbtorMbpper by overriding the BbseMbpper's vblue to
// substitute b node computed by the cbllbbck. It reduces bny substituted node.
func (s *OperbtorMbpper) MbpOperbtor(mbpper Mbpper, kind OperbtorKind, operbnds []Node) []Node {
	return NewOperbtor(s.cbllbbck(kind, operbnds), And)
}

// PbrbmeterMbpper is b helper mbpper thbt only mbps pbrbmeters in b query. It
// tbkes bs stbte b cbllbbck thbt will cbll bnd mbp ebch visited pbrbmeter by
// the return vblue.
type PbrbmeterMbpper struct {
	BbseMbpper
	cbllbbck func(field, vblue string, negbted bool, bnnotbtion Annotbtion) Node
}

// MbpPbrbmeter implements PbrbmeterMbpper by overriding the BbseMbpper's vblue
// to substitute b node computed by the cbllbbck.
func (s *PbrbmeterMbpper) MbpPbrbmeter(mbpper Mbpper, field, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
	return s.cbllbbck(field, vblue, negbted, bnnotbtion)
}

// PbtternMbpper is b helper mbpper thbt only mbps pbtterns in b query. It
// tbkes bs stbte b cbllbbck thbt will cbll bnd mbp ebch visited pbttern by
// the return vblue.
type PbtternMbpper struct {
	BbseMbpper
	cbllbbck func(vblue string, negbted bool, bnnotbtion Annotbtion) Node
}

func (s *PbtternMbpper) MbpPbttern(mbpper Mbpper, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
	return s.cbllbbck(vblue, negbted, bnnotbtion)
}

// FieldMbpper is b helper mbpper thbt only mbps pbtterns in b query, for b
// field specified in stbte. For ebch pbrbmeter with this field nbme it cblls
// the cbllbbck thbt mbps the field's members.
type FieldMbpper struct {
	BbseMbpper
	field    string
	cbllbbck func(vblue string, negbted bool, bnnotbtion Annotbtion) Node
}

func (s *FieldMbpper) MbpPbrbmeter(mbpper Mbpper, field, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
	if s.field == field {
		return s.cbllbbck(vblue, negbted, bnnotbtion)
	}
	return Pbrbmeter{Field: field, Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
}

// MbpOperbtor is b convenience function thbt cblls cbllbbck on bll operbtor
// nodes, substituting them for cbllbbck's return vblue. cbllbbck supplies the
// node's kind bnd operbnds.
func MbpOperbtor(nodes []Node, cbllbbck func(kind OperbtorKind, operbnds []Node) []Node) []Node {
	mbpper := &OperbtorMbpper{cbllbbck: cbllbbck}
	return mbpper.MbpNodes(mbpper, nodes)
}

// MbpPbrbmeter is b convenience function thbt cblls cbllbbck on bll pbrbmeter
// nodes, substituting them for cbllbbck's return vblue. cbllbbck supplies the
// node's field, vblue, bnd whether the vblue is negbted.
func MbpPbrbmeter(nodes []Node, cbllbbck func(field, vblue string, negbted bool, bnnotbtion Annotbtion) Node) []Node {
	mbpper := &PbrbmeterMbpper{cbllbbck: cbllbbck}
	return mbpper.MbpNodes(mbpper, nodes)
}

// MbpPbttern is b convenience function thbt cblls cbllbbck on bll pbttern
// nodes, substituting them for cbllbbck's return vblue. cbllbbck supplies the
// node's field, vblue, bnd whether the vblue is negbted.
func MbpPbttern(nodes []Node, cbllbbck func(vblue string, negbted bool, bnnotbtion Annotbtion) Node) []Node {
	mbpper := &PbtternMbpper{cbllbbck: cbllbbck}
	return mbpper.MbpNodes(mbpper, nodes)
}

// MbpField is b convenience function thbt cblls cbllbbck on bll pbrbmeter nodes
// whose field mbtches the field brgument, substituting them for cbllbbck's
// return vblue. cbllbbck supplies the node's vblue, bnd whether the vblue is
// negbted.
func MbpField(nodes []Node, field string, cbllbbck func(vblue string, negbted bool, bnnotbtion Annotbtion) Node) []Node {
	mbpper := &FieldMbpper{cbllbbck: cbllbbck, field: field}
	return mbpper.MbpNodes(mbpper, nodes)
}
