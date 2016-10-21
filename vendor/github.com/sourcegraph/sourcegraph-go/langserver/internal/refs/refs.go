package refs

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

type Def struct {
	// ImportPath is the import path at which the definition is located.
	ImportPath string

	// PackageName is the name the package is imported under.
	PackageName string

	// Path is the path of the definition (e.g. "Router Route" for a
	// method named "Route" with receiver type "Router").
	Path string
}

// Ref represents a reference to a definition.
type Ref struct {
	// Def is the definition being referenced.
	Def Def

	// Position is the position of the reference.
	Position token.Position
}

func (r *Ref) String() string {
	return fmt.Sprintf("&Ref{Def.ImportPath: %q, Def.PackageName: %q, Def.Path: %q, Position: \"%s:%d:%d:%d\"}", r.Def.ImportPath, r.Def.PackageName, r.Def.Path, r.Position.Filename, r.Position.Line, r.Position.Column, r.Position.Offset)
}

type Config struct {
	FileSet  *token.FileSet
	Pkg      *types.Package
	PkgFiles []*ast.File
	Info     *types.Info
}

func (c *Config) Refs(emit func(*Ref)) error {
	ref := func(rootFile *ast.File, pos token.Pos) error {
		d, err := c.defInfo(c.Pkg, c.Info, rootFile, pos)
		if err == errReceiverNotTopLevelNamedType {
			return nil
		}
		if err != nil {
			return err
		}
		for _, f := range strings.Fields(d.Path) {
			if !ast.IsExported(f) {
				return nil
			}
		}
		emit(&Ref{
			Def:      *d,
			Position: c.FileSet.Position(pos),
		})
		return nil
	}

	var firstErr error
	for _, file := range c.PkgFiles {
		ast.Inspect(file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.ImportSpec:
				if err := ref(file, n.Pos()); err != nil {
					firstErr = err
					return false
				}

			case *ast.SelectorExpr:
				if err := ref(file, n.Sel.Pos()); err != nil {
					firstErr = err
					return false
				}

			case *ast.CompositeLit:
				for _, e := range n.Elts {
					kv, ok := e.(*ast.KeyValueExpr)
					if !ok {
						continue
					}

					// Keys themselves can reference external types, when the key
					// is an embedded struct field.
					ident, ok := kv.Key.(*ast.Ident)
					if !ok {
						continue
					}
					if err := ref(file, ident.Pos()); err != nil {
						// Ignore "not a package-level definition errors",
						// since these fall into an edge case (do not represent
						// real errors for us).
						if _, ok := err.(*notPackageLevelDef); !ok {
							firstErr = err
							return false
						}
					}
				}
			}
			return true
		})
		if firstErr != nil {
			return firstErr
		}
	}
	return nil
}

var errReceiverNotTopLevelNamedType = errors.New("receiver is not a top-level named type")

type notPackageLevelDef struct {
	ident *ast.Ident
	obj   types.Object
	t     types.Type
}

func (e *notPackageLevelDef) Error() string {
	return fmt.Sprintf("not a package-level definition (ident: %v, object: %v) and unable to follow type (type: %v)", e.ident, e.obj, e.t)
}

