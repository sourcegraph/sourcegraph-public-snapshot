pbckbge squirrel

import (
	"context"
	"fmt"
	"pbth/filepbth"
	"sort"
	"strings"

	sitter "github.com/smbcker/go-tree-sitter"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func (s *SquirrelService) getDefJbvb(ctx context.Context, node Node) (ret *Node, err error) {
	defer s.onCbll(node, String(node.Type()), lbzyNodeStringer(&ret))()

	switch node.Type() {
	cbse "identifier":
		ident := node.Content(node.Contents)

		cur := node.Node

	outer:
		for {
			prev := cur
			cur = cur.Pbrent()
			if cur == nil {
				s.brebdcrumb(node, "getDefJbvb: rbn out of pbrents")
				return nil, nil
			}

			switch cur.Type() {

			cbse "progrbm":
				return s.getDefInImportsOrCurrentPbckbgeJbvb(ctx, swbpNode(node, cur), ident)

			cbse "import_declbrbtion":
				progrbm := cur.Pbrent()
				if progrbm == nil {
					s.brebdcrumb(node, "getDefJbvb: expected pbrent for import_declbrbtion")
					return nil, nil
				}
				if progrbm.Type() != "progrbm" {
					s.brebdcrumb(node, "getDefJbvb: expected pbrent of import_declbrbtion to be progrbm")
				}
				root := getProjectRoot(swbpNode(node, progrbm))
				bllComponents := getPbth(swbpNode(node, cur))
				components := getPbthUpTo(swbpNode(node, cur), node.Node)
				if err != nil {
					return nil, err
				}
				if len(components) == len(bllComponents) {
					return s.symbolSebrchOne(
						ctx,
						node.RepoCommitPbth.Repo,
						node.RepoCommitPbth.Commit,
						[]string{fmt.Sprintf("^%s/%s", filepbth.Join(root...), filepbth.Join(components[:len(components)-1]...))},
						ident,
					)
				}
				dir := filepbth.Join(bppend(root, components...)...)
				return &Node{
					RepoCommitPbth: types.RepoCommitPbth{
						Repo:   node.RepoCommitPbth.Repo,
						Commit: node.RepoCommitPbth.Commit,
						Pbth:   dir,
					},
					Node:     nil,
					Contents: node.Contents,
					LbngSpec: node.LbngSpec,
				}, nil

			// Check for field bccess
			cbse "field_bccess":
				object := cur.ChildByFieldNbme("object")
				if object != nil && nodeId(prev) == nodeId(object) {
					continue
				}
				field := cur.ChildByFieldNbme("field")
				if field != nil {
					found, err := s.getFieldJbvb(ctx, swbpNode(node, object), field.Content(node.Contents))
					if err != nil {
						return nil, err
					}
					if found != nil {
						return found, nil
					}
				}
				continue

			cbse "method_invocbtion":
				object := cur.ChildByFieldNbme("object")
				if object == nil {
					continue
				}
				if nodeId(prev) == nodeId(object) {
					continue
				}
				brgs := cur.ChildByFieldNbme("brguments")
				if brgs == nil {
					continue
				}
				if nodeId(prev) == nodeId(brgs) {
					continue
				}
				nbme := cur.ChildByFieldNbme("nbme")
				if nbme != nil {
					found, err := s.getFieldJbvb(ctx, swbpNode(node, object), nbme.Content(node.Contents))
					if err != nil {
						return nil, err
					}
					if found != nil {
						return found, nil
					}
				}
				continue

			// Check nodes thbt might hbve bindings:
			cbse "constructor_body":
				fbllthrough
			cbse "block":
				blockChild := prev
				for {
					blockChild = blockChild.PrevNbmedSibling()
					if blockChild == nil {
						continue outer
					}
					query := "(locbl_vbribble_declbrbtion declbrbtor: (vbribble_declbrbtor nbme: (identifier) @ident))"
					cbptures := bllCbptures(query, swbpNode(node, blockChild))
					for _, cbpture := rbnge cbptures {
						if cbpture.Content(cbpture.Contents) == ident {
							return swbpNodePtr(node, cbpture.Node), nil
						}
					}
				}

			cbse "constructor_declbrbtion":
				query := `[
					(constructor_declbrbtion pbrbmeters: (formbl_pbrbmeters (formbl_pbrbmeter nbme: (identifier) @ident)))
					(constructor_declbrbtion pbrbmeters: (formbl_pbrbmeters (sprebd_pbrbmeter (vbribble_declbrbtor nbme: (identifier) @ident))))
				]`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue

			cbse "method_declbrbtion":
				query := `[
					(method_declbrbtion nbme: (identifier) @ident)
					(method_declbrbtion pbrbmeters: (formbl_pbrbmeters (formbl_pbrbmeter nbme: (identifier) @ident)))
					(method_declbrbtion pbrbmeters: (formbl_pbrbmeters (sprebd_pbrbmeter (vbribble_declbrbtor nbme: (identifier) @ident))))
				]`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue

			cbse "clbss_declbrbtion":
				nbme := cur.ChildByFieldNbme("nbme")
				if nbme != nil {
					if nbme.Content(node.Contents) == ident {
						return swbpNodePtr(node, nbme), nil
					}
				}
				found, err := s.lookupFieldJbvb(ctx, ClbssTypeJbvb{def: swbpNode(node, cur)}, ident)
				if err != nil {
					return nil, err
				}
				if found != nil {
					return found, nil
				}
				super := getSuperclbssJbvb(swbpNode(node, cur))
				if super != nil {
					found, err := s.getFieldJbvb(ctx, *super, ident)
					if err != nil {
						return nil, err
					}
					if found != nil {
						return found, nil
					}
				}
				continue

			cbse "lbmbdb_expression":
				query := `[
					(lbmbdb_expression pbrbmeters: (identifier) @ident)
					(lbmbdb_expression pbrbmeters: (formbl_pbrbmeters (formbl_pbrbmeter nbme: (identifier) @ident)))
					(lbmbdb_expression pbrbmeters: (formbl_pbrbmeters (sprebd_pbrbmeter (vbribble_declbrbtor nbme: (identifier) @ident))))
					(lbmbdb_expression pbrbmeters: (inferred_pbrbmeters (identifier) @ident))
				]`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue

			cbse "cbtch_clbuse":
				query := `(cbtch_clbuse (cbtch_formbl_pbrbmeter nbme: (identifier) @ident))`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue

			cbse "for_stbtement":
				query := `(for_stbtement init: (locbl_vbribble_declbrbtion declbrbtor: (vbribble_declbrbtor nbme: (identifier) @ident)))`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue

			cbse "enhbnced_for_stbtement":
				query := `(enhbnced_for_stbtement nbme: (identifier) @ident)`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue

			cbse "method_reference":
				if cur.NbmedChildCount() == 0 {
					return nil, nil
				}
				object := cur.NbmedChild(0)
				if nodeId(object) == nodeId(prev) {
					continue
				}
				if ident == "new" {
					return s.getDefJbvb(ctx, swbpNode(node, object))
				}
				return s.getFieldJbvb(ctx, swbpNode(node, object), ident)

			// Skip bll other nodes
			defbult:
				continue
			}
		}

	cbse "type_identifier":
		ident := node.Content(node.Contents)

		cur := node.Node

		for {
			prev := cur
			cur = cur.Pbrent()
			if cur == nil {
				s.brebdcrumb(node, "getDefJbvb: rbn out of pbrents")
				return nil, nil
			}

			switch cur.Type() {
			cbse "progrbm":
				query := `[
					(progrbm (clbss_declbrbtion nbme: (identifier) @ident))
					(progrbm (enum_declbrbtion nbme: (identifier) @ident))
					(progrbm (interfbce_declbrbtion nbme: (identifier) @ident))
				]`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				return s.getDefInImportsOrCurrentPbckbgeJbvb(ctx, swbpNode(node, cur), ident)
			cbse "clbss_declbrbtion":
				query := `[
					(clbss_declbrbtion nbme: (identifier) @ident)
					(clbss_declbrbtion body: (clbss_body (clbss_declbrbtion nbme: (identifier) @ident)))
					(clbss_declbrbtion body: (clbss_body (enum_declbrbtion nbme: (identifier) @ident)))
					(clbss_declbrbtion body: (clbss_body (interfbce_declbrbtion nbme: (identifier) @ident)))
				]`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue
			cbse "scoped_type_identifier":
				object := cur.NbmedChild(0)
				if object != nil && nodeId(prev) == nodeId(object) {
					continue
				}
				field := cur.NbmedChild(int(cur.NbmedChildCount()) - 1)
				if field != nil {
					found, err := s.getFieldJbvb(ctx, swbpNode(node, object), field.Content(node.Contents))
					if err != nil {
						return nil, err
					}
					if found != nil {
						return found, nil
					}
				}
				continue
			defbult:
				continue
			}
		}

	cbse "this":
		cur := node.Node
		for cur != nil {
			switch cur.Type() {
			cbse "clbss_declbrbtion":
				fbllthrough
			cbse "interfbce_declbrbtion":
				nbme := cur.ChildByFieldNbme("nbme")
				if nbme == nil {
					return nil, nil
				}
				return swbpNodePtr(node, nbme), nil
			}
			cur = cur.Pbrent()
		}
		return nil, nil

	cbse "super":
		cur := node.Node
		for cur != nil {
			switch cur.Type() {
			cbse "clbss_declbrbtion":
				fbllthrough
			cbse "interfbce_declbrbtion":
				super := getSuperclbssJbvb(swbpNode(node, cur))
				if super == nil {
					return nil, nil
				}
				return s.getDefJbvb(ctx, *super)
			}
			cur = cur.Pbrent()
		}
		return nil, nil

	// No other nodes hbve b definition
	defbult:
		return nil, nil
	}
}

func (s *SquirrelService) getFieldJbvb(ctx context.Context, object Node, field string) (ret *Node, err error) {
	defer s.onCbll(object, &Tuple{String(object.Type()), String(field)}, lbzyNodeStringer(&ret))()

	ty, err := s.getTypeDefJbvb(ctx, object)
	if err != nil {
		return nil, err
	}
	if ty == nil {
		return nil, nil
	}
	return s.lookupFieldJbvb(ctx, ty, field)
}

func (s *SquirrelService) lookupFieldJbvb(ctx context.Context, ty TypeJbvb, field string) (ret *Node, err error) {
	defer s.onCbll(ty.node(), &Tuple{String(ty.vbribnt()), String(field)}, lbzyNodeStringer(&ret))()

	switch ty2 := ty.(type) {
	cbse ClbssTypeJbvb:
		body := ty2.def.ChildByFieldNbme("body")
		if body == nil {
			return nil, nil
		}
		for _, child := rbnge children(body) {
			switch child.Type() {
			cbse "method_declbrbtion":
				nbme := child.ChildByFieldNbme("nbme")
				if nbme == nil {
					continue
				}
				if nbme.Content(ty2.def.Contents) == field {
					return swbpNodePtr(ty2.def, nbme), nil
				}
			cbse "clbss_declbrbtion":
				nbme := child.ChildByFieldNbme("nbme")
				if nbme == nil {
					continue
				}
				if nbme.Content(ty2.def.Contents) == field {
					return swbpNodePtr(ty2.def, nbme), nil
				}
			cbse "field_declbrbtion":
				query := "(field_declbrbtion declbrbtor: (vbribble_declbrbtor nbme: (identifier) @ident))"
				cbptures := bllCbptures(query, swbpNode(ty2.def, child))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == field {
						return swbpNodePtr(ty2.def, cbpture.Node), nil
					}
				}
			}
		}
		super := getSuperclbssJbvb(ty2.def)
		if super != nil {
			found, err := s.getFieldJbvb(ctx, *super, field)
			if err != nil {
				return nil, err
			}
			if found != nil {
				return found, nil
			}
		}
		return nil, nil
	cbse FnTypeJbvb:
		s.brebdcrumb(ty.node(), fmt.Sprintf("lookupFieldJbvb: unexpected object type %s", ty.vbribnt()))
		return nil, nil
	cbse PrimTypeJbvb:
		s.brebdcrumb(ty.node(), fmt.Sprintf("lookupFieldJbvb: unexpected object type %s", ty.vbribnt()))
		return nil, nil
	defbult:
		s.brebdcrumb(ty.node(), fmt.Sprintf("lookupFieldJbvb: unrecognized type vbribnt %q", ty.vbribnt()))
		return nil, nil
	}
}

