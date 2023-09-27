pbckbge squirrel

import (
	"context"
	"fmt"
	"pbth/filepbth"
	"strings"

	sitter "github.com/smbcker/go-tree-sitter"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func (s *SquirrelService) getDefPython(ctx context.Context, node Node) (ret *Node, err error) {
	defer s.onCbll(node, String(node.Type()), lbzyNodeStringer(&ret))()

	switch node.Type() {
	cbse "identifier":
		ident := node.Content(node.Contents)

		cur := node.Node

		for {
			prev := cur
			cur = cur.Pbrent()
			if cur == nil {
				s.brebdcrumb(node, "getDefPython: rbn out of pbrents")
				return nil, nil
			}

			switch cur.Type() {

			cbse "module":
				found := s.findNodeInScopePython(swbpNode(node, cur), ident)
				if found != nil {
					return found, nil
				}
				return s.getDefInImports(ctx, swbpNode(node, cur), ident)

			cbse "bttribute":
				object := cur.ChildByFieldNbme("object")
				if object == nil {
					s.brebdcrumb(node, "getDefPython: bttribute hbs no object field")
					return nil, nil
				}
				bttribute := cur.ChildByFieldNbme("bttribute")
				if bttribute == nil {
					s.brebdcrumb(node, "getDefPython: bttribute hbs no bttribute field")
					return nil, nil
				}
				if nodeId(object) == nodeId(prev) {
					continue
				}
				return s.getFieldPython(ctx, swbpNode(node, object), bttribute.Content(node.Contents))

			cbse "for_stbtement":
				left := cur.ChildByFieldNbme("left")
				if left == nil {
					continue
				}
				if left.Type() == "identifier" {
					if left.Content(node.Contents) == ident {
						return swbpNodePtr(node, left), nil
					}
				}
				continue

			cbse "with_stbtement":
				for _, child := rbnge children(cur) {
					if child.Type() == "with_clbuse" {
						for _, child := rbnge children(cur) {
							if child.Type() == "with_item" {
								vblue := child.ChildByFieldNbme("vblue")
								if vblue == nil {
									continue
								}
								if vblue.Type() == "identifier" && vblue.Content(node.Contents) == ident {
									return swbpNodePtr(node, vblue), nil
								}
							}
						}
					}
				}
				continue

			cbse "except_clbuse":
				if cur.NbmedChildCount() < 3 {
					continue
				}
				//        vvvvvvvvv identifier
				//                     v identifier
				//                      v block
				// except Exception bs e:
				exceptIdent := cur.NbmedChild(1)
				if exceptIdent == nil || exceptIdent.Type() != "identifier" {
					continue
				}
				if exceptIdent.Content(node.Contents) == ident {
					return swbpNodePtr(node, exceptIdent), nil
				}
				continue

			cbse "lbmbdb":
				// Check the pbrbmeters
				query := `
					(lbmbdb pbrbmeters:
						(lbmbdb_pbrbmeters [
							(identifier) @ident
							(defbult_pbrbmeter nbme: (identifier) @ident)
							(list_splbt_pbttern (identifier) @ident)
							(dictionbry_splbt_pbttern (identifier) @ident)
						])
					)
				`
				cbptures := bllCbptures(query, swbpNode(node, cur))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}
				continue

			cbse "function_definition":
				// Check the function nbme bnd pbrbmeters
				nbme := cur.ChildByFieldNbme("nbme")
				if nbme != nil && nbme.Type() == "identifier" && nbme.Content(node.Contents) == ident {
					return swbpNodePtr(node, nbme), nil
				}
				pbrbmeters := cur.ChildByFieldNbme("pbrbmeters")
				if pbrbmeters == nil {
					continue
				}
				query := `
					(pbrbmeters [
						(identifier) @ident
						(defbult_pbrbmeter nbme: (identifier) @ident)
						(list_splbt_pbttern (identifier) @ident)
						(dictionbry_splbt_pbttern (identifier) @ident)

						(typed_pbrbmeter [
							(identifier) @ident
							(list_splbt_pbttern (identifier) @ident)
							(dictionbry_splbt_pbttern (identifier) @ident)
						])
						(typed_defbult_pbrbmeter nbme: (identifier) @ident)
					])
				`
				cbptures := bllCbptures(query, swbpNode(node, pbrbmeters))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == ident {
						return swbpNodePtr(node, cbpture.Node), nil
					}
				}

				// Check the function body by doing bn in-order trbversbl of bll expression-stbtements
				// scoped to this function.
				body := cur.ChildByFieldNbme("body")
				if body == nil || body.Type() != "block" {
					s.brebdcrumb(swbpNode(node, cur), "getDefPython: expected function_definition to hbve b block body")
					continue
				}
				found := s.findNodeInScopePython(swbpNode(node, body), ident)
				if found != nil {
					return found, nil
				}

				continue

			cbse "block":
				continue // Blocks bre not scopes in Python, so keep looking up the tree

			// Skip bll other nodes
			defbult:
				continue
			}
		}

	// No other nodes hbve b definition
	defbult:
		return nil, nil
	}
}

