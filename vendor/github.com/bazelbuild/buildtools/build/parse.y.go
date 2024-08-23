//line build/parse.y:29
package build

import __yyfmt__ "fmt"

//line build/parse.y:29

//line build/parse.y:34
type yySymType struct {
	yys int
	// input tokens
	tok    string   // raw input syntax
	str    string   // decoding of quoted string
	pos    Position // position of token
	triple bool     // was string triple quoted?

	// partial syntax trees
	expr    Expr
	exprs   []Expr
	kv      *KeyValueExpr
	kvs     []*KeyValueExpr
	string  *StringExpr
	ifstmt  *IfStmt
	loadarg *struct {
		from Ident
		to   Ident
	}
	loadargs []*struct {
		from Ident
		to   Ident
	}
	def_header *DefStmt // partially filled in def statement, without the body

	// supporting information
	comma    Position // position of trailing comma in list, if present
	lastStmt Expr     // most recent rule, to attach line comments to
}

const _AUGM = 57346
const _AND = 57347
const _COMMENT = 57348
const _EOF = 57349
const _EQ = 57350
const _FOR = 57351
const _GE = 57352
const _IDENT = 57353
const _INT = 57354
const _IF = 57355
const _ELSE = 57356
const _ELIF = 57357
const _IN = 57358
const _IS = 57359
const _LAMBDA = 57360
const _LOAD = 57361
const _LE = 57362
const _NE = 57363
const _STAR_STAR = 57364
const _INT_DIV = 57365
const _BIT_LSH = 57366
const _BIT_RSH = 57367
const _ARROW = 57368
const _NOT = 57369
const _OR = 57370
const _STRING = 57371
const _DEF = 57372
const _RETURN = 57373
const _PASS = 57374
const _BREAK = 57375
const _CONTINUE = 57376
const _INDENT = 57377
const _UNINDENT = 57378
const ShiftInstead = 57379
const _ASSERT = 57380
const _UNARY = 57381

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"'%'",
	"'('",
	"')'",
	"'*'",
	"'+'",
	"','",
	"'-'",
	"'.'",
	"'/'",
	"':'",
	"'<'",
	"'='",
	"'>'",
	"'['",
	"']'",
	"'{'",
	"'}'",
	"'|'",
	"'&'",
	"'^'",
	"'~'",
	"_AUGM",
	"_AND",
	"_COMMENT",
	"_EOF",
	"_EQ",
	"_FOR",
	"_GE",
	"_IDENT",
	"_INT",
	"_IF",
	"_ELSE",
	"_ELIF",
	"_IN",
	"_IS",
	"_LAMBDA",
	"_LOAD",
	"_LE",
	"_NE",
	"_STAR_STAR",
	"_INT_DIV",
	"_BIT_LSH",
	"_BIT_RSH",
	"_ARROW",
	"_NOT",
	"_OR",
	"_STRING",
	"_DEF",
	"_RETURN",
	"_PASS",
	"_BREAK",
	"_CONTINUE",
	"_INDENT",
	"_UNINDENT",
	"ShiftInstead",
	"'\\n'",
	"_ASSERT",
	"_UNARY",
	"';'",
}

var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line build/parse.y:1058

// Go helper code.

// unary returns a unary expression with the given
// position, operator, and subexpression.
func unary(pos Position, op string, x Expr) Expr {
	return &UnaryExpr{
		OpStart: pos,
		Op:      op,
		X:       x,
	}
}

// binary returns a binary expression with the given
// operands, position, and operator.
func binary(x Expr, pos Position, op string, y Expr) Expr {
	_, xend := x.Span()
	ystart, _ := y.Span()

	switch op {
	case "=", "+=", "-=", "*=", "/=", "//=", "%=", "|=":
		return &AssignExpr{
			LHS:       x,
			OpPos:     pos,
			Op:        op,
			LineBreak: xend.Line < ystart.Line,
			RHS:       y,
		}
	}

	return &BinaryExpr{
		X:         x,
		OpStart:   pos,
		Op:        op,
		LineBreak: xend.Line < ystart.Line,
		Y:         y,
	}
}

// typed returns a TypedIdent expression
func typed(x, y Expr) *TypedIdent {
	return &TypedIdent{
		Ident: x.(*Ident),
		Type:  y,
	}
}

// isSimpleExpression returns whether an expression is simple and allowed to exist in
// compact forms of sequences.
// The formal criteria are the following: an expression is considered simple if it's
// a literal (variable, string or a number), a literal with a unary operator or an empty sequence.
func isSimpleExpression(expr *Expr) bool {
	switch x := (*expr).(type) {
	case *LiteralExpr, *StringExpr, *Ident:
		return true
	case *UnaryExpr:
		_, literal := x.X.(*LiteralExpr)
		_, ident := x.X.(*Ident)
		return literal || ident
	case *ListExpr:
		return len(x.List) == 0
	case *TupleExpr:
		return len(x.List) == 0
	case *DictExpr:
		return len(x.List) == 0
	case *SetExpr:
		return len(x.List) == 0
	default:
		return false
	}
}

