package externallink

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// NewRepositoryLinker gets the information necessary to construct links to resources within this
// repository.
//
// It logs errors to the trace but does not return errors, because external links are not worth
// failing any request for.
func NewRepositoryLinker(
	ctx context.Context,
	db database.DB,
	repo *types.Repo,
	defaultBranch string,
) RepositoryLinker {
	tr, ctx := trace.New(ctx, "linksForRepository",
		repo.Name.Attr(),
		attribute.Stringer("externalRepo", repo.ExternalRepo))
	defer tr.End()

	phabRepo, err := db.Phabricator().GetByName(ctx, repo.Name)
	if err != nil && !errcode.IsNotFound(err) {
		tr.SetError(err)
	}

	repoInfo := protocol.NewRepoInfo(repo)

	return RepositoryLinker{
		phabRepo:      phabRepo,
		links:         repoInfo.Links,
		serviceType:   repoInfo.ExternalRepo.ServiceType,
		defaultBranch: defaultBranch,
	}
}

type RepositoryLinker struct {
	phabRepo      *types.PhabricatorRepo
	links         *protocol.RepoLinks
	serviceType   string
	defaultBranch string
}

// Repository returns the external links for a repository.
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
func (rl *RepositoryLinker) Repository() (links []*Resolver) {
	if rl.phabRepo != nil {
		links = append(links, NewResolver(
			strings.TrimSuffix(rl.phabRepo.URL, "/")+"/diffusion/"+rl.phabRepo.Callsign,
			extsvc.TypePhabricator,
		))
	}
	if rl.links != nil && rl.links.Root != "" {
		links = append(links, NewResolver(rl.links.Root, rl.serviceType))
	}
	return links
}

// FileOrDir returns the external links for a file or directory in a repository.
func (rl *RepositoryLinker) FileOrDir(rev, path string, isDir bool) (links []*Resolver) {
	rev = url.PathEscape(rev)

	if rl.phabRepo != nil {
		// We need a branch name to construct the Phabricator URL.
		links = append(links, NewResolver(
			fmt.Sprintf(
				"%s/source/%s/browse/%s/%s;%s",
				strings.TrimSuffix(rl.phabRepo.URL, "/"),
				rl.phabRepo.Callsign,
				url.PathEscape(rl.defaultBranch),
				path,
				rev,
			),
			extsvc.TypePhabricator,
		))
	}

	if rl.links != nil {
		var urlStr string
		if isDir {
			urlStr = rl.links.Tree
		} else {
			urlStr = rl.links.Blob
		}
		if urlStr != "" {
			urlStr = strings.NewReplacer("{rev}", rev, "{path}", path).Replace(urlStr)
			links = append(links, NewResolver(urlStr, rl.serviceType))
		}
	}

	return links
}

// Commit returns the external links for a commit in a repository.
func (rl *RepositoryLinker) Commit(commitID api.CommitID) (links []*Resolver) {
	commitStr := url.PathEscape(string(commitID))

	if rl.phabRepo != nil {
		links = append(links, NewResolver(
			fmt.Sprintf(
				"%s/r%s%s",
				strings.TrimSuffix(rl.phabRepo.URL, "/"),
				rl.phabRepo.Callsign,
				commitStr,
			),
			extsvc.TypePhabricator,
		))
	}

	if rl.links != nil && rl.links.Commit != "" {
		links = append(links, NewResolver(
			strings.ReplaceAll(rl.links.Commit, "{commit}", commitStr),
			rl.serviceType,
		))
	}

	return links
}