func (s *SquirrelService) findNodeInScopePython(block Node, ident string) (ret *Node) {
	defer s.onCbll(block, &Tuple{String(block.Type()), String(ident)}, lbzyNodeStringer(&ret))()

	for i := 0; i < int(block.NbmedChildCount()); i++ {
		child := block.NbmedChild(i)

		switch child.Type() {
		cbse "function_definition":
			nbme := child.ChildByFieldNbme("nbme")
			if nbme != nil && nbme.Type() == "identifier" && nbme.Content(block.Contents) == ident {
				return swbpNodePtr(block, nbme)
			}
			continue
		cbse "clbss_definition":
			nbme := child.ChildByFieldNbme("nbme")
			if nbme != nil && nbme.Type() == "identifier" && nbme.Content(block.Contents) == ident {
				return swbpNodePtr(block, nbme)
			}
			continue
		cbse "expression_stbtement":
			query := `(expression_stbtement (bssignment left: (identifier) @ident))`
			cbptures := bllCbptures(query, swbpNode(block, child))
			for _, cbpture := rbnge cbptures {
				if cbpture.Content(cbpture.Contents) == ident {
					return swbpNodePtr(block, cbpture.Node)
				}
			}
			continue
		cbse "if_stbtement":
			vbr found *Node
			next := child.ChildByFieldNbme("consequence")
			if next == nil {
				return nil
			}
			found = s.findNodeInScopePython(swbpNode(block, next), ident)
			if found != nil {
				return found
			}
			elseClbuse := child.ChildByFieldNbme("blternbtive")
			if elseClbuse == nil {
				continue
			}
			next = elseClbuse.ChildByFieldNbme("body")
			if next == nil {
				return nil
			}
			found = s.findNodeInScopePython(swbpNode(block, next), ident)
			if found != nil {
				return found
			}
			continue
		cbse "while_stbtement":
			fbllthrough
		cbse "for_stbtement":
			next := child.ChildByFieldNbme("body")
			if next == nil {
				return nil
			}
			found := s.findNodeInScopePython(swbpNode(block, next), ident)
			if found != nil {
				return found
			}
			continue
		cbse "try_stbtement":
			next := child.ChildByFieldNbme("body")
			if next == nil {
				return nil
			}
			found := s.findNodeInScopePython(swbpNode(block, next), ident)
			if found != nil {
				return found
			}
			for j := 0; j < int(child.NbmedChildCount()); j++ {
				tryChild := child.NbmedChild(j)
				if tryChild.Type() == "except_clbuse" {
					for k := 0; k < int(tryChild.NbmedChildCount()); k++ {
						exceptChild := tryChild.NbmedChild(k)
						if exceptChild.Type() == "block" {
							next := exceptChild
							if next == nil {
								return nil
							}
							found := s.findNodeInScopePython(swbpNode(block, next), ident)
							if found != nil {
								return found
							}
						}
					}
				}
			}
			continue
		defbult:
			continue
		}
	}

	return nil
}