// forceCompact returns the setting for the ForceCompact field for a call or tuple.
//
// NOTE 1: The field is called ForceCompact, not ForceSingleLine,
// because it only affects the formatting associated with the call or tuple syntax,
// not the formatting of the arguments. For example:
//
//	call([
//		1,
//		2,
//		3,
//	])
//
// is still a compact call even though it runs on multiple lines.
//
// In contrast the multiline form puts a linebreak after the (.
//
//	call(
//		[
//			1,
//			2,
//			3,
//		],
//	)
//
// NOTE 2: Because of NOTE 1, we cannot use start and end on the
// same line as a signal for compact mode: the formatting of an
// embedded list might move the end to a different line, which would
// then look different on rereading and cause buildifier not to be
// idempotent. Instead, we have to look at properties guaranteed
// to be preserved by the reformatting, namely that the opening
// paren and the first expression are on the same line and that
// each subsequent expression begins on the same line as the last
// one ended (no line breaks after comma).
func forceCompact(start Position, list []Expr, end Position) bool {
	if len(list) <= 1 {
		// The call or tuple will probably be compact anyway; don't force it.
		return false
	}

	// If there are any named arguments or non-string, non-literal
	// arguments, cannot force compact mode.
	line := start.Line
	for _, x := range list {
		start, end := x.Span()
		if start.Line != line {
			return false
		}
		line = end.Line
		if !isSimpleExpression(&x) {
			return false
		}
	}
	return end.Line == line
}

// forceMultiLine returns the setting for the ForceMultiLine field.
func forceMultiLine(start Position, list []Expr, end Position) bool {
	if len(list) > 1 {
		// The call will be multiline anyway, because it has multiple elements. Don't force it.
		return false
	}

	if len(list) == 0 {
		// Empty list: use position of brackets.
		return start.Line != end.Line
	}

	// Single-element list.
	// Check whether opening bracket is on different line than beginning of
	// element, or closing bracket is on different line than end of element.
	elemStart, elemEnd := list[0].Span()
	return start.Line != elemStart.Line || end.Line != elemEnd.Line
}

// forceMultiLineComprehension returns the setting for the ForceMultiLine field for a comprehension.
func forceMultiLineComprehension(start Position, expr Expr, clauses []Expr, end Position) bool {
	// Return true if there's at least one line break between start, expr, each clause, and end
	exprStart, exprEnd := expr.Span()
	if start.Line != exprStart.Line {
		return true
	}
	previousEnd := exprEnd
	for _, clause := range clauses {
		clauseStart, clauseEnd := clause.Span()
		if previousEnd.Line != clauseStart.Line {
			return true
		}
		previousEnd = clauseEnd
	}
	return previousEnd.Line != end.Line
}

// extractTrailingComments extracts trailing comments of an indented block starting with the first
// comment line with indentation less than the block indentation.
// The comments can either belong to CommentBlock statements or to the last non-comment statement
// as After-comments.
func extractTrailingComments(stmt Expr) []Expr {
	body := getLastBody(stmt)
	var comments []Expr
	if body != nil && len(*body) > 0 {
		// Get the current indentation level
		start, _ := (*body)[0].Span()
		indentation := start.LineRune

		// Find the last non-comment statement
		lastNonCommentIndex := -1
		for i, stmt := range *body {
			if _, ok := stmt.(*CommentBlock); !ok {
				lastNonCommentIndex = i
			}
		}
		if lastNonCommentIndex == -1 {
			return comments
		}

		// Iterate over the trailing comments, find the first comment line that's not indented enough,
		// dedent it and all the following comments.
		for i := lastNonCommentIndex; i < len(*body); i++ {
			stmt := (*body)[i]
			if comment := extractDedentedComment(stmt, indentation); comment != nil {
				// This comment and all the following CommentBlock statements are to be extracted.
				comments = append(comments, comment)
				comments = append(comments, (*body)[i+1:]...)
				*body = (*body)[:i+1]
				// If the current statement is a CommentBlock statement without any comment lines
				// it should be removed too.
				if i > lastNonCommentIndex && len(stmt.Comment().After) == 0 {
					*body = (*body)[:i]
				}
			}
		}
	}
	return comments
}

// extractDedentedComment extract the first comment line from `stmt` which indentation is smaller
// than `indentation`, and all following comment lines, and returns them in a newly created
// CommentBlock statement.
func extractDedentedComment(stmt Expr, indentation int) Expr {
	for i, line := range stmt.Comment().After {
		// line.Start.LineRune == 0 can't exist in parsed files, it indicates that the comment line
		// has been added by an AST modification. Don't take such lines into account.
		if line.Start.LineRune > 0 && line.Start.LineRune < indentation {
			// This and all the following lines should be dedented
			cb := &CommentBlock{
				Start:    line.Start,
				Comments: Comments{After: stmt.Comment().After[i:]},
			}
			stmt.Comment().After = stmt.Comment().After[:i]
			return cb
		}
	}
	return nil
}

