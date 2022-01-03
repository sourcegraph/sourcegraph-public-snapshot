package highlight

import (
	"context"
	"fmt"
	"html"
	"io"
	"math"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/cockroachdb/errors"
	tree_sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_go "github.com/smacker/go-tree-sitter/golang"
)

const (
	LANG_GO = "go"
)

var (
	SUPPORTED_LANGUAGES = []string{LANG_GO}
	LANGUAGE_MAP        = map[string]*tree_sitter.Language{
		LANG_GO: tree_sitter_go.GetLanguage(),
	}
)

type Highlighting struct {
	prefix []byte
	suffix []byte
}

var (
	HL_SOURCE_FILE = Highlighting{[]byte("<pre>"), []byte("</pre>")}

	HL_COMMENT = Highlighting{[]byte("<span class=\"comment\">"), []byte("</span>")}

	HL_KEYWORD = Highlighting{[]byte("<span class=\"keyword\">"), []byte("</span>")}

	// TODO: We should have cross-language names that work here.
	// The key bit is *mutable* global variables.
	HL_LOCAL_IDENTIFIER          = Highlighting{[]byte("<span class=\"local-identifier\">"), []byte("</span>")}
	HL_MUTABLE_GLOBAL_IDENTIFIER = Highlighting{[]byte("<span class=\"global-identifier\""), []byte("</span>")}
	HL_CONST_GLOBAL_IDENTIFIER   = Highlighting{[]byte("<span class=\"const-identifier\">"), []byte("</span>")}
	HL_FUNCTION_IDENTIFIER       = Highlighting{[]byte("<span class=\"function-identifier\">"), []byte("</span>")}
	HL_PARAMETER_IDENTIFIER      = Highlighting{[]byte("<span class=\"parameter-identifier\">"), []byte("</span>")}
	HL_FIELD_IDENTIFIER          = Highlighting{[]byte("<span class=\"field-identifier\">"), []byte("</span>")}
	HL_TYPE_IDENTIFIER           = Highlighting{[]byte("<span class=\"type-identifier\">"), []byte("</span>")}

	HL_STRING_LITERAL        = Highlighting{[]byte("<span class=\"string-literal\">"), []byte("</span>")}
	HL_RAW_STRING_LITERAL    = Highlighting{[]byte("<span class=\"raw-string-literal\">"), []byte("</span>")}
	HL_NIL_LITERAL           = Highlighting{[]byte("<span class=\"nil-literal\">"), []byte("</span>")}
	HL_STRING_LITERAL_ESCAPE = Highlighting{[]byte("<span class=\"string-literal-escape\">"), []byte("</span>")}
	HL_INTEGER_LITERAL       = Highlighting{[]byte("<span class=\"integer-literal\">"), []byte("</span>")}
	HL_FLOAT_LITERAL         = Highlighting{[]byte("<span class=\"float-literal\">"), []byte("</span>")}
	HL_CHARACTER_LITERAL     = Highlighting{[]byte("<span class=\"character-literal\">"), []byte("</span>")}
	HL_BOOLEAN_LITERAL       = Highlighting{[]byte("<span class=\"boolean-literal\">"), []byte("</span>")}
)

type HighlightingContext struct {
	context       context.Context
	input         []byte
	output        DirectWriter
	grammar       *GoGrammar
	lastByteIndex uint32
}

func NewHighlightingContext(context context.Context, input []byte, output DirectWriter, lang *tree_sitter.Language) HighlightingContext {
	if lang.SymbolCount() > math.MaxUint16 {
		panic("Number of kinds of symbols in language should be < MaxUint16")
	}
	symbolMap := map[string]uint16{}
	for i := uint16(0); i < uint16(lang.SymbolCount()); i++ {
		symbolMap[lang.SymbolName(tree_sitter.Symbol(i))] = i
	}
	var g GoGrammar
	gref := reflect.ValueOf(&g).Elem()
	for _, field := range reflect.VisibleFields(reflect.TypeOf(g)) {
		snakeCaseFieldName := convertCamelCaseToSnakeCase(field.Name)
		f := gref.FieldByName(field.Name)
		f.Set(reflect.ValueOf(tree_sitter.Symbol(symbolMap[snakeCaseFieldName])))
	}
	return HighlightingContext{
		context, input, output, &g, 0,
	}
}

/*
 */

func convertCamelCaseToSnakeCase(camelCase string) string {
	var builder strings.Builder
	for i, c := range camelCase {
		if unicode.IsUpper(c) {
			if i != 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(unicode.ToLower(c))
		} else {
			builder.WriteRune(c)
		}
	}
	return builder.String()
}

