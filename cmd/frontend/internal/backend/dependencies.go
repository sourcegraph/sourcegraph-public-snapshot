package backend

import (
	"context"
	"fmt"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
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
		deps, err := (dependencies{}).listForLanguageInRepo(ctx, lang, repo, commitID)
		if err == nil {
			err = db.GlobalDeps.UpdateIndexForLanguage(ctx, lang, repo, deps)
		}
		if err != nil {
			log15.Error("refreshing index failed", "language", lang, "error", err)
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

func (dependencies) listForLanguageInRepo(ctx context.Context, language string, repo *types.Repo, commitID api.CommitID) (deps []lspext.DependencyReference, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "listForLanguageInRepo "+language+" "+string(repo.URI))
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	vcs := "git" // TODO: store VCS type in *types.Repo object.

	// Query all external dependencies for the repository. We do this using the
	// "<language>_bg" mode which runs this request on a separate language
	// server explicitly for background tasks such as workspace/xdependencies.
	// This makes it such that indexing repositories does not interfere in
	// terms of resource usage with real user requests.
	rootURI := lsp.DocumentURI(vcs + "://" + string(repo.URI) + "?" + string(commitID))
	err = unsafeXLangCall(ctx, language+"_bg", rootURI, "workspace/xdependencies", map[string]string{}, &deps)
	if err != nil {
		return nil, errors.Wrap(err, "LSP Call workspace/xdependencies")
	}
	return deps, nil
}

// List returns the repository's dependencies at the given revision (by collecting the
// workspace/xdependencies LSP results from all relevant language servers).
//
// To retrieve the cached results from a recent default branch commit of the repository, use
// db.GlobalDeps.Dependencies instead.
func (dependencies) List(ctx context.Context, repo *types.Repo, rev api.CommitID) ([]*api.DependencyReference, error) {
	if Mocks.Dependencies.List != nil {
		return Mocks.Dependencies.List(repo, rev)
	}

	langs, err := languagesForRepo(ctx, repo, rev)
	if err != nil {
		return nil, err
	}

	var allDeps []*api.DependencyReference
	for _, lang := range langs {
		deps, err := (dependencies{}).listForLanguageInRepo(ctx, lang, repo, rev)
		if err != nil {
			return nil, errors.Wrap(err, "listForLanguageInRepo "+lang)
		}
		for _, dep := range deps {
			allDeps = append(allDeps, &api.DependencyReference{
				Language: lang,
				RepoID:   repo.ID,
				DepData:  dep.Attributes,
				Hints:    dep.Hints,
			})
		}
	}
	return allDeps, nil
}

// MockDependencies allows mocking of Dependencies backend methods (by setting Mocks.Dependencies's
// fields).
type MockDependencies struct {
	List func(repo *types.Repo, rev api.CommitID) ([]*api.DependencyReference, error)
}