// getLastBody returns the last body of a block statement (the only body for For- and DefStmt
// objects, the last in a if-elif-else chain
func getLastBody(stmt Expr) *[]Expr {
	switch block := stmt.(type) {
	case *DefStmt:
		return &block.Body
	case *ForStmt:
		return &block.Body
	case *IfStmt:
		if len(block.False) == 0 {
			return &block.True
		} else if len(block.False) == 1 {
			if next, ok := block.False[0].(*IfStmt); ok {
				// Recursively find the last block of the chain
				return getLastBody(next)
			}
		}
		return &block.False
	}
	return nil
}

//line yacctab:1
var yyExca = [...]int16{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 79,
	6, 55,
	-2, 128,
	-1, 169,
	20, 125,
	-2, 126,
}

const yyPrivate = 57344

const yyLast = 1052

var yyAct = [...]int16{
	20, 31, 234, 250, 146, 2, 29, 108, 186, 106,
	7, 147, 195, 95, 160, 43, 152, 9, 23, 187,
	159, 105, 240, 220, 174, 40, 87, 88, 89, 90,
	44, 84, 39, 36, 93, 98, 101, 49, 209, 54,
	131, 173, 53, 57, 83, 58, 103, 55, 113, 218,
	114, 39, 110, 36, 118, 119, 120, 121, 122, 123,
	124, 125, 126, 127, 128, 129, 130, 200, 132, 133,
	134, 135, 136, 137, 138, 139, 140, 35, 217, 56,
	238, 219, 189, 38, 51, 52, 143, 110, 92, 33,
	36, 34, 155, 156, 54, 85, 157, 53, 13, 164,
	76, 94, 55, 162, 36, 37, 163, 36, 39, 168,
	263, 171, 32, 48, 167, 109, 165, 212, 190, 116,
	36, 77, 39, 175, 100, 207, 181, 162, 213, 179,
	166, 86, 182, 47, 56, 54, 97, 162, 53, 57,
	117, 58, 201, 55, 111, 112, 248, 196, 188, 115,
	193, 247, 191, 197, 158, 205, 79, 194, 261, 206,
	84, 227, 78, 154, 211, 231, 47, 154, 80, 211,
	221, 214, 216, 204, 208, 56, 72, 73, 210, 149,
	208, 44, 47, 47, 223, 215, 102, 180, 45, 47,
	222, 245, 244, 142, 202, 196, 228, 229, 46, 232,
	233, 197, 225, 235, 151, 42, 260, 230, 178, 148,
	237, 226, 200, 47, 169, 153, 264, 224, 236, 192,
	172, 141, 91, 239, 104, 177, 1, 10, 243, 18,
	249, 241, 246, 188, 176, 242, 99, 96, 251, 253,
	41, 50, 19, 252, 12, 256, 257, 7, 8, 235,
	203, 258, 4, 30, 259, 161, 262, 150, 184, 185,
	81, 82, 251, 266, 265, 35, 144, 252, 27, 145,
	26, 38, 0, 0, 0, 0, 0, 33, 0, 34,
	0, 0, 0, 0, 28, 0, 0, 6, 0, 0,
	11, 0, 36, 37, 22, 0, 0, 0, 0, 24,
	32, 0, 0, 0, 0, 0, 0, 0, 25, 0,
	39, 21, 14, 15, 16, 17, 0, 254, 35, 5,
	0, 27, 0, 26, 38, 0, 0, 0, 0, 0,
	33, 0, 34, 0, 0, 0, 0, 28, 0, 0,
	6, 3, 0, 11, 0, 36, 37, 22, 0, 0,
	0, 54, 24, 32, 53, 57, 0, 58, 0, 55,
	0, 25, 0, 39, 21, 14, 15, 16, 17, 70,
	35, 0, 5, 27, 0, 26, 38, 0, 0, 0,
	0, 0, 33, 0, 34, 0, 0, 0, 0, 28,
	0, 56, 72, 73, 0, 0, 0, 36, 37, 0,
	0, 0, 0, 0, 24, 32, 0, 0, 0, 0,
	0, 0, 0, 25, 0, 39, 0, 14, 15, 16,
	17, 0, 54, 0, 107, 53, 57, 0, 58, 0,
	55, 0, 59, 255, 60, 0, 0, 0, 0, 69,
	70, 71, 0, 0, 68, 0, 0, 61, 0, 64,
	0, 0, 75, 0, 0, 65, 74, 0, 0, 62,
	63, 0, 56, 72, 73, 54, 66, 67, 53, 57,
	0, 58, 0, 55, 170, 59, 0, 60, 0, 0,
	0, 0, 69, 70, 71, 0, 0, 68, 0, 0,
	61, 0, 64, 0, 0, 75, 0, 0, 65, 74,
	0, 0, 62, 63, 0, 56, 72, 73, 54, 66,
	67, 53, 57, 0, 58, 0, 55, 0, 59, 0,
	60, 0, 0, 0, 0, 69, 70, 71, 0, 0,
	68, 0, 0, 61, 0, 64, 0, 0, 75, 183,
	0, 65, 74, 0, 0, 62, 63, 0, 56, 72,
	73, 54, 66, 67, 53, 57, 0, 58, 0, 55,
	0, 59, 0, 60, 0, 0, 0, 0, 69, 70,
	71, 0, 0, 68, 0, 0, 61, 162, 64, 0,
	0, 75, 0, 0, 65, 74, 0, 0, 62, 63,
	0, 56, 72, 73, 54, 66, 67, 53, 57, 0,
	58, 0, 55, 0, 59, 0, 60, 0, 0, 0,
	0, 69, 70, 71, 0, 0, 68, 0, 0, 61,
	0, 64, 0, 0, 75, 0, 0, 65, 74, 0,
	0, 62, 63, 0, 56, 72, 73, 35, 66, 67,
	27, 0, 26, 38, 0, 0, 0, 0, 0, 33,
	0, 34, 0, 0, 0, 0, 28, 0, 0, 0,
	0, 0, 0, 0, 36, 37, 0, 0, 0, 0,
	0, 24, 32, 0, 0, 0, 0, 0, 0, 0,
	25, 0, 39, 0, 14, 15, 16, 17, 54, 0,
	0, 53, 57, 0, 58, 0, 55, 0, 59, 0,
	60, 0, 0, 0, 0, 69, 70, 71, 0, 0,
	68, 0, 0, 61, 0, 64, 0, 0, 0, 0,
	0, 65, 74, 0, 0, 62, 63, 0, 56, 72,
	73, 54, 66, 67, 53, 57, 0, 58, 0, 55,
	0, 59, 0, 60, 0, 0, 0, 0, 69, 70,
	71, 0, 0, 68, 0, 0, 61, 0, 64, 0,
	0, 0, 0, 0, 65, 0, 0, 0, 62, 63,
	0, 56, 72, 73, 54, 66, 67, 53, 57, 0,
	58, 0, 55, 0, 59, 0, 60, 0, 0, 0,
	0, 69, 70, 71, 0, 0, 68, 0, 0, 61,
	0, 64, 0, 0, 0, 0, 0, 65, 0, 0,
	0, 62, 63, 0, 56, 72, 73, 54, 66, 0,
	53, 57, 0, 58, 0, 55, 0, 59, 0, 60,
	0, 0, 0, 0, 69, 70, 71, 0, 0, 0,
	0, 0, 61, 0, 64, 0, 0, 0, 0, 0,
	65, 0, 0, 0, 62, 63, 0, 56, 72, 73,
	35, 66, 198, 27, 200, 26, 38, 0, 0, 0,
	0, 0, 33, 0, 34, 0, 0, 0, 0, 28,
	0, 0, 0, 0, 0, 0, 0, 36, 37, 0,
	0, 0, 0, 54, 24, 32, 53, 57, 199, 58,
	0, 55, 0, 25, 35, 39, 198, 27, 0, 26,
	38, 70, 71, 0, 0, 0, 33, 0, 34, 0,
	0, 0, 0, 28, 0, 0, 0, 0, 0, 0,
	0, 36, 37, 56, 72, 73, 0, 0, 24, 32,
	35, 0, 199, 27, 200, 26, 38, 25, 0, 39,
	0, 0, 33, 0, 34, 0, 0, 0, 0, 28,
	0, 0, 0, 0, 0, 0, 0, 36, 37, 0,
	0, 0, 0, 0, 24, 32, 35, 0, 0, 27,
	0, 26, 38, 25, 0, 39, 0, 0, 33, 0,
	34, 0, 0, 0, 0, 28, 0, 0, 0, 0,
	0, 0, 0, 36, 37, 0, 0, 0, 0, 54,
	24, 32, 53, 57, 0, 58, 0, 55, 0, 25,
	0, 39, 0, 0, 0, 0, 69, 70, 71, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 56,
	72, 73,
}

