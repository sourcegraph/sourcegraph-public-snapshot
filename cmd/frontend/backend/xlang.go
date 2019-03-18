package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/xlang"
)

// cachedUnsafeXLangCall invokes the xlang method with the specified
// arguments. This exists as an intermediary between this package and
// xlang.UnsafeOneShotClientRequest to enable mocking of xlang in unit
// tests.
var cachedUnsafeXLangCall = cachedUnsafeXLangCall_

var xlangCache = rcache.NewWithTTL("backend-xlang", 600)

func cachedUnsafeXLangCall_(ctx context.Context, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error {
	key := fmt.Sprintf("%s:%s:%s:%+v", mode, rootURI, method, params)
	cacheable := !strings.HasSuffix(mode, "_bg") // don't cache _bg requests because hit rate will be very low
	if cacheable {
		b, ok := xlangCache.Get(key)
		if ok {
			return json.Unmarshal(b, results)
		}
	}

	if err := xlang.UnsafeOneShotClientRequest(ctx, mode, rootURI, method, params, results); err != nil {
		return err
	}
	if cacheable {
		b, err := json.Marshal(results)
		if err != nil {
			return err
		}
		xlangCache.Set(key, b)
	}
	return nil
}

func mockXLang(fn func(ctx context.Context, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error) (done func()) {
	cachedUnsafeXLangCall = fn
	return func() {
		cachedUnsafeXLangCall = cachedUnsafeXLangCall_
	}
}

var xlangSupportedLanguages = map[string]struct{}{
	"go":         {},
	"php":        {},
	"typescript": {},
	"javascript": {},
	"java":       {},
	"python":     {},
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

// IsLanguageSupported returns true if we have LSP-based
// code intelligence support for the given language and the language server is configured, false otherwise
func IsLanguageSupported(lang string) bool {
	for _, langserver := range conf.EnabledLangservers() {
		if langserver.Language == lang {
			return true
		}
	}
	return false
}