type GoGrammar struct {
	Package                      tree_sitter.Symbol
	SourceFile                   tree_sitter.Symbol
	Comment                      tree_sitter.Symbol
	BlankIdentifier              tree_sitter.Symbol
	Identifier                   tree_sitter.Symbol
	FunctionDeclaration          tree_sitter.Symbol
	ParameterDeclaration         tree_sitter.Symbol
	VariadicParameterDeclaration tree_sitter.Symbol
	VarDeclaration               tree_sitter.Symbol
	ConstDeclaration             tree_sitter.Symbol
	TypeIdentifier               tree_sitter.Symbol
	FieldIdentifier              tree_sitter.Symbol
	PackageIdentifier            tree_sitter.Symbol
	CallExpression               tree_sitter.Symbol
	InterpretedStringLiteral     tree_sitter.Symbol
	RawStringLiteral             tree_sitter.Symbol
	Nil                          tree_sitter.Symbol
	True                         tree_sitter.Symbol
	False                        tree_sitter.Symbol
	IntLiteral                   tree_sitter.Symbol
	RuneLiteral                  tree_sitter.Symbol
	FloatLiteral                 tree_sitter.Symbol
	EscapeSequence               tree_sitter.Symbol
}

type DirectWriter interface {
	io.Writer
	io.ReaderFrom
}

var _ TreeVisitor = &HighlightingContext{}

func (h *HighlightingContext) startVisit(node *tree_sitter.Node) error {
	if h.context != nil && h.context.Err() != nil {
		// TODO: Better error reporting
		return fmt.Errorf("timed out while trying to highlight node pos=%v text=\"%s\"", node.StartPoint(), h.contentOf(node))
	}
	if node.StartByte() > h.lastByteIndex {
		if err := h.writeEscaped(string(h.input[h.lastByteIndex:node.StartByte()])); err != nil {
			return err
		}
		h.lastByteIndex = node.StartByte()
	}
	// TODO: We should profile how much time is taken up by switching on strings... 😬
	g := h.grammar
	switch node.Symbol() {
	case g.SourceFile:
		return h.writeBytesWithoutEscaping(HL_SOURCE_FILE.prefix)
	case g.Comment:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_COMMENT)
	case g.BlankIdentifier, g.PackageIdentifier:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_LOCAL_IDENTIFIER)
	case g.Identifier:
		if err := checkLeaf(node); err != nil {
			return err
		}
		parent := node.Parent()
		var around Highlighting
		if parent == nil {
			around = HL_LOCAL_IDENTIFIER
		} else {
			switch parent.Symbol() {
			case g.FunctionDeclaration:
				around = HL_FUNCTION_IDENTIFIER
			case g.ParameterDeclaration, g.VariadicParameterDeclaration:
				around = HL_PARAMETER_IDENTIFIER
			case g.VarDeclaration:
				grandparent := parent.Parent()
				if grandparent != nil && grandparent.Symbol() == g.SourceFile {
					around = HL_MUTABLE_GLOBAL_IDENTIFIER
				} else {
					around = HL_LOCAL_IDENTIFIER
				}
			case g.ConstDeclaration:
				around = HL_CONST_GLOBAL_IDENTIFIER
			default:
				around = HL_LOCAL_IDENTIFIER
			}
		}
		return h.highlightLeaf(node, around)
	case g.TypeIdentifier:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_TYPE_IDENTIFIER)
	case g.FieldIdentifier:
		if err := checkLeaf(node); err != nil {
			return err
		}
		if parent := node.Parent(); parent != nil {
			if grandparent := parent.Parent(); grandparent != nil && grandparent.Symbol() == g.CallExpression {
				return h.highlightLeaf(node, HL_FUNCTION_IDENTIFIER)
			}
		}
		// This will highlight un-applied references to methods as fields, not
		// as methods, but we can't do the "right thing" without type-based
		// name lookup.
		return h.highlightLeaf(node, HL_FIELD_IDENTIFIER)
	case g.InterpretedStringLiteral:
		// string literals may have escapes in-between
		return h.writeBytesWithoutEscaping(HL_STRING_LITERAL.prefix)
	case g.RawStringLiteral:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_RAW_STRING_LITERAL)
	case g.Nil:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_NIL_LITERAL)
	case g.True, g.False:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_BOOLEAN_LITERAL)
	case g.IntLiteral:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_INTEGER_LITERAL)
	case g.RuneLiteral:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_CHARACTER_LITERAL)
	case g.FloatLiteral:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_FLOAT_LITERAL)
	case g.EscapeSequence:
		if err := checkLeaf(node); err != nil {
			return err
		}
		return h.highlightLeaf(node, HL_STRING_LITERAL_ESCAPE)
	default:
		return nil
	}
}

func (h *HighlightingContext) afterVisitingChild(node *tree_sitter.Node, visitedChildIndex int) error {
	return nil
}

