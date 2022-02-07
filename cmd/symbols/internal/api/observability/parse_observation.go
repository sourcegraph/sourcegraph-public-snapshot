package observability

import (
	"context"
)

type (
	parseAmountKey int
	parseAmount    string
)

const (
	parseAmntKey parseAmountKey = iota

	FullParse    parseAmount = "full-parse"
	PartialParse parseAmount = "partial-parse"
	CachedParse  parseAmount = "cached-parse"
)

func SeedParseAmount(ctx context.Context) context.Context {
	// we use a pointer so that we can replace the value by dereferencing
	// further down the callstack
	amount := new(parseAmount)
	return context.WithValue(ctx, parseAmntKey, amount)
}

func SetParseAmount(ctx context.Context, amount parseAmount) {
	if amnt, ok := ctx.Value(parseAmntKey).(*parseAmount); ok {
		*amnt = amount
		return
	}
}

func GetParseAmount(ctx context.Context) string {
	if amnt, ok := ctx.Value(parseAmntKey).(*parseAmount); ok && amnt != nil {
		return string(*amnt)
	}
	return ""
}