var yyPact = [...]int16{
	-1000, -1000, 313, -1000, -1000, -1000, -34, -1000, -1000, -1000,
	192, 72, -1000, 173, 971, -1000, -1000, -1000, -10, 49,
	590, 68, 971, 151, 88, 971, 971, 971, 971, -1000,
	-1000, -1000, 217, 971, 971, 971, -1000, 175, 13, -1000,
	-1000, -41, 365, 78, 151, 971, 971, 971, 204, 971,
	971, 106, -1000, 971, 971, 971, 971, 971, 971, 971,
	971, 971, 971, 971, 971, 971, 3, 971, 971, 971,
	971, 971, 971, 971, 971, 971, 216, 180, 54, 200,
	971, 191, 206, -1000, 152, 21, 21, -1000, -1000, -1000,
	-1000, 200, 136, 547, 200, 73, 110, 205, 461, 200,
	214, 590, 8, -1000, -35, 632, -1000, -1000, -1000, 971,
	72, 204, 204, 684, 590, 174, 365, -1000, -1000, -1000,
	-1000, -1000, 90, 90, 1005, 1005, 1005, 1005, 1005, 1005,
	1005, 971, 770, 813, 889, 131, 347, 35, 35, 727,
	504, 75, 365, -1000, 213, 200, 899, 203, -1000, 124,
	181, 971, -1000, 88, 971, -1000, -1000, -18, -1000, 107,
	4, -1000, 72, 935, -1000, 97, -1000, 108, 935, -1000,
	971, 935, -1000, -1000, -1000, -1000, 22, -36, 157, 151,
	365, -1000, 1005, 971, 211, 202, -1000, -1000, 148, 21,
	21, -1000, -1000, -1000, 855, -1000, 590, 150, 971, 971,
	-1000, -1000, 971, -1000, -1000, 590, 200, -1000, 4, 971,
	43, 590, -1000, -1000, 590, -1000, 461, -1000, -37, -1000,
	-1000, 365, -1000, 684, -1000, -1000, 75, 971, 179, 178,
	-1000, 971, 590, 590, 133, 590, 58, 684, 971, 260,
	-1000, -1000, -1000, 418, 971, 971, 590, -1000, 971, 197,
	-1000, -1000, 143, 684, -1000, 971, 590, 590, 92, 210,
	1, -18, 590, -1000, -1000, -1000, -1000,
}

