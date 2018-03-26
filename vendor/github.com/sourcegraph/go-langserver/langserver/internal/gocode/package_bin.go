package gocode

import (
	"encoding/binary"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

//-------------------------------------------------------------------------
// gc_bin_parser
//
// The following part of the code may contain portions of the code from the Go
// standard library, which tells me to retain their copyright notice:
//
// Copyright (c) 2012 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//-------------------------------------------------------------------------

type gc_bin_parser struct {
	data    []byte
	buf     []byte // for reading strings
	version int    // export format version

	// object lists
	strList       []string   // in order of appearance
	pathList      []string   // in order of appearance
	pkgList       []string   // in order of appearance
	typList       []ast.Expr // in order of appearance
	callback      func(pkg string, decl ast.Decl)
	pfc           *package_file_cache
	trackAllTypes bool

	// position encoding
	posInfoFormat bool
	prevFile      string
	prevLine      int

	// debugging support
	debugFormat bool
	read        int // bytes read

}

func (p *gc_bin_parser) init(data []byte, pfc *package_file_cache) {
	p.data = data
	p.version = -1            // unknown version
	p.strList = []string{""}  // empty string is mapped to 0
	p.pathList = []string{""} // empty string is mapped to 0
	p.pfc = pfc
}

func (p *gc_bin_parser) parse_export(callback func(string, ast.Decl)) {
	p.callback = callback

	// read version info
	var versionstr string
	if b := p.rawByte(); b == 'c' || b == 'd' {
		// Go1.7 encoding; first byte encodes low-level
		// encoding format (compact vs debug).
		// For backward-compatibility only (avoid problems with
		// old installed packages). Newly compiled packages use
		// the extensible format string.
		// TODO(gri) Remove this support eventually; after Go1.8.
		if b == 'd' {
			p.debugFormat = true
		}
		p.trackAllTypes = p.rawByte() == 'a'
		p.posInfoFormat = p.int() != 0
		versionstr = p.string()
		if versionstr == "v1" {
			p.version = 0
		}
	} else {
		// Go1.8 extensible encoding
		// read version string and extract version number (ignore anything after the version number)
		versionstr = p.rawStringln(b)
		if s := strings.SplitN(versionstr, " ", 3); len(s) >= 2 && s[0] == "version" {
			if v, err := strconv.Atoi(s[1]); err == nil && v > 0 {
				p.version = v
			}
		}
	}

	// read version specific flags - extend as necessary
	switch p.version {
	// case 6:
	// 	...
	//	fallthrough
	case 5, 4, 3, 2, 1:
		p.debugFormat = p.rawStringln(p.rawByte()) == "debug"
		p.trackAllTypes = p.int() != 0
		p.posInfoFormat = p.int() != 0
	case 0:
		// Go1.7 encoding format - nothing to do here
	default:
		panic(fmt.Errorf("unknown export format version %d (%q)", p.version, versionstr))
	}

	// --- generic export data ---

	// populate typList with predeclared "known" types
	p.typList = append(p.typList, predeclared...)

	// read package data
	pkgName := p.pkg()
	p.pfc.defalias = pkgName[strings.LastIndex(pkgName, "!")+1:]

	// read objects of phase 1 only (see cmd/compiler/internal/gc/bexport.go)
	objcount := 0
	for {
		tag := p.tagOrIndex()
		if tag == endTag {
			break
		}
		p.obj(tag)
		objcount++
	}

	// self-verification
	if count := p.int(); count != objcount {
		panic(fmt.Sprintf("got %d objects; want %d", objcount, count))
	}
}

func (p *gc_bin_parser) pkg() string {
	// if the package was seen before, i is its index (>= 0)
	i := p.tagOrIndex()
	if i >= 0 {
		return p.pkgList[i]
	}

	// otherwise, i is the package tag (< 0)
	if i != packageTag {
		panic(fmt.Sprintf("unexpected package tag %d version %d", i, p.version))
	}

	// read package data
	name := p.string()
	var path string
	if p.version >= 5 {
		path = p.path()
	} else {
		path = p.string()
	}

	// we should never see an empty package name
	if name == "" {
		panic("empty package name in import")
	}

	// an empty path denotes the package we are currently importing;
	// it must be the first package we see
	if (path == "") != (len(p.pkgList) == 0) {
		panic(fmt.Sprintf("package path %q for pkg index %d", path, len(p.pkgList)))
	}

	var fullName string
	if path != "" {
		fullName = "!" + path + "!" + name
		p.pfc.add_package_to_scope(fullName, path)
	} else {
		fullName = "!" + p.pfc.name + "!" + name
	}

	// if the package was imported before, use that one; otherwise create a new one
	p.pkgList = append(p.pkgList, fullName)
	return p.pkgList[len(p.pkgList)-1]
}

func (p *gc_bin_parser) obj(tag int) {
	switch tag {
	case constTag:
		p.pos()
		pkg, name := p.qualifiedName()
		typ := p.typ("")
		p.skipValue() // ignore const value, gocode's not interested
		p.callback(pkg, &ast.GenDecl{
			Tok: token.CONST,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names:  []*ast.Ident{ast.NewIdent(name)},
					Type:   typ,
					Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
				},
			},
		})

	case aliasTag:
		// TODO(gri) verify type alias hookup is correct
		p.pos()
		pkg, name := p.qualifiedName()
		typ := p.typ("")
		p.callback(pkg, &ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{typeAliasSpec(name, typ)},
		})

	case typeTag:
		_ = p.typ("")

	case varTag:
		p.pos()
		pkg, name := p.qualifiedName()
		typ := p.typ("")
		p.callback(pkg, &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent(name)},
					Type:  typ,
				},
			},
		})

	case funcTag:
		p.pos()
		pkg, name := p.qualifiedName()
		params := p.paramList()
		results := p.paramList()
		p.callback(pkg, &ast.FuncDecl{
			Name: ast.NewIdent(name),
			Type: &ast.FuncType{Params: params, Results: results},
		})

	default:
		panic(fmt.Sprintf("unexpected object tag %d", tag))
	}
}

