package highlight

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// returns (nil, nil) if highlighting the given language is not supported.
func highlightWithChroma(code string, filepath string) (*scip.Document, error) {
	// Identify the Chroma lexer to use.
	lexer := lexers.Match(filepath)
	if lexer == nil {
		lexer = lexers.Analyse(code)
		if lexer == nil {
			return nil, nil
		}
	}

	// Some lexers can be extremely chatty. To mitigate this, we use the coalescing lexer to
	// coalesce runs of identical token types into a single token:
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, errors.Wrap(err, "Tokenise")
	}

	formatter := &chromaSCIPFormatter{}
	occurrences, err := formatter.Format(iterator)
	if err != nil {
		return nil, errors.Wrap(err, "Format")
	}
	return &scip.Document{
		RelativePath: filepath,
		Occurrences:  occurrences,
		Symbols:      []*scip.SymbolInformation{},
	}, nil
}

// A Chroma formatter which produces a SCIP occurrences  with highlighting information.
type chromaSCIPFormatter struct {
}

func (f *chromaSCIPFormatter) Format(iterator chroma.Iterator) (occurrences []*scip.Occurrence, err error) {
	tokens := iterator.Tokens()
	lines := chroma.SplitTokensIntoLines(tokens)

	for line, tokens := range lines {
		offset := 0
		for _, token := range tokens {
			offsetEnd := offset + len(token.Value)
			occurrence := &scip.Occurrence{
				// [startLine, startCharacter, endCharacter]
				Range:      []int32{int32(line), int32(offset), int32(offsetEnd)},
				SyntaxKind: translateTokenType(token.Type),
			}
			occurrences = append(occurrences, occurrence)
			offset = offsetEnd
		}
	}
	return occurrences, nil
}

func translateTokenType(t chroma.TokenType) scip.SyntaxKind {
	direct := map[chroma.TokenType]scip.SyntaxKind{
		chroma.Operator:           scip.SyntaxKind_IdentifierOperator,
		chroma.OperatorWord:       scip.SyntaxKind_IdentifierOperator,
		chroma.LiteralDate:        scip.SyntaxKind_StringLiteralSpecial,
		chroma.LiteralOther:       scip.SyntaxKind_StringLiteralSpecial,
		chroma.NameNamespace:      scip.SyntaxKind_IdentifierNamespace,
		chroma.NameFunction:       scip.SyntaxKind_IdentifierFunction,
		chroma.NameConstant:       scip.SyntaxKind_IdentifierConstant,
		chroma.Punctuation:        scip.SyntaxKind_PunctuationBracket,
		chroma.NameVariableGlobal: scip.SyntaxKind_IdentifierMutableGlobal,
		chroma.NameTag:            scip.SyntaxKind_Tag,
		chroma.NameAttribute:      scip.SyntaxKind_IdentifierAttribute,
	}
	categories := map[chroma.TokenType]scip.SyntaxKind{
		chroma.Comment:       scip.SyntaxKind_Comment,
		chroma.Keyword:       scip.SyntaxKind_IdentifierBuiltin,
		chroma.LiteralNumber: scip.SyntaxKind_NumericLiteral,
		chroma.LiteralString: scip.SyntaxKind_StringLiteral,
		chroma.NameBuiltin:   scip.SyntaxKind_Identifier, // Chroma considers "foo" in "package foo" to be a builtin
		chroma.Name:          scip.SyntaxKind_Identifier,
	}

	mapping, ok := direct[t]
	if ok {
		return mapping
	}
	for category, mapping := range categories {
		if t.InCategory(category) {
			return mapping
		}
	}

	// could not be tokenised, or some other type of token we don't have a
	// mapping for.
	return scip.SyntaxKind_UnspecifiedSyntaxKind
}