var yyPgo = [...]int16{
	0, 16, 11, 4, 12, 269, 266, 19, 261, 260,
	8, 259, 258, 0, 2, 88, 18, 98, 257, 101,
	15, 255, 14, 20, 6, 253, 5, 252, 248, 244,
	242, 241, 7, 17, 240, 13, 237, 236, 1, 9,
	234, 3, 230, 229, 227, 226, 225, 224,
}

var yyR1 = [...]int8{
	0, 45, 39, 39, 46, 46, 40, 40, 40, 26,
	26, 26, 26, 27, 27, 43, 44, 44, 28, 28,
	28, 30, 30, 29, 29, 31, 31, 32, 34, 34,
	33, 33, 33, 33, 33, 33, 33, 33, 47, 47,
	16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
	16, 16, 16, 16, 16, 6, 6, 5, 5, 4,
	4, 4, 4, 42, 42, 41, 41, 9, 9, 12,
	12, 8, 8, 11, 11, 7, 7, 7, 7, 7,
	10, 10, 10, 10, 10, 17, 17, 18, 18, 13,
	13, 13, 13, 13, 13, 13, 13, 13, 13, 13,
	13, 13, 13, 13, 13, 13, 13, 13, 13, 13,
	13, 13, 13, 13, 13, 13, 13, 13, 19, 19,
	14, 14, 15, 15, 1, 1, 2, 2, 3, 3,
	35, 37, 37, 36, 36, 36, 20, 20, 38, 24,
	25, 25, 25, 25, 21, 22, 22, 23, 23,
}

var yyR2 = [...]int8{
	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 5, 1, 3, 3, 6,
	1, 4, 5, 1, 4, 2, 1, 4, 0, 3,
	1, 2, 1, 3, 3, 1, 1, 1, 0, 1,
	1, 1, 1, 3, 8, 4, 4, 6, 8, 3,
	4, 4, 3, 4, 3, 0, 2, 2, 3, 1,
	3, 2, 2, 1, 3, 1, 3, 0, 2, 0,
	2, 1, 3, 1, 3, 1, 3, 2, 1, 2,
	1, 3, 5, 4, 4, 1, 3, 0, 1, 1,
	4, 2, 2, 2, 2, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 4, 3,
	3, 3, 3, 3, 3, 3, 3, 5, 1, 3,
	0, 1, 0, 2, 0, 1, 1, 2, 0, 1,
	3, 1, 3, 0, 1, 2, 1, 3, 1, 1,
	3, 2, 2, 1, 4, 1, 3, 1, 2,
}

var yyChk = [...]int16{
	-1000, -45, -26, 28, -27, 59, 27, -32, -28, -33,
	-44, 30, -29, -17, 52, 53, 54, 55, -43, -30,
	-13, 51, 34, -16, 39, 48, 10, 8, 24, -24,
	-25, -38, 40, 17, 19, 5, 32, 33, 11, 50,
	59, -34, 13, -20, -16, 15, 25, 9, -17, 47,
	-31, 35, 36, 7, 4, 12, 44, 8, 10, 14,
	16, 29, 41, 42, 31, 37, 48, 49, 26, 21,
	22, 23, 45, 46, 38, 34, 32, -17, 11, 5,
	17, -9, -8, -7, -24, 7, 43, -13, -13, -13,
	-13, 5, -15, -13, -19, -35, -36, -19, -13, -37,
	-15, -13, 11, 33, -47, 62, -39, 59, -32, 37,
	9, -17, -17, -13, -13, -17, 13, 34, -13, -13,
	-13, -13, -13, -13, -13, -13, -13, -13, -13, -13,
	-13, 37, -13, -13, -13, -13, -13, -13, -13, -13,
	-13, 5, 13, 32, -6, -5, -3, -2, 9, -17,
	-18, 13, -1, 9, 15, -24, -24, -3, 18, -23,
	-22, -21, 30, -2, -3, -23, 20, -1, -2, 9,
	13, -2, 6, 33, 59, -33, -40, -46, -17, -16,
	13, -39, -13, 35, -12, -11, -10, -7, -24, 7,
	43, -39, 6, -3, -2, -4, -13, -24, 7, 43,
	9, 18, 13, -17, -7, -13, -38, 18, -22, 34,
	-20, -13, 20, 20, -13, -35, -13, 56, 27, 59,
	59, 13, -39, -13, 6, -1, 9, 13, -24, -24,
	-4, 15, -13, -13, -14, -13, -2, -13, 37, -26,
	59, -39, -10, -13, 13, 13, -13, 18, 13, -42,
	-41, -38, -24, -13, 57, 15, -13, -13, -14, -3,
	9, 15, -13, 18, 6, -41, -38,
}