const deltaNewFile = -64 // see cmd/compile/internal/gc/bexport.go

func (p *gc_bin_parser) pos() {
	if !p.posInfoFormat {
		return
	}

	file := p.prevFile
	line := p.prevLine
	delta := p.int()
	line += delta
	if p.version >= 5 {
		if delta == deltaNewFile {
			if n := p.int(); n >= 0 {
				// file changed
				file = p.path()
				line = n
			}
		}
	} else {
		if delta == 0 {
			if n := p.int(); n >= 0 {
				// file changed
				file = p.prevFile[:n] + p.string()
				line = p.int()
			}
		}
	}
	p.prevFile = file
	p.prevLine = line

	// TODO(gri) register new position
}

func (p *gc_bin_parser) qualifiedName() (pkg string, name string) {
	name = p.string()
	pkg = p.pkg()
	return pkg, name
}

func (p *gc_bin_parser) reserveMaybe() int {
	if p.trackAllTypes {
		p.typList = append(p.typList, nil)
		return len(p.typList) - 1
	} else {
		return -1
	}
}

func (p *gc_bin_parser) recordMaybe(idx int, t ast.Expr) ast.Expr {
	if idx == -1 {
		return t
	}
	p.typList[idx] = t
	return t
}

func (p *gc_bin_parser) record(t ast.Expr) {
	p.typList = append(p.typList, t)
}

