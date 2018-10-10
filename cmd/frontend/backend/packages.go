package backend

import (
	"context"
	"fmt"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/proxy"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Packages contains backend methods related to code packages.
var Packages packages

type packages struct{}

// RefreshIndex refreshes the global packages index for the specified repo@commit.
func (packages) RefreshIndex(ctx context.Context, repo *types.Repo, commitID api.CommitID) error {
	langs, err := languagesForRepo(ctx, repo, commitID)
	if err != nil {
		return err
	}
	var errs []string
	for _, lang := range langs {
		pkgs, err := (packages{}).listForLanguageInRepo(ctx, lang, repo, commitID, true)
		if err == nil {
			err = db.Pkgs.UpdateIndexForLanguage(ctx, lang, repo.ID, pkgs)
		}
		if err != nil && !proxy.IsModeNotFound(err) {
			log15.Error("Refreshing repository packages index failed.", "repo", repo.URI, "language", lang, "error", err)
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

func (packages) listForLanguageInRepo(ctx context.Context, language string, repo *types.Repo, commitID api.CommitID, background bool) (pkgs []lspext.PackageInformation, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "listForLanguageInRepo "+language+" "+string(repo.URI))
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	vcs := "git" // TODO: store VCS type in *types.Repo object.

	// Query all external packages for the repository. If background is true, we do this using the
	// "<language>_bg" mode which runs this request on a separate language
	// server explicitly for background tasks such as workspace/xpackages.
	// This makes it such that indexing repositories does not interfere in
	// terms of resource usage with real user requests.
	if _, ok := xlang.HasXDefinitionAndXPackages[language]; !ok { // TODO(keegancsmith) this list is no longer accurate
		// The language does not support xpackages, so there is no indexing to
		// perform.
		return nil, nil
	}
	var bgSuffix string
	if background {
		bgSuffix = "_bg"
	}
	rootURI := lsp.DocumentURI(vcs + "://" + string(repo.URI) + "?" + string(commitID))
	var allPks []lspext.PackageInformation
	err = cachedUnsafeXLangCall(ctx, language+bgSuffix, rootURI, "workspace/xpackages", map[string]string{}, &allPks)
	if err != nil {
		return nil, errors.Wrap(err, "LSP Call workspace/xpackages")
	}

	// Filter out vendored packages (don't want them treated as canonical sources)
	pks := make([]lspext.PackageInformation, 0, len(allPks))
	for _, pk := range allPks {
		if baseDir, hasBaseDir := pk.Package["baseDir"]; hasBaseDir {
			if baseDir, isStr := baseDir.(string); isStr && strings.Contains(baseDir, "/vendor") {
				continue
			}
		}
		pks = append(pks, pk)
	}
	return pks, nil
}

// List returns the repository's packages at the given revision (by collecting the
// workspace/xpackages LSP results from all relevant language servers).
//
// To retrieve the cached results from a recent default branch commit of the repository, use
// db.Pkgs.ListPackages instead.
func (packages) List(ctx context.Context, repo *types.Repo, rev api.CommitID) ([]*api.PackageInfo, error) {
	if Mocks.Packages.List != nil {
		return Mocks.Packages.List(repo, rev)
	}

	langs, err := languagesForRepo(ctx, repo, rev)
	if err != nil {
		return nil, err
	}

	var allPkgs []*api.PackageInfo
	for _, lang := range langs {
		pkgs, err := (packages{}).listForLanguageInRepo(ctx, lang, repo, rev, false)
		if err != nil {
			if proxy.IsModeNotFound(err) {
				log15.Debug("Packages.List skipping language because no language server is registered", "lang", lang, "err", err)
			} else {
				return nil, errors.Wrap(err, "listForLanguageInRepo "+lang)
			}
		}
		for _, pkg := range pkgs {
			allPkgs = append(allPkgs, &api.PackageInfo{
				Lang:         lang,
				RepoID:       repo.ID,
				Pkg:          pkg.Package,
				Dependencies: pkg.Dependencies,
			})
		}
	}
	return allPkgs, nil
}

// MockPackages allows mocking of Packages backend methods (by setting Mocks.Packages's
// fields).
type MockPackages struct {
	List func(repo *types.Repo, rev api.CommitID) ([]*api.PackageInfo, error)
}