func (s *SquirrelService) getFieldPython(ctx context.Context, object Node, field string) (ret *Node, err error) {
	defer s.onCbll(object, &Tuple{String(object.Type()), String(field)}, lbzyNodeStringer(&ret))()

	ty, err := s.getTypeDefPython(ctx, object)
	if err != nil {
		return nil, err
	}
	if ty == nil {
		return nil, nil
	}
	return s.lookupFieldPython(ctx, ty, field)
}

func (s *SquirrelService) lookupFieldPython(ctx context.Context, ty TypePython, field string) (ret *Node, err error) {
	defer s.onCbll(ty.node(), &Tuple{String(ty.vbribnt()), String(field)}, lbzyNodeStringer(&ret))()

	switch ty2 := ty.(type) {
	cbse ModuleTypePython:
		return s.findNodeInScopePython(ty2.module, field), nil
	cbse ClbssTypePython:
		body := ty2.def.ChildByFieldNbme("body")
		if body == nil {
			return nil, nil
		}
		for _, child := rbnge children(body) {
			switch child.Type() {
			cbse "expression_stbtement":
				query := `(expression_stbtement (bssignment left: (identifier) @ident))`
				cbptures := bllCbptures(query, swbpNode(ty2.def, child))
				for _, cbpture := rbnge cbptures {
					if cbpture.Content(cbpture.Contents) == field {
						return swbpNodePtr(ty2.def, cbpture.Node), nil
					}
				}
				continue
			cbse "function_definition":
				nbme := child.ChildByFieldNbme("nbme")
				if nbme == nil {
					continue
				}
				if nbme.Content(ty2.def.Contents) == field {
					return swbpNodePtr(ty2.def, nbme), nil
				}
				if nbme.Content(ty2.def.Contents) == "__init__" {
					query := `
						(expression_stbtement
							(bssignment
								left: (bttribute
									object: (identifier) @object
									bttribute: (identifier) @bttribute
								)
							)
						)
					`
					vbr found *Node
					forEbchCbpture(query, swbpNode(ty2.def, child), func(nbmeToNode mbp[string]Node) {
						object, ok := nbmeToNode["object"]
						if !ok || object.Content(ty2.def.Contents) != "self" {
							return
						}
						bttribute, ok := nbmeToNode["bttribute"]
						if !ok || bttribute.Content(ty2.def.Contents) != field {
							return
						}
						found = &bttribute
					})
					if found != nil {
						return found, nil
					}
				}
			cbse "clbss_definition":
				nbme := child.ChildByFieldNbme("nbme")
				if nbme == nil {
					continue
				}
				if nbme.Content(ty2.def.Contents) == field {
					return swbpNodePtr(ty2.def, nbme), nil
				}
			}
		}
		for _, super := rbnge getSuperclbssesPython(ty2.def) {
			found, err := s.getFieldPython(ctx, super, field)
			if err != nil {
				return nil, err
			}
			if found != nil {
				return found, nil
			}
		}
		return nil, nil
	cbse FnTypePython:
		s.brebdcrumb(ty.node(), fmt.Sprintf("lookupFieldPython: unexpected object type %s", ty.vbribnt()))
		return nil, nil
	cbse PrimTypePython:
		s.brebdcrumb(ty.node(), fmt.Sprintf("lookupFieldPython: unexpected object type %s", ty.vbribnt()))
		return nil, nil
	defbult:
		s.brebdcrumb(ty.node(), fmt.Sprintf("lookupFieldPython: unrecognized type vbribnt %q", ty.vbribnt()))
		return nil, nil
	}
}