// parent is the package which declared the type; parent == nil means
// the package currently imported. The parent package is needed for
// exported struct fields and interface methods which don't contain
// explicit package information in the export data.
func (p *gc_bin_parser) typ(parent string) ast.Expr {
	// if the type was seen before, i is its index (>= 0)
	i := p.tagOrIndex()
	if i >= 0 {
		return p.typList[i]
	}

	// otherwise, i is the type tag (< 0)
	switch i {
	case namedTag:
		// read type object
		p.pos()
		parent, name := p.qualifiedName()
		tdecl := &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(name),
				},
			},
		}

		// record it right away (underlying type can contain refs to t)
		t := &ast.SelectorExpr{X: ast.NewIdent(parent), Sel: ast.NewIdent(name)}
		p.record(t)

		// parse underlying type
		t0 := p.typ(parent)
		tdecl.Specs[0].(*ast.TypeSpec).Type = t0

		p.callback(parent, tdecl)

		// interfaces have no methods
		if _, ok := t0.(*ast.InterfaceType); ok {
			return t
		}

		// read associated methods
		for i := p.int(); i > 0; i-- {
			// TODO(gri) replace this with something closer to fieldName
			p.pos()
			name := p.string()
			if !exported(name) {
				p.pkg()
			}

			recv := p.paramList()
			params := p.paramList()
			results := p.paramList()
			p.int() // go:nointerface pragma - discarded

			strip_method_receiver(recv)
			p.callback(parent, &ast.FuncDecl{
				Recv: recv,
				Name: ast.NewIdent(name),
				Type: &ast.FuncType{Params: params, Results: results},
			})
		}
		return t
	case arrayTag:
		i := p.reserveMaybe()
		n := p.int64()
		elt := p.typ(parent)
		return p.recordMaybe(i, &ast.ArrayType{
			Len: &ast.BasicLit{Kind: token.INT, Value: fmt.Sprint(n)},
			Elt: elt,
		})

	case sliceTag:
		i := p.reserveMaybe()
		elt := p.typ(parent)
		return p.recordMaybe(i, &ast.ArrayType{Len: nil, Elt: elt})

	case dddTag:
		i := p.reserveMaybe()
		elt := p.typ(parent)
		return p.recordMaybe(i, &ast.Ellipsis{Elt: elt})

	case structTag:
		i := p.reserveMaybe()
		return p.recordMaybe(i, p.structType(parent))

	case pointerTag:
		i := p.reserveMaybe()
		elt := p.typ(parent)
		return p.recordMaybe(i, &ast.StarExpr{X: elt})

	case signatureTag:
		i := p.reserveMaybe()
		params := p.paramList()
		results := p.paramList()
		return p.recordMaybe(i, &ast.FuncType{Params: params, Results: results})

	case interfaceTag:
		i := p.reserveMaybe()
		var embeddeds []*ast.SelectorExpr
		for n := p.int(); n > 0; n-- {
			p.pos()
			if named, ok := p.typ(parent).(*ast.SelectorExpr); ok {
				embeddeds = append(embeddeds, named)
			}
		}
		methods := p.methodList(parent)
		for _, field := range embeddeds {
			methods = append(methods, &ast.Field{Type: field})
		}
		return p.recordMaybe(i, &ast.InterfaceType{Methods: &ast.FieldList{List: methods}})

	case mapTag:
		i := p.reserveMaybe()
		key := p.typ(parent)
		val := p.typ(parent)
		return p.recordMaybe(i, &ast.MapType{Key: key, Value: val})

	case chanTag:
		i := p.reserveMaybe()
		dir := ast.SEND | ast.RECV
		switch d := p.int(); d {
		case 1:
			dir = ast.RECV
		case 2:
			dir = ast.SEND
		case 3:
			// already set
		default:
			panic(fmt.Sprintf("unexpected channel dir %d", d))
		}
		elt := p.typ(parent)
		return p.recordMaybe(i, &ast.ChanType{Dir: dir, Value: elt})

	default:
		panic(fmt.Sprintf("unexpected type tag %d", i))
	}
}

func (p *gc_bin_parser) structType(parent string) *ast.StructType {
	var fields []*ast.Field
	if n := p.int(); n > 0 {
		fields = make([]*ast.Field, n)
		for i := range fields {
			fields[i], _ = p.field(parent) // (*ast.Field, tag), not interested in tags
		}
	}
	return &ast.StructType{Fields: &ast.FieldList{List: fields}}
}

func (p *gc_bin_parser) field(parent string) (*ast.Field, string) {
	p.pos()
	_, name, _ := p.fieldName(parent)
	typ := p.typ(parent)
	tag := p.string()

	var names []*ast.Ident
	if name != "" {
		names = []*ast.Ident{ast.NewIdent(name)}
	}
	return &ast.Field{
		Names: names,
		Type:  typ,
	}, tag
}

func (p *gc_bin_parser) methodList(parent string) (methods []*ast.Field) {
	if n := p.int(); n > 0 {
		methods = make([]*ast.Field, n)
		for i := range methods {
			methods[i] = p.method(parent)
		}
	}
	return
}

