package backend

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

// xlangCall invokes the xlang method with the specified
// arguments. This exists as an intermediary between this package and
// xlang.UnsafeOneShotClientRequest to enable mocking of xlang in unit
// tests.
var unsafeXLangCall = unsafeXLangCall_

func unsafeXLangCall_(ctx context.Context, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error {
	return xlang.UnsafeOneShotClientRequest(ctx, mode, rootURI, method, params, results)
}

func mockXLang(fn func(ctx context.Context, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error) (done func()) {
	unsafeXLangCall = fn
	return func() {
		unsafeXLangCall = unsafeXLangCall_
	}
}

var xlangSupportedLanguages = map[string]struct{}{
	"go":         struct{}{},
	"php":        struct{}{},
	"typescript": struct{}{},
	"javascript": struct{}{},
	"java":       struct{}{},
	"python":     struct{}{},
}

func languagesForRepo(ctx context.Context, repo *types.Repo, commitID api.CommitID) (languages []string, err error) {
	inv, err := Repos.GetInventory(ctx, repo, commitID)
	if err != nil {
		return nil, errors.Wrap(err, "Repos.GetInventory")
	}
	for _, lang := range inv.Languages {
		langName := strings.ToLower(lang.Name)
		if _, enabled := xlangSupportedLanguages[langName]; enabled {
			languages = append(languages, langName)
		}
	}
	return languages, nil
}