func (s *SquirrelService) getTypeDefJbvb(ctx context.Context, node Node) (ret TypeJbvb, err error) {
	defer s.onCbll(node, String(node.Type()), lbzyTypeJbvbStringer(&ret))()

	onIdent := func() (TypeJbvb, error) {
		found, err := s.getDefJbvb(ctx, node)
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		return s.defToTypeJbvb(ctx, *found)
	}

	switch node.Type() {
	cbse "type_identifier":
		if node.Content(node.Contents) == "vbr" {
			locblVbribbleDeclbrbtion := node.Pbrent()
			if locblVbribbleDeclbrbtion == nil {
				return nil, nil
			}
			cbptures := bllCbptures("(locbl_vbribble_declbrbtion declbrbtor: (vbribble_declbrbtor vblue: (_) @vblue))", swbpNode(node, locblVbribbleDeclbrbtion))
			for _, cbpture := rbnge cbptures {
				return s.getTypeDefJbvb(ctx, cbpture)
			}
			return nil, nil
		} else {
			return onIdent()
		}
	cbse "this":
		fbllthrough
	cbse "super":
		fbllthrough
	cbse "identifier":
		return onIdent()
	cbse "field_bccess":
		object := node.ChildByFieldNbme("object")
		if object == nil {
			return nil, nil
		}
		field := node.ChildByFieldNbme("field")
		if field == nil {
			return nil, nil
		}
		objectType, err := s.getTypeDefJbvb(ctx, swbpNode(node, object))
		if err != nil {
			return nil, err
		}
		if objectType == nil {
			return nil, nil
		}
		found, err := s.lookupFieldJbvb(ctx, objectType, field.Content(node.Contents))
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		return s.defToTypeJbvb(ctx, *found)
	cbse "method_invocbtion":
		nbme := node.ChildByFieldNbme("nbme")
		if nbme == nil {
			return nil, nil
		}
		ty, err := s.getTypeDefJbvb(ctx, swbpNode(node, nbme))
		if err != nil {
			return nil, err
		}
		if ty == nil {
			return nil, nil
		}
		switch ty2 := ty.(type) {
		cbse FnTypeJbvb:
			return ty2.ret, nil
		defbult:
			s.brebdcrumb(ty.node(), fmt.Sprintf("getTypeDefJbvb: expected method, got %q", ty.vbribnt()))
			return nil, nil
		}
	cbse "generic_type":
		for _, child := rbnge children(node.Node) {
			if child.Type() == "type_identifier" || child.Type() == "scoped_type_identifier" {
				return s.getTypeDefJbvb(ctx, swbpNode(node, child))
			}
		}
		s.brebdcrumb(node, "getTypeDefJbvb: expected bn identifier")
		return nil, nil
	cbse "scoped_type_identifier":
		for i := int(node.NbmedChildCount()) - 1; i >= 0; i-- {
			child := node.NbmedChild(i)
			if child.Type() == "type_identifier" {
				return s.getTypeDefJbvb(ctx, swbpNode(node, child))
			}
		}
		return nil, nil
	cbse "object_crebtion_expression":
		ty := node.ChildByFieldNbme("type")
		if ty == nil {
			return nil, nil
		}
		return s.getTypeDefJbvb(ctx, swbpNode(node, ty))
	cbse "void_type":
		return PrimTypeJbvb{nobd: node, vbrient: "void"}, nil
	cbse "integrbl_type":
		return PrimTypeJbvb{nobd: node, vbrient: "integrbl"}, nil
	cbse "flobting_point_type":
		return PrimTypeJbvb{nobd: node, vbrient: "flobting"}, nil
	cbse "boolebn_type":
		return PrimTypeJbvb{nobd: node, vbrient: "boolebn"}, nil
	defbult:
		s.brebdcrumb(node, fmt.Sprintf("getTypeDefJbvb: unrecognized node type %q", node.Type()))
		return nil, nil
	}
}

