package externallink

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// Repository returns the external links for a repository.
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
func Repository(ctx context.Context, db database.DB, repo *types.Repo) (links []*Resolver, err error) {
	phabRepo, link, serviceType := linksForRepository(ctx, db, repo)
	if phabRepo != nil {
		links = append(links, NewResolver(
			strings.TrimSuffix(phabRepo.URL, "/")+"/diffusion/"+phabRepo.Callsign,
			extsvc.TypePhabricator,
		))
	}
	if link != nil && link.Root != "" {
		links = append(links, NewResolver(link.Root, serviceType))
	}
	return links, nil
}

// FileOrDir returns the external links for a file or directory in a repository.
func FileOrDir(ctx context.Context, db database.DB, repo *types.Repo, rev, path string, isDir bool) (links []*Resolver, err error) {
	rev = url.PathEscape(rev)

	phabRepo, link, serviceType := linksForRepository(ctx, db, repo)
	if phabRepo != nil {
		// We need a branch name to construct the Phabricator URL.
		branchName, _, err := git.GetDefaultBranchShort(ctx, db, repo.Name)
		if err == nil && branchName != "" {
			links = append(links, NewResolver(
				fmt.Sprintf("%s/source/%s/browse/%s/%s;%s", strings.TrimSuffix(phabRepo.URL, "/"), phabRepo.Callsign, url.PathEscape(branchName), path, rev),
				extsvc.TypePhabricator,
			))
		}
	}

	if link != nil {
		var url string
		if isDir {
			url = link.Tree
		} else {
			url = link.Blob
		}
		if url != "" {
			url = strings.NewReplacer("{rev}", rev, "{path}", path).Replace(url)
			links = append(links, NewResolver(url, serviceType))
		}
	}

	return links, nil
}

// Commit returns the external links for a commit in a repository.
func Commit(ctx context.Context, db database.DB, repo *types.Repo, commitID api.CommitID) (links []*Resolver, err error) {
	commitStr := url.PathEscape(string(commitID))

	phabRepo, link, serviceType := linksForRepository(ctx, db, repo)
	if phabRepo != nil {
		links = append(links, NewResolver(
			fmt.Sprintf("%s/r%s%s", strings.TrimSuffix(phabRepo.URL, "/"), phabRepo.Callsign, commitStr),
			extsvc.TypePhabricator,
		))
	}

	if link != nil && link.Commit != "" {
		links = append(links, NewResolver(
			strings.ReplaceAll(link.Commit, "{commit}", commitStr),
			serviceType,
		))
	}

	return links, nil
}

// linksForRepository gets the information necessary to construct links to resources within this
// repository.
//
// It logs errors to the trace but does not return errors, because external links are not worth
// failing any request for.
func linksForRepository(
	ctx context.Context,
	db database.DB,
	repo *types.Repo,
) (phabRepo *types.PhabricatorRepo, links *protocol.RepoLinks, serviceType string) {
	span, ctx := ot.StartSpanFromContext(ctx, "externallink.linksForRepository")
	defer span.Finish()
	span.SetTag("Repo", repo.Name)
	span.SetTag("ExternalRepo", repo.ExternalRepo)

	var err error
	phabRepo, err = db.Phabricator().GetByName(ctx, repo.Name)
	if err != nil && !errcode.IsNotFound(err) {
		ext.Error.Set(span, true)
		span.SetTag("phabErr", err.Error())
	}

	repoInfo := protocol.NewRepoInfo(repo)

	return phabRepo, repoInfo.Links, repoInfo.ExternalRepo.ServiceType
}