var yyDef = [...]int16{
	9, -2, 0, 1, 10, 11, 0, 13, 14, 28,
	0, 0, 20, 30, 32, 35, 36, 37, 16, 23,
	85, 0, 0, 89, 67, 0, 0, 0, 0, 40,
	41, 42, 0, 122, 133, 122, 139, 143, 0, 138,
	12, 38, 0, 0, 136, 0, 0, 0, 31, 0,
	0, 0, 26, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, -2,
	87, 0, 124, 71, 75, 78, 0, 91, 92, 93,
	94, 128, 0, 118, 128, 131, 0, 124, 118, 134,
	0, 118, 141, 142, 0, 39, 18, 6, 4, 0,
	0, 33, 34, 86, 17, 0, 0, 25, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 0, 109, 110, 111, 112, 113, 114, 115, 116,
	0, 69, 0, 43, 0, 128, 0, 129, 126, 88,
	0, 0, 68, 125, 0, 77, 79, 0, 49, 0,
	147, 145, 0, 129, 123, 0, 52, 0, 0, -2,
	0, 135, 54, 140, 27, 29, 0, 3, 0, 137,
	0, 24, 108, 0, 0, 124, 73, 80, 75, 78,
	0, 21, 45, 56, 129, 57, 59, 40, 0, 0,
	127, 46, 120, 90, 72, 76, 0, 50, 148, 0,
	0, 119, 51, 53, 130, 132, 0, 9, 0, 8,
	5, 0, 22, 117, 15, 70, 125, 0, 77, 79,
	58, 0, 61, 62, 0, 121, 0, 146, 0, 0,
	7, 19, 74, 81, 0, 0, 60, 47, 120, 128,
	63, 65, 0, 144, 2, 0, 83, 84, 0, 0,
	126, 0, 82, 48, 44, 64, 66,
}

var yyTok1 = [...]int8{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	59, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 22, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 62,
	14, 15, 16, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 17, 3, 18, 23, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 19, 21, 20, 24,
}

var yyTok2 = [...]int8{
	2, 3, 25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40, 41, 42,
	43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
	53, 54, 55, 56, 57, 58, 60, 61,
}