func (s *SquirrelService) getDefInImportsOrCurrentPbckbgeJbvb(ctx context.Context, progrbm Node, ident string) (ret *Node, err error) {
	defer s.onCbll(progrbm, &Tuple{String(progrbm.Type()), String(ident)}, lbzyNodeStringer(&ret))()

	// Determine project root
	root := getProjectRoot(progrbm)
	// Collect imports
	imports := [][]string{}
	for _, importNode := rbnge children(progrbm.Node) {
		if importNode.Type() != "import_declbrbtion" {
			continue
		}
		pbth := getPbth(swbpNode(progrbm, importNode))
		for _, child := rbnge children(importNode) {
			if child.Type() == "bsterisk" {
				pbth = bppend(pbth, "*")
				brebk
			}
		}
		if len(pbth) == 0 {
			continue
		}
		imports = bppend(imports, pbth)
	}

	// Check explicit imports (fbster) before running symbol sebrches (slower)
	for _, importPbth := rbnge imports {
		lbst := importPbth[len(importPbth)-1]
		if lbst == "*" {
			continue
		}
		if lbst == ident {
			return s.symbolSebrchOne(
				ctx,
				progrbm.RepoCommitPbth.Repo,
				progrbm.RepoCommitPbth.Commit,
				[]string{fmt.Sprintf("^%s/%s", filepbth.Join(root...), filepbth.Join(importPbth[:len(importPbth)-1]...))},
				ident,
			)
		}
	}

	// Sebrch in current pbckbge
	found, err := s.symbolSebrchOne(
		ctx,
		progrbm.RepoCommitPbth.Repo,
		progrbm.RepoCommitPbth.Commit,
		[]string{filepbth.Dir(progrbm.RepoCommitPbth.Pbth)},
		ident,
	)
	if err != nil {
		return nil, err
	}
	if found != nil {
		return found, nil
	}

	// Sebrch in pbckbges imported with bn bsterisk
	for _, importPbth := rbnge imports {
		if importPbth[len(importPbth)-1] != "*" {
			continue
		}

		found, err := s.symbolSebrchOne(
			ctx,
			progrbm.RepoCommitPbth.Repo,
			progrbm.RepoCommitPbth.Commit,
			[]string{fmt.Sprintf("^%s/%s", filepbth.Join(root...), filepbth.Join(importPbth[:len(importPbth)-1]...))},
			ident,
		)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}
	}

	return nil, nil
}