func (c *Config) defInfo(pkg *types.Package, info *types.Info, rootFile *ast.File, pos token.Pos) (*Def, error) {
	nodes, _ := astutil.PathEnclosingInterval(rootFile, pos, pos)

	// Import statements.
	if len(nodes) > 2 {
		if im, ok := nodes[1].(*ast.ImportSpec); ok {
			pkgPath, err := strconv.Unquote(im.Path.Value)
			if err != nil {
				return nil, err
			}

			var impPkg *types.Package
			for _, p := range pkg.Imports() {
				path := p.Path()
				if strings.Contains(path, "vendor/") {
					path = path[strings.LastIndex(path, "vendor/")+len("vendor/"):]
				}
				if path == pkgPath {
					impPkg = p
					break
				}
			}
			if impPkg == nil {
				return nil, fmt.Errorf("could not find package %q in imports", pkgPath)
			}
			return &Def{ImportPath: impPkg.Path(), PackageName: impPkg.Name()}, nil
		}
	}

	var identX *ast.Ident
	var selX *ast.SelectorExpr
	selX, ok := nodes[0].(*ast.SelectorExpr)
	if ok {
		identX = selX.Sel
	} else {
		identX, ok = nodes[0].(*ast.Ident)
		if !ok {
			return nil, errors.New("no identifier found")
		}
		if len(nodes) > 1 {
			selX, _ = nodes[1].(*ast.SelectorExpr)
		}
	}

	if obj := info.Defs[identX]; obj != nil {
		switch t := obj.Type().(type) {
		case *types.Signature:
			if t.Recv() == nil {
				// Top-level func.
				return objectString(obj), nil
			}
			// Method or interface method.
			return &Def{
				ImportPath:  obj.Pkg().Path(),
				PackageName: obj.Pkg().Name(),
				Path:        fmt.Sprintf("%v %v", dereferenceType(t.Recv().Type()).(*types.Named).Obj().Name(), identX.Name),
			}, nil
		}

		if obj.Parent() == pkg.Scope() {
			// Top-level package def.
			return objectString(obj), nil
		}

		// Struct field.
		if _, ok := nodes[1].(*ast.Field); ok {
			if typ, ok := nodes[4].(*ast.TypeSpec); ok {
				return &Def{
					ImportPath:  obj.Pkg().Path(),
					PackageName: obj.Pkg().Name(),
					Path:        fmt.Sprintf("%v %v", typ.Name.Name, obj.Name()),
				}, nil
			}
		}

		if pkg, pkgName, name, ok := typeName(dereferenceType(obj.Type())); ok {
			return &Def{ImportPath: pkg, PackageName: pkgName, Path: name}, nil
		}
		return nil, fmt.Errorf("unable to identify def (ident: %v, object: %v)", identX, obj)
	}

	obj := info.Uses[identX]
	if obj == nil {
		return nil, fmt.Errorf("no type information for identifier %q at %d", identX.Name, pos)
	}

	if obj, ok := obj.(*types.Var); ok && obj.IsField() {
		// Struct literal
		if lit, ok := nodes[2].(*ast.CompositeLit); ok {
			if parent, ok := lit.Type.(*ast.SelectorExpr); ok {
				return &Def{
					ImportPath:  obj.Pkg().Path(),
					PackageName: obj.Pkg().Name(),
					Path:        fmt.Sprintf("%v %v", parent.Sel, obj.Id()),
				}, nil
			} else if parent, ok := lit.Type.(*ast.Ident); ok {
				return &Def{
					ImportPath:  obj.Pkg().Path(),
					PackageName: obj.Pkg().Name(),
					Path:        fmt.Sprintf("%v %v", parent, obj.Id()),
				}, nil
			}
		}
	}

	if pkgName, ok := obj.(*types.PkgName); ok {
		return &Def{ImportPath: pkgName.Imported().Path(), PackageName: pkgName.Imported().Name()}, nil
	} else if selX == nil {
		if pkg.Scope().Lookup(identX.Name) == obj {
			return objectString(obj), nil
		} else if types.Universe.Lookup(identX.Name) == obj {
			return &Def{ImportPath: "builtin", PackageName: "builtin", Path: obj.Name()}, nil
		}
		t := dereferenceType(obj.Type())
		if pkg, pkgName, name, ok := typeName(t); ok {
			return &Def{ImportPath: pkg, PackageName: pkgName, Path: name}, nil
		}
		return nil, &notPackageLevelDef{
			ident: identX,
			obj:   obj,
			t:     t,
		}
	} else if sel, ok := info.Selections[selX]; ok {
		recv, ok := dereferenceType(deepRecvType(sel)).(*types.Named)
		if !ok || recv == nil || recv.Obj() == nil || recv.Obj().Pkg() == nil || recv.Obj().Pkg().Scope().Lookup(recv.Obj().Name()) != recv.Obj() {
			return nil, errReceiverNotTopLevelNamedType
		}

		field, _, _ := types.LookupFieldOrMethod(sel.Recv(), true, pkg, identX.Name)
		if field == nil {
			// field invoked, but object is selected
			t := dereferenceType(obj.Type())
			if pkg, pkgName, name, ok := typeName(t); ok {
				return &Def{ImportPath: pkg, PackageName: pkgName, Path: name}, nil
			}
			return nil, fmt.Errorf("method or field not found")
		}

		d := objectString(recv.Obj())
		d.Path = fmt.Sprintf("%v %v", d.Path, identX.Name)
		return d, nil
	}
	// Qualified reference (to another package's top-level
	// definition).
	if obj := info.Uses[selX.Sel]; obj != nil {
		return objectString(obj), nil
	}
	return nil, errors.New("no selector type")
}

// deepRecvType gets the embedded struct's name that the method or
// field is actually defined on, not just the original/outer recv
// type.
func deepRecvType(sel *types.Selection) types.Type {
	var offset int
	offset = 1
	if sel.Kind() == types.MethodVal || sel.Kind() == types.MethodExpr {
		offset = 0
	}

	typ := sel.Recv()
	idx := sel.Index()
	for k, i := range idx[:len(idx)-offset] {
		final := k == len(idx)-offset-1
		t := getMethod(typ, i, final, sel.Kind() != types.FieldVal)
		if t == nil {
			log.Printf("failed to get method/field at index %v on recv %s", idx, typ)
			return nil
		}
		typ = t.Type()
	}
	return typ
}

func dereferenceType(typ types.Type) types.Type {
	if typ, ok := typ.(*types.Pointer); ok {
		return typ.Elem()
	}
	return typ
}

func typeName(typ types.Type) (pkg, pkgName, name string, ok bool) {
	switch typ := typ.(type) {
	case *types.Named:
		p := typ.Obj().Pkg()
		if p == nil {
			return
		}
		return p.Path(), p.Name(), typ.Obj().Name(), true
	case *types.Basic:
		return "builtin", "builtin", typ.Name(), true
	}
	return
}

func getMethod(typ types.Type, idx int, final bool, method bool) (obj types.Object) {
	switch obj := typ.(type) {
	case *types.Pointer:
		return getMethod(obj.Elem(), idx, final, method)

	case *types.Named:
		if final && method {
			switch obj2 := dereferenceType(obj.Underlying()).(type) {
			case *types.Interface:
				recvObj := obj2.Method(idx).Type().(*types.Signature).Recv()
				if recvObj.Type() == obj.Underlying() {
					return obj.Obj()
				}
				return recvObj
			}
			return obj.Method(idx).Type().(*types.Signature).Recv()
		}
		return getMethod(obj.Underlying(), idx, final, method)

	case *types.Struct:
		return obj.Field(idx)

	case *types.Interface:
		// Our index is among all methods, but we want to get the
		// interface that defines the method at our index.
		return obj.Method(idx).Type().(*types.Signature).Recv()
	}
	return nil
}

func objectString(obj types.Object) *Def {
	if obj.Pkg() != nil {
		return &Def{ImportPath: obj.Pkg().Path(), PackageName: obj.Pkg().Name(), Path: obj.Name()}
	}
	return &Def{Path: obj.Name()}
}