func (h *HighlightingContext) endVisit(node *tree_sitter.Node) error {
	h.lastByteIndex = node.EndByte()
	switch node.Symbol() {
	case h.grammar.SourceFile:
		return h.writeBytesWithoutEscaping(HL_SOURCE_FILE.suffix)
	case h.grammar.InterpretedStringLiteral:
		return h.writeBytesWithoutEscaping(HL_STRING_LITERAL.suffix)
	}
	return nil
}

var (
	keywordRegexp = regexp.MustCompile(
		"break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var",
	)
)

func (h *HighlightingContext) visitAnonymous(node *tree_sitter.Node) error {
	if node.StartByte() > h.lastByteIndex {
		if err := h.writeEscaped(string(h.input[h.lastByteIndex:node.StartByte()])); err != nil {
			return err
		}
		h.lastByteIndex = node.StartByte()
	}
	// TODO: tree-sitter sometimes leaves empty nodes at the end of a file.
	// For example, https://sourcegraph.com/github.com/kubernetes/kubernetes/-/blob/vendor/github.com/modern-go/concurrent/log.go
	if int(node.StartByte()) == len(h.input) {
		return nil
	}
	firstByte := h.input[node.StartByte()]
	// https://go.dev/ref/spec#Keywords
	if firstByte < 'b' || firstByte > 'v' {
		return h.writeContent(node)
	}
	content := h.contentOf(node)
	matchResult := keywordRegexp.FindStringIndex(content)
	if matchResult == nil || matchResult[0] != 0 || matchResult[1] != len(content) {
		return h.writeContent(node)
	}
	if err := checkLeaf(node); err != nil {
		return err
	}
	return h.highlightLeaf(node, HL_KEYWORD)
}

// Helper functions and methods

func highlightingFailError(err error, node *tree_sitter.Node) error {
	pt := node.StartPoint()
	// Use +1 for errors since tree-sitter uses 0-based numbering for rows and columns.
	return errors.Wrapf(err, "failed to highlight node type=%s pos=%d:%d", node.Type(), pt.Row+1, pt.Column+1)
}

func (h *HighlightingContext) highlightLeaf(node *tree_sitter.Node, around Highlighting) error {
	if err := h.writeBytesWithoutEscaping(around.prefix); err != nil {
		return err
	}
	if err := h.writeContent(node); err != nil {
		return highlightingFailError(err, node)
	}
	return h.writeBytesWithoutEscaping(around.suffix)
}

func checkLeaf(node *tree_sitter.Node) error {
	if node.ChildCount() != 0 {
		err := fmt.Errorf("expected node to be leaf but found: child_count=%d", node.ChildCount())
		return highlightingFailError(err, node)
	}
	return nil
}

// Helper method to get the content of a node.
func (h *HighlightingContext) contentOf(node *tree_sitter.Node) string {
	// NOTE: This triggers a needless allocation + memcpy.
	return node.Content(h.input)
}

func (h *HighlightingContext) writeContent(node *tree_sitter.Node) error {
	err := h.writeEscaped(h.contentOf(node))
	h.lastByteIndex = node.EndByte()
	return err
}

func (h *HighlightingContext) writeContentBeforeChild(node *tree_sitter.Node, nextChildIndex uint32) error {
	if nextChildIndex > node.ChildCount() {
		panic(fmt.Errorf("out-of-range index: node=[%d:%d-%d:%d] nextChildIndex=%d, childCount=%d",
			node.StartPoint().Row+1, node.StartPoint().Column+1,
			node.EndPoint().Row+1, node.EndPoint().Column+1,
			nextChildIndex, node.ChildCount()))
	}
	if node.ChildCount() == 0 {
		return h.writeContent(node)
	}
	var start uint32
	var end uint32
	if nextChildIndex == 0 {
		start = node.StartByte()
		end = node.Child(0).StartByte()
	} else if nextChildIndex == node.ChildCount() {
		start = node.Child(int(nextChildIndex - 1)).EndByte()
		end = node.EndByte()
	} else {
		start = node.Child(int(nextChildIndex) - 1).EndByte()
		end = node.Child(int(nextChildIndex)).StartByte()
	}
	return h.writeEscaped(string(h.input[start:end]))
}

func (h *HighlightingContext) writeEscaped(s string) error {
	htmlText := html.EscapeString(s)
	reader := strings.NewReader(htmlText)
	written, err := io.CopyN(h.output, reader, int64(len(htmlText)))
	if err != nil {
		return errors.Wrapf(err, "failed to fully copy content for node copied_bytes=%d", written)
	}
	return nil
}

func (h *HighlightingContext) writeBytesWithoutEscaping(bytes []byte) error {
	nwritten, err := h.output.Write(bytes)
	if err != nil {
		return errors.Wrapf(err, "failed to write bytes nwritten=%d", nwritten)
	}
	return nil
}