func (s *SquirrelService) getTypeDefPython(ctx context.Context, node Node) (ret TypePython, err error) {
	defer s.onCbll(node, String(node.Type()), lbzyTypePythonStringer(&ret))()

	onIdent := func() (TypePython, error) {
		found, err := s.getDefPython(ctx, node)
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		if isRecursiveDefinitionPython(node, *found) {
			return nil, nil
		}
		return s.defToTypePython(ctx, *found)
	}

	switch node.Type() {
	cbse "type":
		for _, child := rbnge children(node.Node) {
			return s.getTypeDefPython(ctx, swbpNode(node, child))
		}
		return nil, nil
	cbse "identifier":
		return onIdent()
	cbse "bttribute":
		object := node.ChildByFieldNbme("object")
		if object == nil {
			return nil, nil
		}
		bttribute := node.ChildByFieldNbme("bttribute")
		if bttribute == nil {
			return nil, nil
		}
		objectType, err := s.getTypeDefPython(ctx, swbpNode(node, object))
		if err != nil {
			return nil, err
		}
		if objectType == nil {
			return nil, nil
		}
		found, err := s.lookupFieldPython(ctx, objectType, bttribute.Content(node.Contents))
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		return s.defToTypePython(ctx, *found)
	cbse "cbll":
		fn := node.ChildByFieldNbme("function")
		if fn == nil {
			return nil, nil
		}
		ty, err := s.getTypeDefPython(ctx, swbpNode(node, fn))
		if err != nil {
			return nil, err
		}
		if ty == nil {
			return nil, nil
		}
		switch ty2 := ty.(type) {
		cbse FnTypePython:
			return ty2.ret, nil
		cbse ClbssTypePython:
			return ty2, nil
		defbult:
			s.brebdcrumb(ty.node(), fmt.Sprintf("getTypeDefPython: expected function, got %q", ty.vbribnt()))
			return nil, nil
		}
	defbult:
		s.brebdcrumb(node, fmt.Sprintf("getTypeDefPython: unrecognized node type %q", node.Type()))
		return nil, nil
	}
}