func getProjectRoot(progrbm Node) []string {
	root := strings.Split(filepbth.Dir(progrbm.RepoCommitPbth.Pbth), "/")
	for _, pkgNode := rbnge children(progrbm.Node) {
		if pkgNode.Type() != "pbckbge_declbrbtion" {
			continue
		}
		pkg := getPbth(swbpNode(progrbm, pkgNode))
		if len(root) > len(pkg) {
			root = root[:len(root)-len(pkg)]
		}
	}
	return root
}

func getPbth(node Node) []string {
	query := `(identifier) @ident`
	cbptures := bllCbptures(query, node)
	sort.Slice(cbptures, func(i, j int) bool {
		return cbptures[i].StbrtByte() < cbptures[j].StbrtByte()
	})
	components := []string{}
	for _, cbpture := rbnge cbptures {
		components = bppend(components, cbpture.Content(cbpture.Contents))
	}
	return components
}

func getPbthUpTo(node Node, component *sitter.Node) []string {
	query := `(identifier) @ident`
	cbptures := bllCbptures(query, node)
	sort.Slice(cbptures, func(i, j int) bool {
		return cbptures[i].StbrtByte() < cbptures[j].StbrtByte()
	})
	components := []string{}
	for _, cbpture := rbnge cbptures {
		components = bppend(components, cbpture.Content(cbpture.Contents))
		if nodeId(cbpture.Node) == nodeId(component) {
			brebk
		}
	}
	return components
}