func (p *gc_bin_parser) method(parent string) *ast.Field {
	p.pos()
	_, name, _ := p.fieldName(parent)
	params := p.paramList()
	results := p.paramList()
	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  &ast.FuncType{Params: params, Results: results},
	}
}

func (p *gc_bin_parser) fieldName(parent string) (string, string, bool) {
	name := p.string()
	pkg := parent
	if p.version == 0 && name == "_" {
		// version 0 didn't export a package for _ fields
		return pkg, name, false
	}
	var alias bool
	switch name {
	case "":
		// 1) field name matches base type name and is exported: nothing to do
	case "?":
		// 2) field name matches base type name and is not exported: need package
		name = ""
		pkg = p.pkg()
	case "@":
		// 3) field name doesn't match type name (alias)
		name = p.string()
		alias = true
		fallthrough
	default:
		if !exported(name) {
			pkg = p.pkg()
		}
	}
	return pkg, name, alias
}

func (p *gc_bin_parser) paramList() *ast.FieldList {
	n := p.int()
	if n == 0 {
		return nil
	}
	// negative length indicates unnamed parameters
	named := true
	if n < 0 {
		n = -n
		named = false
	}
	// n > 0
	flds := make([]*ast.Field, n)
	for i := range flds {
		flds[i] = p.param(named)
	}
	return &ast.FieldList{List: flds}
}

func (p *gc_bin_parser) param(named bool) *ast.Field {
	t := p.typ("")

	name := "?"
	if named {
		name = p.string()
		if name == "" {
			panic("expected named parameter")
		}
		if name != "_" {
			p.pkg()
		}
		if i := strings.Index(name, "·"); i > 0 {
			name = name[:i] // cut off gc-specific parameter numbering
		}
	}

	// read and discard compiler-specific info
	p.string()

	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  t,
	}
}

func exported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

func (p *gc_bin_parser) skipValue() {
	switch tag := p.tagOrIndex(); tag {
	case falseTag, trueTag:
	case int64Tag:
		p.int64()
	case floatTag:
		p.float()
	case complexTag:
		p.float()
		p.float()
	case stringTag:
		p.string()
	default:
		panic(fmt.Sprintf("unexpected value tag %d", tag))
	}
}

func (p *gc_bin_parser) float() {
	sign := p.int()
	if sign == 0 {
		return
	}

	p.int()    // exp
	p.string() // mant
}

// ----------------------------------------------------------------------------
// Low-level decoders

func (p *gc_bin_parser) tagOrIndex() int {
	if p.debugFormat {
		p.marker('t')
	}

	return int(p.rawInt64())
}

func (p *gc_bin_parser) int() int {
	x := p.int64()
	if int64(int(x)) != x {
		panic("exported integer too large")
	}
	return int(x)
}

func (p *gc_bin_parser) int64() int64 {
	if p.debugFormat {
		p.marker('i')
	}

	return p.rawInt64()
}

func (p *gc_bin_parser) path() string {
	if p.debugFormat {
		p.marker('p')
	}
	// if the path was seen before, i is its index (>= 0)
	// (the empty string is at index 0)
	i := p.rawInt64()
	if i >= 0 {
		return p.pathList[i]
	}
	// otherwise, i is the negative path length (< 0)
	a := make([]string, -i)
	for n := range a {
		a[n] = p.string()
	}
	s := strings.Join(a, "/")
	p.pathList = append(p.pathList, s)
	return s
}

func (p *gc_bin_parser) string() string {
	if p.debugFormat {
		p.marker('s')
	}
	// if the string was seen before, i is its index (>= 0)
	// (the empty string is at index 0)
	i := p.rawInt64()
	if i >= 0 {
		return p.strList[i]
	}
	// otherwise, i is the negative string length (< 0)
	if n := int(-i); n <= cap(p.buf) {
		p.buf = p.buf[:n]
	} else {
		p.buf = make([]byte, n)
	}
	for i := range p.buf {
		p.buf[i] = p.rawByte()
	}
	s := string(p.buf)
	p.strList = append(p.strList, s)
	return s
}

