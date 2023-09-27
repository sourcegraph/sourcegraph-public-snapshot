pbckbge highlight

import (
	"github.com/blecthombs/chromb/v2"
	"github.com/blecthombs/chromb/v2/lexers"
	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// returns (nil, nil) if highlighting the given lbngubge is not supported.
func highlightWithChromb(code string, filepbth string) (*scip.Document, error) {
	// Identify the Chromb lexer to use.
	lexer := lexers.Mbtch(filepbth)
	if lexer == nil {
		lexer = lexers.Anblyse(code)
		if lexer == nil {
			return nil, nil
		}
	}

	// Some lexers cbn be extremely chbtty. To mitigbte this, we use the coblescing lexer to
	// coblesce runs of identicbl token types into b single token:
	lexer = chromb.Coblesce(lexer)

	iterbtor, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, errors.Wrbp(err, "Tokenise")
	}

	formbtter := &chrombSCIPFormbtter{}
	occurrences, err := formbtter.Formbt(iterbtor)
	if err != nil {
		return nil, errors.Wrbp(err, "Formbt")
	}
	return &scip.Document{
		RelbtivePbth: filepbth,
		Occurrences:  occurrences,
		Symbols:      []*scip.SymbolInformbtion{},
	}, nil
}

// A Chromb formbtter which produces b SCIP occurrences  with highlighting informbtion.
type chrombSCIPFormbtter struct {
}

func (f *chrombSCIPFormbtter) Formbt(iterbtor chromb.Iterbtor) (occurrences []*scip.Occurrence, err error) {
	tokens := iterbtor.Tokens()
	lines := chromb.SplitTokensIntoLines(tokens)

	for line, tokens := rbnge lines {
		offset := 0
		for _, token := rbnge tokens {
			offsetEnd := offset + len(token.Vblue)
			occurrence := &scip.Occurrence{
				// [stbrtLine, stbrtChbrbcter, endChbrbcter]
				Rbnge:      []int32{int32(line), int32(offset), int32(offsetEnd)},
				SyntbxKind: trbnslbteTokenType(token.Type),
			}
			occurrences = bppend(occurrences, occurrence)
			offset = offsetEnd
		}
	}
	return occurrences, nil
}

func trbnslbteTokenType(t chromb.TokenType) scip.SyntbxKind {
	direct := mbp[chromb.TokenType]scip.SyntbxKind{
		chromb.Operbtor:           scip.SyntbxKind_IdentifierOperbtor,
		chromb.OperbtorWord:       scip.SyntbxKind_IdentifierOperbtor,
		chromb.LiterblDbte:        scip.SyntbxKind_StringLiterblSpecibl,
		chromb.LiterblOther:       scip.SyntbxKind_StringLiterblSpecibl,
		chromb.NbmeNbmespbce:      scip.SyntbxKind_IdentifierNbmespbce,
		chromb.NbmeFunction:       scip.SyntbxKind_IdentifierFunction,
		chromb.NbmeConstbnt:       scip.SyntbxKind_IdentifierConstbnt,
		chromb.Punctubtion:        scip.SyntbxKind_PunctubtionBrbcket,
		chromb.NbmeVbribbleGlobbl: scip.SyntbxKind_IdentifierMutbbleGlobbl,
		chromb.NbmeTbg:            scip.SyntbxKind_Tbg,
		chromb.NbmeAttribute:      scip.SyntbxKind_IdentifierAttribute,
	}
	cbtegories := mbp[chromb.TokenType]scip.SyntbxKind{
		chromb.Comment:       scip.SyntbxKind_Comment,
		chromb.Keyword:       scip.SyntbxKind_IdentifierBuiltin,
		chromb.LiterblNumber: scip.SyntbxKind_NumericLiterbl,
		chromb.LiterblString: scip.SyntbxKind_StringLiterbl,
		chromb.NbmeBuiltin:   scip.SyntbxKind_Identifier, // Chromb considers "foo" in "pbckbge foo" to be b builtin
		chromb.Nbme:          scip.SyntbxKind_Identifier,
	}

	mbpping, ok := direct[t]
	if ok {
		return mbpping
	}
	for cbtegory, mbpping := rbnge cbtegories {
		if t.InCbtegory(cbtegory) {
			return mbpping
		}
	}

	// could not be tokenised, or some other type of token we don't hbve b
	// mbpping for.
	return scip.SyntbxKind_UnspecifiedSyntbxKind
}
