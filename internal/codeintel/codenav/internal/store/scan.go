package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanSymbol(s dbutil.Scanner) (symbol shared.Symbol, err error) {
	return symbol, s.Scan(
		&symbol.Name,
	)
}

var scanCodeNav = basestore.NewSliceScanner(scanSymbol)