func (p *gc_bin_parser) marker(want byte) {
	if got := p.rawByte(); got != want {
		panic(fmt.Sprintf("incorrect marker: got %c; want %c (pos = %d)", got, want, p.read))
	}

	pos := p.read
	if n := int(p.rawInt64()); n != pos {
		panic(fmt.Sprintf("incorrect position: got %d; want %d", n, pos))
	}
}

// rawInt64 should only be used by low-level decoders.
func (p *gc_bin_parser) rawInt64() int64 {
	i, err := binary.ReadVarint(p)
	if err != nil {
		panic(fmt.Sprintf("read error: %v", err))
	}
	return i
}

// rawStringln should only be used to read the initial version string.
func (p *gc_bin_parser) rawStringln(b byte) string {
	p.buf = p.buf[:0]
	for b != '\n' {
		p.buf = append(p.buf, b)
		b = p.rawByte()
	}
	return string(p.buf)
}

// needed for binary.ReadVarint in rawInt64
func (p *gc_bin_parser) ReadByte() (byte, error) {
	return p.rawByte(), nil
}

// byte is the bottleneck interface for reading p.data.
// It unescapes '|' 'S' to '$' and '|' '|' to '|'.
// rawByte should only be used by low-level decoders.
func (p *gc_bin_parser) rawByte() byte {
	b := p.data[0]
	r := 1
	if b == '|' {
		b = p.data[1]
		r = 2
		switch b {
		case 'S':
			b = '$'
		case '|':
			// nothing to do
		default:
			panic("unexpected escape sequence in export data")
		}
	}
	p.data = p.data[r:]
	p.read += r
	return b

}

// ----------------------------------------------------------------------------
// Export format

// Tags. Must be < 0.
const (
	// Objects
	packageTag = -(iota + 1)
	constTag
	typeTag
	varTag
	funcTag
	endTag

	// Types
	namedTag
	arrayTag
	sliceTag
	dddTag
	structTag
	pointerTag
	signatureTag
	interfaceTag
	mapTag
	chanTag

	// Values
	falseTag
	trueTag
	int64Tag
	floatTag
	fractionTag // not used by gc
	complexTag
	stringTag
	nilTag     // only used by gc (appears in exported inlined function bodies)
	unknownTag // not used by gc (only appears in packages with errors)

	// Type aliases
	aliasTag
)

var predeclared = []ast.Expr{
	// basic types
	ast.NewIdent("bool"),
	ast.NewIdent("int"),
	ast.NewIdent("int8"),
	ast.NewIdent("int16"),
	ast.NewIdent("int32"),
	ast.NewIdent("int64"),
	ast.NewIdent("uint"),
	ast.NewIdent("uint8"),
	ast.NewIdent("uint16"),
	ast.NewIdent("uint32"),
	ast.NewIdent("uint64"),
	ast.NewIdent("uintptr"),
	ast.NewIdent("float32"),
	ast.NewIdent("float64"),
	ast.NewIdent("complex64"),
	ast.NewIdent("complex128"),
	ast.NewIdent("string"),

	// basic type aliases
	ast.NewIdent("byte"),
	ast.NewIdent("rune"),

	// error
	ast.NewIdent("error"),

	// TODO(nsf): don't think those are used in just package type info,
	// maybe for consts, but we are not interested in that
	// untyped types
	ast.NewIdent(">_<"), // TODO: types.Typ[types.UntypedBool],
	ast.NewIdent(">_<"), // TODO: types.Typ[types.UntypedInt],
	ast.NewIdent(">_<"), // TODO: types.Typ[types.UntypedRune],
	ast.NewIdent(">_<"), // TODO: types.Typ[types.UntypedFloat],
	ast.NewIdent(">_<"), // TODO: types.Typ[types.UntypedComplex],
	ast.NewIdent(">_<"), // TODO: types.Typ[types.UntypedString],
	ast.NewIdent(">_<"), // TODO: types.Typ[types.UntypedNil],

	// package unsafe
	&ast.SelectorExpr{X: ast.NewIdent("unsafe"), Sel: ast.NewIdent("Pointer")},

	// invalid type
	ast.NewIdent(">_<"), // TODO: types.Typ[types.Invalid], // only appears in packages with errors

	// used internally by gc; never used by this package or in .a files
	ast.NewIdent("any"),
}
