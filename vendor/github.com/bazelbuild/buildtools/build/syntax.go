/*
Copyright 2016 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package build implements parsing and printing of BUILD files.
package build

// Syntax data structure definitions.

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// A Position describes the position between two bytes of input.
type Position struct {
	Line     int // line in input (starting at 1)
	LineRune int // rune in line (starting at 1)
	Byte     int // byte in input (starting at 0)
}

// add returns the position at the end of s, assuming it starts at p.
func (p Position) add(s string) Position {
	p.Byte += len(s)
	if n := strings.Count(s, "\n"); n > 0 {
		p.Line += n
		s = s[strings.LastIndex(s, "\n")+1:]
		p.LineRune = 1
	}
	p.LineRune += utf8.RuneCountInString(s)
	return p
}

// An Expr represents an input element.
type Expr interface {
	// Span returns the start and end position of the expression,
	// excluding leading or trailing comments.
	Span() (start, end Position)

	// Comment returns the comments attached to the expression.
	// This method would normally be named 'Comments' but that
	// would interfere with embedding a type of the same name.
	Comment() *Comments

	// Copy returns a non-deep copy of the node. Can be useful if
	// the actual node type is hidden by the Expr interface and
	// not relevant.
	Copy() Expr
}

// A Comment represents a single # comment.
type Comment struct {
	Start Position
	Token string // without trailing newline
}

// Span returns the start and end positions of the node
func (c Comment) Span() (start, end Position) {
	return c.Start, c.Start.add(c.Token)
}

// Comments collects the comments associated with an expression.
type Comments struct {
	Before []Comment // whole-line comments before this expression
	Suffix []Comment // end-of-line comments after this expression

	// For top-level expressions only, After lists whole-line
	// comments following the expression.
	After []Comment
}

// Comment returns the receiver. This isn't useful by itself, but
// a Comments struct is embedded into all the expression
// implementation types, and this gives each of those a Comment
// method to satisfy the Expr interface.
func (c *Comments) Comment() *Comments {
	return c
}

// stmtsEnd returns the end position of the last non-nil statement
func stmtsEnd(stmts []Expr) Position {
	for i := len(stmts) - 1; i >= 0; i-- {
		if stmts[i] != nil {
			_, end := stmts[i].Span()
			return end
		}
	}
	return Position{}
}

// A File represents an entire BUILD or .bzl file.
type File struct {
	Path          string // absolute file path
	Pkg           string // optional; the package of the file (always forward slashes)
	Label         string // optional; file path relative to the package name (always forward slashes)
	WorkspaceRoot string // optional; path to the directory containing the WORKSPACE file
	Type          FileType
	Comments
	Stmt []Expr
}

// DisplayPath returns the filename if it's not empty, "<stdin>" otherwise
func (f *File) DisplayPath() string {
	if f.Path == "" {
		return "<stdin>"
	}
	return f.Path
}

// CanonicalPath returns the path of a file relative to the workspace root with forward slashes only
func (f *File) CanonicalPath() string {
	if f.Pkg == "" {
		return "//" + f.Label
	}
	return fmt.Sprintf("//%s/%s", f.Pkg, f.Label)
}

// Span returns the start and end positions of the node
func (f *File) Span() (start, end Position) {
	if len(f.Stmt) == 0 {
		p := Position{Line: 1, LineRune: 1}
		return p, p
	}
	start = Position{}
	end = stmtsEnd(f.Stmt)
	return start, end
}

//Copy creates and returns a non-deep copy of File
func (f *File) Copy() Expr {
	n := *f
	return &n
}

// A CommentBlock represents a top-level block of comments separate
// from any rule.
type CommentBlock struct {
	Comments
	Start Position
}

// Span returns the start and end positions of the node
func (x *CommentBlock) Span() (start, end Position) {
	return x.Start, x.Start
}

//Copy creates and returns a non-deep copy of CommentBlock
func (x *CommentBlock) Copy() Expr {
	n := *x
	return &n
}

// An Ident represents an identifier.
type Ident struct {
	Comments
	NamePos Position
	Name    string
}

// Span returns the start and end positions of the node
func (x *Ident) Span() (start, end Position) {
	return x.NamePos, x.NamePos.add(x.Name)
}

//Copy creates and returns a non-deep copy of Ident
func (x *Ident) Copy() Expr {
	n := *x
	return &n
}

func (x *Ident) asString() *StringExpr {
	_, end := x.Span()
	return &StringExpr{
		Comments: x.Comments,
		Start:    x.NamePos,
		Value:    x.Name,
		End:      end,
	}
}

// An TypedIdent represents an identifier with type annotation: "foo: int".
type TypedIdent struct {
	Comments
	Ident *Ident
	Type  Expr
}

// Span returns the start and end positions of the node
func (x *TypedIdent) Span() (start, end Position) {
	start, _ = x.Ident.Span()
	_, end = x.Type.Span()
	return start, end
}

//Copy creates and returns a non-deep copy of TypedIdent
func (x *TypedIdent) Copy() Expr {
	n := *x
	return &n
}

// BranchStmt represents a `pass`, `break`, or `continue` statement.
type BranchStmt struct {
	Comments
	Token    string // pass, break, continue
	TokenPos Position
}

// Span returns the start and end positions of the node
func (x *BranchStmt) Span() (start, end Position) {
	return x.TokenPos, x.TokenPos.add(x.Token)
}

//Copy creates and returns a non-deep copy of BranchStmt
func (x *BranchStmt) Copy() Expr {
	n := *x
	return &n
}

// A LiteralExpr represents a literal number.
type LiteralExpr struct {
	Comments
	Start Position
	Token string // identifier token
}

// Span returns the start and end positions of the node
func (x *LiteralExpr) Span() (start, end Position) {
	return x.Start, x.Start.add(x.Token)
}

//Copy creates and returns a non-deep copy of LiteralExpr
func (x *LiteralExpr) Copy() Expr {
	n := *x
	return &n
}

// A StringExpr represents a single literal string.
type StringExpr struct {
	Comments
	Start       Position
	Value       string // string value (decoded)
	TripleQuote bool   // triple quote output
	End         Position

	// To allow specific formatting of string literals,
	// at least within our requirements, record the
	// preferred form of Value. This field is a hint:
	// it is only used if it is a valid quoted form for Value.
	Token string
}

// Span returns the start and end positions of the node
func (x *StringExpr) Span() (start, end Position) {
	return x.Start, x.End
}

//Copy creates and returns a non-deep copy of StringExpr
func (x *StringExpr) Copy() Expr {
	n := *x
	return &n
}

// An End represents the end of a parenthesized or bracketed expression.
// It is a place to hang comments.
type End struct {
	Comments
	Pos Position
}

// Span returns the start and end positions of the node
func (x *End) Span() (start, end Position) {
	return x.Pos, x.Pos.add(")")
}

//Copy creates and returns a non-deep copy of End
func (x *End) Copy() Expr {
	n := *x
	return &n
}

// A CallExpr represents a function call expression: X(List).
type CallExpr struct {
	Comments
	X              Expr
	ListStart      Position // position of (
	List           []Expr
	End                 // position of )
	ForceCompact   bool // force compact (non-multiline) form when printing
	ForceMultiLine bool // force multiline form when printing
}

// Span returns the start and end positions of the node
func (x *CallExpr) Span() (start, end Position) {
	start, _ = x.X.Span()
	return start, x.End.Pos.add(")")
}

//Copy creates and returns a non-deep copy of CallExpr
func (x *CallExpr) Copy() Expr {
	n := *x
	return &n
}

// A DotExpr represents a field selector: X.Name.
type DotExpr struct {
	Comments
	X       Expr
	Dot     Position
	NamePos Position
	Name    string
}

// Span returns the start and end positions of the node
func (x *DotExpr) Span() (start, end Position) {
	start, _ = x.X.Span()
	return start, x.NamePos.add(x.Name)
}

//Copy creates and returns a non-deep copy of DotExpr
func (x *DotExpr) Copy() Expr {
	n := *x
	return &n
}

// A Comprehension represents a list comprehension expression: [X for ... if ...].
type Comprehension struct {
	Comments
	Curly          bool // curly braces (as opposed to square brackets)
	Lbrack         Position
	Body           Expr
	Clauses        []Expr // = *ForClause | *IfClause
	ForceMultiLine bool   // split expression across multiple lines
	End
}

// Span returns the start and end positions of the node
func (x *Comprehension) Span() (start, end Position) {
	return x.Lbrack, x.End.Pos.add("]")
}

//Copy creates and returns a non-deep copy of Comprehension
func (x *Comprehension) Copy() Expr {
	n := *x
	return &n
}

// A ForClause represents a for clause in a list comprehension: for Var in Expr.
type ForClause struct {
	Comments
	For  Position
	Vars Expr
	In   Position
	X    Expr
}

// Span returns the start and end positions of the node
func (x *ForClause) Span() (start, end Position) {
	_, end = x.X.Span()
	return x.For, end
}

//Copy creates and returns a non-deep copy of ForClause
func (x *ForClause) Copy() Expr {
	n := *x
	return &n
}

// An IfClause represents an if clause in a list comprehension: if Cond.
type IfClause struct {
	Comments
	If   Position
	Cond Expr
}

// Span returns the start and end positions of the node
func (x *IfClause) Span() (start, end Position) {
	_, end = x.Cond.Span()
	return x.If, end
}

//Copy creates and returns a non-deep copy of IfClause
func (x *IfClause) Copy() Expr {
	n := *x
	return &n
}

// A KeyValueExpr represents a dictionary entry: Key: Value.
type KeyValueExpr struct {
	Comments
	Key   Expr
	Colon Position
	Value Expr
}

// Span returns the start and end positions of the node
func (x *KeyValueExpr) Span() (start, end Position) {
	start, _ = x.Key.Span()
	_, end = x.Value.Span()
	return start, end
}

//Copy creates and returns a non-deep copy of KeyValueExpr
func (x *KeyValueExpr) Copy() Expr {
	n := *x
	return &n
}

// A DictExpr represents a dictionary literal: { List }.
type DictExpr struct {
	Comments
	Start Position
	List  []*KeyValueExpr
	End
	ForceMultiLine bool // force multiline form when printing
}

// Span returns the start and end positions of the node
func (x *DictExpr) Span() (start, end Position) {
	return x.Start, x.End.Pos.add("}")
}

//Copy creates and returns a non-deep copy of DictExpr
func (x *DictExpr) Copy() Expr {
	n := *x
	return &n
}

// A ListExpr represents a list literal: [ List ].
type ListExpr struct {
	Comments
	Start Position
	List  []Expr
	End
	ForceMultiLine bool // force multiline form when printing
}

// Span returns the start and end positions of the node
func (x *ListExpr) Span() (start, end Position) {
	return x.Start, x.End.Pos.add("]")
}

//Copy creates and returns a non-deep copy of ListExpr
func (x *ListExpr) Copy() Expr {
	n := *x
	return &n
}

// A SetExpr represents a set literal: { List }.
type SetExpr struct {
	Comments
	Start Position
	List  []Expr
	End
	ForceMultiLine bool // force multiline form when printing
}

// Span returns the start and end positions of the node
func (x *SetExpr) Span() (start, end Position) {
	return x.Start, x.End.Pos.add("}")
}

//Copy creates and returns a non-deep copy of SetExpr
func (x *SetExpr) Copy() Expr {
	n := *x
	return &n
}

// A TupleExpr represents a tuple literal: (List)
type TupleExpr struct {
	Comments
	NoBrackets bool // true if a tuple has no brackets, e.g. `a, b = x`
	Start      Position
	List       []Expr
	End
	ForceCompact   bool // force compact (non-multiline) form when printing
	ForceMultiLine bool // force multiline form when printing
}

// Span returns the start and end positions of the node
func (x *TupleExpr) Span() (start, end Position) {
	if !x.NoBrackets {
		return x.Start, x.End.Pos.add(")")
	}
	start, _ = x.List[0].Span()
	_, end = x.List[len(x.List)-1].Span()
	return start, end
}

//Copy creates and returns a non-deep copy of TupleExpr
func (x *TupleExpr) Copy() Expr {
	n := *x
	return &n
}

// A UnaryExpr represents a unary expression: Op X.
type UnaryExpr struct {
	Comments
	OpStart Position
	Op      string
	X       Expr
}

// Span returns the start and end positions of the node
func (x *UnaryExpr) Span() (start, end Position) {
	if x.X == nil {
		return x.OpStart, x.OpStart
	}
	_, end = x.X.Span()
	return x.OpStart, end
}

//Copy creates and returns a non-deep copy of UnaryExpr
func (x *UnaryExpr) Copy() Expr {
	n := *x
	return &n
}

// A BinaryExpr represents a binary expression: X Op Y.
type BinaryExpr struct {
	Comments
	X         Expr
	OpStart   Position
	Op        string
	LineBreak bool // insert line break between Op and Y
	Y         Expr
}

// Span returns the start and end positions of the node
func (x *BinaryExpr) Span() (start, end Position) {
	start, _ = x.X.Span()
	_, end = x.Y.Span()
	return start, end
}

//Copy creates and returns a non-deep copy of BinaryExpr
func (x *BinaryExpr) Copy() Expr {
	n := *x
	return &n
}

// An AssignExpr represents a binary expression with `=`: LHS = RHS.
type AssignExpr struct {
	Comments
	LHS       Expr
	OpPos     Position
	Op        string
	LineBreak bool // insert line break between Op and RHS
	RHS       Expr
}

// Span returns the start and end positions of the node
func (x *AssignExpr) Span() (start, end Position) {
	start, _ = x.LHS.Span()
	_, end = x.RHS.Span()
	return start, end
}

//Copy creates and returns a non-deep copy of AssignExpr
func (x *AssignExpr) Copy() Expr {
	n := *x
	return &n
}

// A ParenExpr represents a parenthesized expression: (X).
type ParenExpr struct {
	Comments
	Start Position
	X     Expr
	End
	ForceMultiLine bool // insert line break after opening ( and before closing )
}

// Span returns the start and end positions of the node
func (x *ParenExpr) Span() (start, end Position) {
	return x.Start, x.End.Pos.add(")")
}

//Copy creates and returns a non-deep copy of ParenExpr
func (x *ParenExpr) Copy() Expr {
	n := *x
	return &n
}

// A SliceExpr represents a slice expression: expr[from:to] or expr[from:to:step] .
type SliceExpr struct {
	Comments
	X           Expr
	SliceStart  Position
	From        Expr
	FirstColon  Position
	To          Expr
	SecondColon Position
	Step        Expr
	End         Position
}

// Span returns the start and end positions of the node
func (x *SliceExpr) Span() (start, end Position) {
	start, _ = x.X.Span()
	return start, x.End.add("]")
}

//Copy creates and returns a non-deep copy of SliceExpr
func (x *SliceExpr) Copy() Expr {
	n := *x
	return &n
}

// An IndexExpr represents an index expression: X[Y].
type IndexExpr struct {
	Comments
	X          Expr
	IndexStart Position
	Y          Expr
	End        Position
}

// Span returns the start and end positions of the node
func (x *IndexExpr) Span() (start, end Position) {
	start, _ = x.X.Span()
	return start, x.End.add("]")
}

//Copy creates and returns a non-deep copy of IndexExpr
func (x *IndexExpr) Copy() Expr {
	n := *x
	return &n
}

// A Function represents the common parts of LambdaExpr and DefStmt
type Function struct {
	Comments
	StartPos Position // position of DEF or LAMBDA token
	Params   []Expr
	Body     []Expr
}

// Span returns the start and end positions of the node
func (x *Function) Span() (start, end Position) {
	_, end = x.Body[len(x.Body)-1].Span()
	return x.StartPos, end
}

//Copy creates and returns a non-deep copy of Function
func (x *Function) Copy() Expr {
	n := *x
	return &n
}

// A LambdaExpr represents a lambda expression: lambda Var: Expr.
type LambdaExpr struct {
	Comments
	Function
}

// Span returns the start and end positions of the node
func (x *LambdaExpr) Span() (start, end Position) {
	return x.Function.Span()
}

//Copy creates and returns a non-deep copy of LambdaExpr
func (x *LambdaExpr) Copy() Expr {
	n := *x
	return &n
}

// ConditionalExpr represents the conditional: X if TEST else ELSE.
type ConditionalExpr struct {
	Comments
	Then      Expr
	IfStart   Position
	Test      Expr
	ElseStart Position
	Else      Expr
}

// Span returns the start and end position of the expression,
// excluding leading or trailing comments.
// Span returns the start and end positions of the node
func (x *ConditionalExpr) Span() (start, end Position) {
	start, _ = x.Then.Span()
	_, end = x.Else.Span()
	return start, end
}

//Copy creates and returns a non-deep copy of ConditionalExpr
func (x *ConditionalExpr) Copy() Expr {
	n := *x
	return &n
}

// A LoadStmt loads another module and binds names from it:
// load(Module, "x", y="foo").
//
// The AST is slightly unfaithful to the concrete syntax here because
// Skylark's load statement, so that it can be implemented in Python,
// binds some names (like y above) with an identifier and some (like x)
// without.  For consistency we create fake identifiers for all the
// strings.
type LoadStmt struct {
	Comments
	Load         Position
	Module       *StringExpr
	From         []*Ident // name defined in loading module
	To           []*Ident // name in loaded module
	Rparen       End
	ForceCompact bool // force compact (non-multiline) form when printing
}

// Span returns the start and end positions of the node
func (x *LoadStmt) Span() (start, end Position) {
	return x.Load, x.Rparen.Pos.add(")")
}

//Copy creates and returns a non-deep copy of LoadStmt
func (x *LoadStmt) Copy() Expr {
	n := *x
	return &n
}

// A DefStmt represents a function definition expression: def foo(List):.
type DefStmt struct {
	Comments
	Function
	Name           string
	ColonPos       Position // position of the ":"
	ForceCompact   bool     // force compact (non-multiline) form when printing the arguments
	ForceMultiLine bool     // force multiline form when printing the arguments
	Type           Expr     // type annotation
}

// Span returns the start and end positions of the node
func (x *DefStmt) Span() (start, end Position) {
	return x.Function.Span()
}

//Copy creates and returns a non-deep copy of DefStmt
func (x *DefStmt) Copy() Expr {
	n := *x
	return &n
}

// HeaderSpan returns the span of the function header `def f(...):`
func (x *DefStmt) HeaderSpan() (start, end Position) {
	return x.Function.StartPos, x.ColonPos
}

// A ReturnStmt represents a return statement: return f(x).
type ReturnStmt struct {
	Comments
	Return Position
	Result Expr // may be nil
}

// Span returns the start and end positions of the node
func (x *ReturnStmt) Span() (start, end Position) {
	if x.Result == nil {
		return x.Return, x.Return.add("return")
	}
	_, end = x.Result.Span()
	return x.Return, end
}

//Copy creates and returns a non-deep copy of ReturnStmt
func (x *ReturnStmt) Copy() Expr {
	n := *x
	return &n
}

// A ForStmt represents a for loop block: for x in range(10):.
type ForStmt struct {
	Comments
	Function
	For  Position // position of for
	Vars Expr
	X    Expr
	Body []Expr
}

// Span returns the start and end positions of the node
func (x *ForStmt) Span() (start, end Position) {
	end = stmtsEnd(x.Body)
	return x.For, end
}

//Copy creates and returns a non-deep copy of ForStmt
func (x *ForStmt) Copy() Expr {
	n := *x
	return &n
}

// An IfStmt represents an if-else block: if x: ... else: ... .
// `elif`s are treated as a chain of `IfStmt`s.
type IfStmt struct {
	Comments
	If      Position // position of if
	Cond    Expr
	True    []Expr
	ElsePos End    // position of else or elif
	False   []Expr // optional
}

// Span returns the start and end positions of the node
func (x *IfStmt) Span() (start, end Position) {
	body := x.False
	if body == nil {
		body = x.True
	}
	end = stmtsEnd(body)
	return x.If, end
}

//Copy creates and returns a non-deep copy of IfStmt
func (x *IfStmt) Copy() Expr {
	n := *x
	return &n
}
