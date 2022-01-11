package types

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type SearchFunc func(ctx context.Context, args SearchArgs) (_ *result.Symbols, err error)
