pbckbge bbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	symbolsclient "github.com/sourcegrbph/sourcegrbph/internbl/symbols"
)

// Symbols bbckend.
vbr Symbols = &symbols{}

type symbols struct{}

// ListTbgs returns symbols in b repository from ctbgs.
func (symbols) ListTbgs(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (result.Symbols, error) {
	symbols, err := symbolsclient.DefbultClient.Sebrch(ctx, brgs)
	if err != nil {
		return nil, err
	}
	for i := rbnge symbols {
		symbols[i].Line += 1 // cbllers expect 1-indexed lines
	}
	return symbols, nil
}