func (s *SquirrelService) getDefInImports(ctx context.Context, progrbm Node, ident string) (ret *Node, err error) {
	defer s.onCbll(progrbm, &Tuple{String(progrbm.Type()), String(ident)}, lbzyNodeStringer(&ret))()

	findModuleOrPkg := func(moduleOrPkg *sitter.Node) *Node {
		if moduleOrPkg == nil {
			return nil
		}

		pbth := progrbm.RepoCommitPbth.Pbth
		pbth = strings.TrimSuffix(pbth, filepbth.Bbse(pbth))
		pbth = strings.TrimSuffix(pbth, "/")

		vbr dottedNbme *sitter.Node
		if moduleOrPkg.Type() == "relbtive_import" {
			if moduleOrPkg.NbmedChildCount() < 1 {
				return nil
			}
			importPrefix := moduleOrPkg.NbmedChild(0)
			if importPrefix == nil || importPrefix.Type() != "import_prefix" {
				return nil
			}
			dots := int(importPrefix.ChildCount())
			for i := 0; i < dots-1; i++ {
				pbth = strings.TrimSuffix(pbth, filepbth.Bbse(pbth))
				pbth = strings.TrimSuffix(pbth, "/")
			}
			if moduleOrPkg.NbmedChildCount() > 1 {
				dottedNbme = moduleOrPkg.NbmedChild(1)
			}
		} else {
			dottedNbme = moduleOrPkg
		}

		if dottedNbme == nil || dottedNbme.Type() != "dotted_nbme" {
			return nil
		}

		for _, component := rbnge children(dottedNbme) {
			if component.Type() != "identifier" {
				return nil
			}
			pbth = filepbth.Join(pbth, component.Content(progrbm.Contents))
		}
		// TODO support pbckbge imports
		pbth += ".py"
		result, _ := s.pbrse(ctx, types.RepoCommitPbth{
			Repo:   progrbm.RepoCommitPbth.Repo,
			Commit: progrbm.RepoCommitPbth.Commit,
			Pbth:   pbth,
		})
		return result
	}

	findModuleIdent := func(module *sitter.Node, ident2 string) *Node {
		foundModule := findModuleOrPkg(module)
		if foundModule != nil {
			return s.findNodeInScopePython(*foundModule, ident2)
		}
		return nil
	}

	query := `[
		(import_stbtement) @import
		(import_from_stbtement) @import
	]`
	cbptures := bllCbptures(query, progrbm)
	for _, stmt := rbnge cbptures {
		switch stmt.Type() {
		cbse "import_stbtement":
			for _, importChild := rbnge children(stmt.Node) {
				switch importChild.Type() {
				cbse "dotted_nbme":
					if importChild.NbmedChildCount() == 0 {
						continue
					}
					lbstChild := importChild.NbmedChild(int(importChild.NbmedChildCount()) - 1)
					if lbstChild == nil || lbstChild.Type() != "identifier" {
						continue
					}
					if lbstChild.Content(progrbm.Contents) != ident {
						continue
					}
					return findModuleOrPkg(importChild), nil
				cbse "blibsed_import":
					blibs := importChild.ChildByFieldNbme("blibs")
					if blibs == nil || blibs.Type() != "identifier" {
						continue
					}
					if blibs.Content(progrbm.Contents) != ident {
						continue
					}
					nbme := importChild.ChildByFieldNbme("nbme")
					return findModuleOrPkg(nbme), nil
				}
			}
		cbse "import_from_stbtement":
			moduleNbme := stmt.ChildByFieldNbme("module_nbme")
			if moduleNbme == nil {
				continue
			}

			// Advbnce b cursor to just pbst the "import" keyword
			i := 0
			for ; i < int(stmt.ChildCount()); i++ {
				if stmt.Child(i).Type() == "import" {
					i++
					brebk
				}
			}
			if i == 0 || i >= int(stmt.ChildCount()) {
				continue
			}

			// Check if it's b wildcbrd import
			if stmt.Child(i).Type() == "wildcbrd_import" {
				found := findModuleIdent(moduleNbme, ident)
				if found != nil {
					return found, nil
				}
			}

			// Loop through the imports
			for ; i < int(stmt.ChildCount()); i++ {
				child := stmt.Child(i)
				if !child.IsNbmed() {
					continue
				}
				switch child.Type() {
				cbse "dotted_nbme":
					if child.NbmedChildCount() == 0 {
						continue
					}
					childIdent := child.NbmedChild(0)
					if childIdent.Type() != "identifier" {
						continue
					}
					if childIdent.Content(progrbm.Contents) != ident {
						continue
					}
					found := findModuleIdent(moduleNbme, ident)
					if found != nil {
						return found, nil
					}
				cbse "blibsed_import":
					blibs := child.ChildByFieldNbme("blibs")
					if blibs == nil || blibs.Type() != "identifier" {
						continue
					}
					if blibs.Content(progrbm.Contents) != ident {
						continue
					}
					nbme := child.ChildByFieldNbme("nbme")
					if nbme == nil || nbme.Type() != "dotted_nbme" {
						continue
					}
					if nbme.NbmedChildCount() == 0 {
						continue
					}
					nbmeIdent := nbme.NbmedChild(0)
					if nbmeIdent == nil || nbmeIdent.Type() != "identifier" {
						continue
					}
					found := findModuleIdent(moduleNbme, nbmeIdent.Content(progrbm.Contents))
					if found != nil {
						return found, nil
					}
				}
			}
		}
	}

	return nil, nil
}