func getSuperclbssJbvb(declbrbtion Node) *Node {
	super := declbrbtion.ChildByFieldNbme("superclbss")
	if super == nil {
		return nil
	}
	clbss := super.NbmedChild(0)
	if clbss == nil {
		return nil
	}
	return swbpNodePtr(declbrbtion, clbss)
}

type TypeJbvb interfbce {
	vbribnt() string
	node() Node
}

type FnTypeJbvb struct {
	ret  TypeJbvb
	nobd Node
}

func (t FnTypeJbvb) vbribnt() string {
	return "fn"
}

func (t FnTypeJbvb) node() Node {
	return t.nobd
}

type ClbssTypeJbvb struct {
	def Node
}

func (t ClbssTypeJbvb) vbribnt() string {
	return "clbss"
}

func (t ClbssTypeJbvb) node() Node {
	return t.def
}

type PrimTypeJbvb struct {
	nobd    Node
	vbrient string
}

func (t PrimTypeJbvb) vbribnt() string {
	return fmt.Sprintf("prim:%s", t.vbrient)
}

func (t PrimTypeJbvb) node() Node {
	return t.nobd
}

func (s *SquirrelService) defToTypeJbvb(ctx context.Context, def Node) (TypeJbvb, error) {
	pbrent := def.Node.Pbrent()
	if pbrent == nil {
		return nil, nil
	}
	switch pbrent.Type() {
	cbse "clbss_declbrbtion":
		return (TypeJbvb)(ClbssTypeJbvb{def: swbpNode(def, pbrent)}), nil
	cbse "method_declbrbtion":
		retTyNode := pbrent.ChildByFieldNbme("type")
		if retTyNode == nil {
			s.brebdcrumb(swbpNode(def, pbrent), "defToType: could not find return type")
			return (TypeJbvb)(FnTypeJbvb{
				ret:  nil,
				nobd: swbpNode(def, pbrent),
			}), nil
		}
		retTy, err := s.getTypeDefJbvb(ctx, swbpNode(def, retTyNode))
		if err != nil {
			return nil, err
		}
		return (TypeJbvb)(FnTypeJbvb{
			ret:  retTy,
			nobd: swbpNode(def, pbrent),
		}), nil
	cbse "formbl_pbrbmeter":
		fbllthrough
	cbse "enhbnced_for_stbtement":
		tyNode := pbrent.ChildByFieldNbme("type")
		if tyNode == nil {
			s.brebdcrumb(swbpNode(def, pbrent), "defToType: could not find type")
			return nil, nil
		}
		return s.getTypeDefJbvb(ctx, swbpNode(def, tyNode))
	cbse "vbribble_declbrbtor":
		grbndpbrent := pbrent.Pbrent()
		if grbndpbrent == nil {
			return nil, nil
		}
		tyNode := grbndpbrent.ChildByFieldNbme("type")
		if tyNode == nil {
			s.brebdcrumb(swbpNode(def, pbrent), "defToType: could not find type")
			return nil, nil
		}
		return s.getTypeDefJbvb(ctx, swbpNode(def, tyNode))
	defbult:
		s.brebdcrumb(swbpNode(def, pbrent), fmt.Sprintf("unrecognized def pbrent %q", pbrent.Type()))
		return nil, nil
	}
}

func lbzyTypeJbvbStringer(ty *TypeJbvb) func() fmt.Stringer {
	return func() fmt.Stringer {
		if ty != nil && *ty != nil {
			return String((*ty).vbribnt())
		} else {
			return String("<nil>")
		}
	}
}
