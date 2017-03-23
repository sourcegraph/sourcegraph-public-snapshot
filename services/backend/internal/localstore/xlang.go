package localstore

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

// xlangCall invokes the xlang method with the specified
// arguments. This exists as an intermediary between this package and
// xlang.UnsafeOneShotClientRequest to enable mocking of xlang in unit
// tests.
var unsafeXLangCall = unsafeXLangCall_

func unsafeXLangCall_(ctx context.Context, mode, rootPath, method string, params, results interface{}) error {
	return xlang.UnsafeOneShotClientRequest(ctx, mode, rootPath, method, params, results)
}

func mockXLang(fn func(ctx context.Context, mode, rootPath, method string, params, results interface{}) error) (done func()) {
	unsafeXLangCall = fn
	return func() {
		unsafeXLangCall = unsafeXLangCall_
	}
}