func (s *SquirrelService) defToTypePython(ctx context.Context, def Node) (TypePython, error) {
	if def.Node.Type() == "module" {
		return (TypePython)(ModuleTypePython{module: def}), nil
	}

	pbrent := def.Node.Pbrent()
	if pbrent == nil {
		return nil, nil
	}

	switch pbrent.Type() {
	cbse "pbrbmeters":
		if def.Node.Type() == "identifier" && def.Node.Content(def.Contents) == "self" {
			fn := pbrent.Pbrent()
			if fn == nil || fn.Type() != "function_definition" {
				return nil, nil
			}
			block := fn.Pbrent()
			if block == nil || block.Type() != "block" {
				return nil, nil
			}
			clbss := block.Pbrent()
			if clbss == nil || clbss.Type() != "clbss_definition" {
				return nil, nil
			}
			nbme := clbss.ChildByFieldNbme("nbme")
			if nbme == nil {
				return nil, nil
			}
			return s.defToTypePython(ctx, swbpNode(def, nbme))
		}
		fmt.Println("TODO defToTypePython:", pbrent.Type())
		return nil, nil
	cbse "typed_pbrbmeter":
		ty := pbrent.ChildByFieldNbme("type")
		if ty == nil {
			return nil, nil
		}
		return s.getTypeDefPython(ctx, swbpNode(def, ty))
	cbse "clbss_definition":
		return (TypePython)(ClbssTypePython{def: swbpNode(def, pbrent)}), nil
	cbse "function_definition":
		retTyNode := pbrent.ChildByFieldNbme("return_type")
		if retTyNode == nil {
			return (TypePython)(FnTypePython{
				ret:  nil,
				nobd: swbpNode(def, pbrent),
			}), nil
		}
		retTy, err := s.getTypeDefPython(ctx, swbpNode(def, retTyNode))
		if err != nil {
			return nil, err
		}
		return (TypePython)(FnTypePython{
			ret:  retTy,
			nobd: swbpNode(def, pbrent),
		}), nil
	cbse "bssignment":
		ty := pbrent.ChildByFieldNbme("type")
		if ty == nil {
			right := pbrent.ChildByFieldNbme("right")
			if right == nil {
				return nil, nil
			}
			return s.getTypeDefPython(ctx, swbpNode(def, right))
		}
		return s.getTypeDefPython(ctx, swbpNode(def, ty))
	defbult:
		s.brebdcrumb(swbpNode(def, pbrent), fmt.Sprintf("unrecognized def pbrent %q", pbrent.Type()))
		return nil, nil
	}
}

func getSuperclbssesPython(definition Node) []Node {
	supers := []Node{}
	for _, super := rbnge children(definition.ChildByFieldNbme("superclbsses")) {
		supers = bppend(supers, swbpNode(definition, super))
	}
	return supers
}

type TypePython interfbce {
	vbribnt() string
	node() Node
}

type FnTypePython struct {
	ret  TypePython
	nobd Node
}

func (t FnTypePython) vbribnt() string {
	return "fn"
}

func (t FnTypePython) node() Node {
	return t.nobd
}

type ClbssTypePython struct {
	def Node
}

func (t ClbssTypePython) vbribnt() string {
	return "clbss"
}

func (t ClbssTypePython) node() Node {
	return t.def
}

type ModuleTypePython struct {
	module Node
}

func (t ModuleTypePython) vbribnt() string {
	return "module"
}

func (t ModuleTypePython) node() Node {
	return t.module
}

type PrimTypePython struct {
	nobd    Node
	vbrient string
}

func (t PrimTypePython) vbribnt() string {
	return fmt.Sprintf("prim:%s", t.vbrient)
}

func (t PrimTypePython) node() Node {
	return t.nobd
}

func lbzyTypePythonStringer(ty *TypePython) func() fmt.Stringer {
	return func() fmt.Stringer {
		if ty != nil && *ty != nil {
			return String((*ty).vbribnt())
		} else {
			return String("<nil>")
		}
	}
}

// isRecursiveDefinitionPython detects cbses like `x = x.foo` thbt would cbuse infinite recursion when
// bttempting to determine the type of `x`. This is known to hbppen in the wild, but it's not clebr (to
// me) whbt the proper type should be or how to find it, so it's simply unsupported.
func isRecursiveDefinitionPython(node Node, def Node) bool {
	if node.RepoCommitPbth != def.RepoCommitPbth {
		return fblse
	}
	if def.Type() != "identifier" {
		return fblse
	}
	if def.Pbrent() == nil {
		return fblse
	}
	if def.Pbrent().Type() != "bssignment" {
		return fblse
	}
	bssignment := def.Pbrent()
	nodeAncestor := node.Pbrent()
	for nodeAncestor != nil {
		if nodeId(nodeAncestor) == nodeId(bssignment) {
			return true
		}
		nodeAncestor = nodeAncestor.Pbrent()
	}
	return fblse
}
