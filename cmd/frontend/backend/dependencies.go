package backend

import (
	"context"
	"fmt"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	xlang_lspext "github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/proxy"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Dependencies contains backend methods related to code dependencies.
var Dependencies dependencies

type dependencies struct{}

// RefreshIndex refreshes the global deps index for the specified repo@commit.
func (dependencies) RefreshIndex(ctx context.Context, repo *types.Repo, commitID api.CommitID) error {
	langs, err := languagesForRepo(ctx, repo, commitID)
	if err != nil {
		return err
	}
	var errs []string
	for _, lang := range langs {
		deps, err := (dependencies{}).listForLanguageInRepo(ctx, lang, repo, commitID, true)
		if err == nil {
			err = db.GlobalDeps.UpdateIndexForLanguage(ctx, lang, repo.ID, deps)
		}
		if err != nil && !proxy.IsModeNotFound(err) {
			log15.Error("Refreshing repository dependencies index failed.", "repo", repo.Name, "language", lang, "error", err)
			errs = append(errs, fmt.Sprintf("refreshing index failed language=%s error=%v", lang, err))
		}
	}
	if len(errs) == 1 {
		return errors.New(errs[0])
	} else if len(errs) > 1 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func (dependencies) listForLanguageInRepo(ctx context.Context, language string, repo *types.Repo, commitID api.CommitID, background bool) (deps []xlang_lspext.DependencyReference, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "listForLanguageInRepo "+language+" "+string(repo.Name))
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	vcs := "git" // TODO: store VCS type in *types.Repo object.

	// Query all external dependencies for the repository. If background is true, we do
	// this using the "<language>_bg" mode which runs this request on a separate language
	// server explicitly for background tasks such as workspace/xdependencies.  This makes
	// it such that indexing repositories does not interfere in terms of resource usage
	// with real user requests.
	var bgSuffix string
	if background {
		bgSuffix = "_bg"
	}
	rootURI := lsp.DocumentURI(vcs + "://" + string(repo.Name) + "?" + string(commitID))
	err = cachedUnsafeXLangCall(ctx, language+bgSuffix, rootURI, "workspace/xdependencies", map[string]string{}, &deps)
	if err != nil {
		return nil, errors.Wrap(err, "LSP Call workspace/xdependencies")
	}
	return deps, nil
}