var yyTok3 = [...]int8{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int(yyPact[state])
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && int(yyChk[int(yyAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || int(yyExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := int(yyExca[i])
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int(yyTok1[0])
		goto out
	}
	if char < len(yyTok1) {
		token = int(yyTok1[char])
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = int(yyTok2[char-yyPrivate])
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = int(yyTok3[i+0])
		if token == char {
			token = int(yyTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(yyTok2[1]) /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = int(yyPact[yystate])
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = int(yyAct[yyn])
	if int(yyChk[yyn]) == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = int(yyDef[yystate])
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && int(yyExca[xi+1]) == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = int(yyExca[xi+0])
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = int(yyExca[xi+1])
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = int(yyPact[yyS[yyp].yys]) + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = int(yyAct[yyn]) /* simulate a shift of "error" */
					if int(yyChk[yystate]) == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= int(yyR2[yyn])
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = int(yyR1[yyn])
	yyg := int(yyPgo[yyn])
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = int(yyAct[yyg])
	} else {
		yystate = int(yyAct[yyj])
		if int(yyChk[yystate]) != -yyn {
			yystate = int(yyAct[yyg])
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:218
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:225
		{
			statements := yyDollar[4].exprs
			if yyDollar[2].exprs != nil {
				// $2 can only contain *CommentBlock objects, each of them contains a non-empty After slice
				cb := yyDollar[2].exprs[len(yyDollar[2].exprs)-1].(*CommentBlock)
				// $4 can't be empty and can't start with a comment
				stmt := yyDollar[4].exprs[0]
				start, _ := stmt.Span()
				if start.Line-cb.After[len(cb.After)-1].Start.Line == 1 {
					// The first statement of $4 starts on the next line after the last comment of $2.
					// Attach the last comment to the first statement
					stmt.Comment().Before = cb.After
					yyDollar[2].exprs = yyDollar[2].exprs[:len(yyDollar[2].exprs)-1]
				}
				statements = append(yyDollar[2].exprs, yyDollar[4].exprs...)
			}
			yyVAL.exprs = statements
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:245
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:253
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:258
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].lastStmt
			if yyVAL.lastStmt == nil {
				cb := &CommentBlock{Start: yyDollar[2].pos}
				yyVAL.exprs = append(yyVAL.exprs, cb)
				yyVAL.lastStmt = cb
			}
			com := yyVAL.lastStmt.Comment()
			com.After = append(com.After, Comment{Start: yyDollar[2].pos, Token: yyDollar[2].tok})
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:270
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:276
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:281
		{
			// If this statement follows a comment block,
			// attach the comments to the statement.
			if cb, ok := yyDollar[1].lastStmt.(*CommentBlock); ok {
				yyVAL.exprs = append(yyDollar[1].exprs[:len(yyDollar[1].exprs)-1], yyDollar[2].exprs...)
				yyDollar[2].exprs[0].Comment().Before = cb.After
				yyVAL.lastStmt = yyDollar[2].lastStmt
				break
			}

			// Otherwise add to list.
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
			yyVAL.lastStmt = yyDollar[2].lastStmt

			// Consider this input:
			//
			//	foo()
			//	# bar
			//	baz()
			//
			// If we've just parsed baz(), the # bar is attached to
			// foo() as an After comment. Make it a Before comment
			// for baz() instead.
			if x := yyDollar[1].lastStmt; x != nil {
				com := x.Comment()
				// stmt is never empty
				yyDollar[2].exprs[0].Comment().Before = com.After
				com.After = nil
			}
		}
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:312
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:318
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].lastStmt
			if yyVAL.lastStmt == nil {
				cb := &CommentBlock{Start: yyDollar[2].pos}
				yyVAL.exprs = append(yyVAL.exprs, cb)
				yyVAL.lastStmt = cb
			}
			com := yyVAL.lastStmt.Comment()
			com.After = append(com.After, Comment{Start: yyDollar[2].pos, Token: yyDollar[2].tok})
		}
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:332
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].exprs[len(yyDollar[1].exprs)-1]
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:337
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
			yyVAL.lastStmt = yyDollar[1].expr
			if cbs := extractTrailingComments(yyDollar[1].expr); len(cbs) > 0 {
				yyVAL.exprs = append(yyVAL.exprs, cbs...)
				yyVAL.lastStmt = cbs[len(cbs)-1]
				if yyDollar[1].lastStmt == nil {
					yyVAL.lastStmt = nil
				}
			}
		}
	case 15:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:351
		{
			yyVAL.def_header = &DefStmt{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[4].exprs,
				},
				Name:           yyDollar[2].tok,
				ForceCompact:   forceCompact(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
				ForceMultiLine: forceMultiLine(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
			}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:366
		{
			yyDollar[1].def_header.Type = yyDollar[3].expr
			yyVAL.def_header = yyDollar[1].def_header
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:373
		{
			yyDollar[1].def_header.Function.Body = yyDollar[3].exprs
			yyDollar[1].def_header.ColonPos = yyDollar[2].pos
			yyVAL.expr = yyDollar[1].def_header
			yyVAL.lastStmt = yyDollar[3].lastStmt
		}
	case 19:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:380
		{
			yyVAL.expr = &ForStmt{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				X:    yyDollar[4].expr,
				Body: yyDollar[6].exprs,
			}
			yyVAL.lastStmt = yyDollar[6].lastStmt
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:390
		{
			yyVAL.expr = yyDollar[1].ifstmt
			yyVAL.lastStmt = yyDollar[1].lastStmt
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:398
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].exprs,
			}
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 22:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:407
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = End{Pos: yyDollar[2].pos}
			inner.False = []Expr{
				&IfStmt{
					If:   yyDollar[2].pos,
					Cond: yyDollar[3].expr,
					True: yyDollar[5].exprs,
				},
			}
			yyVAL.lastStmt = yyDollar[5].lastStmt
		}
	case 24:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:428
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = End{Pos: yyDollar[2].pos}
			inner.False = yyDollar[4].exprs
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:445
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastStmt = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 28:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:451
		{
			yyVAL.exprs = []Expr{}
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:455
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:462
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:469
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:474
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:475
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 35:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:477
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 36:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:484
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:491
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 42:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:505
		{
			yyVAL.expr = yyDollar[1].string
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:509
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 44:
		yyDollar = yyS[yypt-8 : yypt+1]
//line build/parse.y:518
		{
			load := &LoadStmt{
				Load:         yyDollar[1].pos,
				Module:       yyDollar[4].string,
				Rparen:       End{Pos: yyDollar[8].pos},
				ForceCompact: yyDollar[2].pos.Line == yyDollar[8].pos.Line,
			}
			for _, arg := range yyDollar[6].loadargs {
				load.From = append(load.From, &arg.from)
				load.To = append(load.To, &arg.to)
			}
			yyVAL.expr = load
		}
	case 45:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:532
		{
			yyVAL.expr = &CallExpr{
				X:              yyDollar[1].expr,
				ListStart:      yyDollar[2].pos,
				List:           yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceCompact:   forceCompact(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
				ForceMultiLine: forceMultiLine(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:543
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:552
		{
			yyVAL.expr = &SliceExpr{
				X:          yyDollar[1].expr,
				SliceStart: yyDollar[2].pos,
				From:       yyDollar[3].expr,
				FirstColon: yyDollar[4].pos,
				To:         yyDollar[5].expr,
				End:        yyDollar[6].pos,
			}
		}
	case 48:
		yyDollar = yyS[yypt-8 : yypt+1]
//line build/parse.y:563
		{
			yyVAL.expr = &SliceExpr{
				X:           yyDollar[1].expr,
				SliceStart:  yyDollar[2].pos,
				From:        yyDollar[3].expr,
				FirstColon:  yyDollar[4].pos,
				To:          yyDollar[5].expr,
				SecondColon: yyDollar[6].pos,
				Step:        yyDollar[7].expr,
				End:         yyDollar[8].pos,
			}
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:576
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 50:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:585
		{
			yyVAL.expr = &Comprehension{
				Curly:          false,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].expr,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLineComprehension(yyDollar[1].pos, yyDollar[2].expr, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:596
		{
			yyVAL.expr = &Comprehension{
				Curly:          true,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].kv,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLineComprehension(yyDollar[1].pos, yyDollar[2].kv, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:607
		{
			exprValues := make([]Expr, 0, len(yyDollar[2].kvs))
			for _, kv := range yyDollar[2].kvs {
				exprValues = append(exprValues, Expr(kv))
			}
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].kvs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, exprValues, yyDollar[3].pos),
			}
		}
	case 53:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:620
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[4].pos),
			}
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:629
		{
			if len(yyDollar[2].exprs) == 1 && yyDollar[2].comma.Line == 0 {
				// Just a parenthesized expression, not a tuple.
				yyVAL.expr = &ParenExpr{
					Start:          yyDollar[1].pos,
					X:              yyDollar[2].exprs[0],
					End:            End{Pos: yyDollar[3].pos},
					ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
				}
			} else {
				yyVAL.expr = &TupleExpr{
					Start:          yyDollar[1].pos,
					List:           yyDollar[2].exprs,
					End:            End{Pos: yyDollar[3].pos},
					ForceCompact:   forceCompact(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
					ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
				}
			}
		}
	case 55:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:650
		{
			yyVAL.exprs = nil
		}
	case 56:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:654
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:660
		{
			yyVAL.exprs = []Expr{yyDollar[2].expr}
		}
	case 58:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:664
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 60:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:671
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 61:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:675
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:679
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:684
		{
			yyVAL.loadargs = []*struct {
				from Ident
				to   Ident
			}{yyDollar[1].loadarg}
		}
	case 64:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:688
		{
			yyDollar[1].loadargs = append(yyDollar[1].loadargs, yyDollar[3].loadarg)
			yyVAL.loadargs = yyDollar[1].loadargs
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:694
		{
			start := yyDollar[1].string.Start.add("'")
			if yyDollar[1].string.TripleQuote {
				start = start.add("''")
			}
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: start,
				},
				to: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: start,
				},
			}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:711
		{
			start := yyDollar[3].string.Start.add("'")
			if yyDollar[3].string.TripleQuote {
				start = start.add("''")
			}
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[3].string.Value,
					NamePos: start,
				},
				to: *yyDollar[1].expr.(*Ident),
			}
		}
	case 67:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:726
		{
			yyVAL.exprs = nil
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:730
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 69:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:735
		{
			yyVAL.exprs = nil
		}
	case 70:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:739
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:745
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:749
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 73:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:756
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 74:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:760
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 76:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:767
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:771
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 78:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:775
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, nil)
		}
	case 79:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:779
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:788
		{
			yyVAL.expr = typed(yyDollar[1].expr, yyDollar[3].expr)
		}
	case 82:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:792
		{
			yyVAL.expr = binary(typed(yyDollar[1].expr, yyDollar[3].expr), yyDollar[4].pos, yyDollar[4].tok, yyDollar[5].expr)
		}
	case 83:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:796
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, typed(yyDollar[2].expr, yyDollar[4].expr))
		}
	case 84:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:800
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, typed(yyDollar[2].expr, yyDollar[4].expr))
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:807
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.NoBrackets {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					NoBrackets:     true,
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 87:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:822
		{
			yyVAL.expr = nil
		}
	case 90:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:830
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:839
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:840
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:841
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:842
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:843
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:844
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:845
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:846
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:847
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:848
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:849
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:850
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:851
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:852
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 105:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:853
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:854
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 107:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:855
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 108:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:856
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 109:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:857
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:858
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:859
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 112:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:860
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 113:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:861
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 114:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:862
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 115:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:863
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 116:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:865
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 117:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:873
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:885
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:889
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 120:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:894
		{
			yyVAL.expr = nil
		}
	case 122:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:900
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 123:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:904
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 124:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:914
		{
			yyVAL.pos = Position{}
		}
	case 127:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:925
		{
			yyVAL.pos = yyDollar[1].pos
		}
	case 128:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:933
		{
			yyVAL.pos = Position{}
		}
	case 130:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:940
		{
			yyVAL.kv = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:950
		{
			yyVAL.kvs = []*KeyValueExpr{yyDollar[1].kv}
		}
	case 132:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:954
		{
			yyVAL.kvs = append(yyDollar[1].kvs, yyDollar[3].kv)
		}
	case 133:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:959
		{
			yyVAL.kvs = nil
		}
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:963
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 135:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:967
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 137:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:974
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.NoBrackets {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					NoBrackets:     true,
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:990
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1002
		{
			yyVAL.expr = &Ident{NamePos: yyDollar[1].pos, Name: yyDollar[1].tok}
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:1008
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "." + yyDollar[3].tok}
		}
	case 141:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:1012
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "."}
		}
	case 142:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:1016
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: "." + yyDollar[2].tok}
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1020
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 144:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:1026
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 145:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1037
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 146:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:1041
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 147:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1050
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 148:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:1054
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
