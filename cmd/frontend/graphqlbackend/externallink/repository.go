package externallink

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// Repository returns the external links for a repository.
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
func Repository(ctx context.Context, db dbutil.DB, repo *types.Repo) (links []*Resolver, err error) {
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
func FileOrDir(ctx context.Context, db dbutil.DB, repo *types.Repo, rev, path string, isDir bool) (links []*Resolver, err error) {
	rev = url.PathEscape(rev)

	phabRepo, link, serviceType := linksForRepository(ctx, db, repo)
	if phabRepo != nil {
		// We need a branch name to construct the Phabricator URL.
		branchName, _, _, err := git.ExecSafe(ctx, repo.Name, []string{"symbolic-ref", "--short", "HEAD"})
		branchName = bytes.TrimSpace(branchName)
		if err == nil && string(branchName) != "" {
			links = append(links, NewResolver(
				fmt.Sprintf("%s/source/%s/browse/%s/%s;%s", strings.TrimSuffix(phabRepo.URL, "/"), phabRepo.Callsign, url.PathEscape(string(branchName)), path, rev),
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
func Commit(ctx context.Context, db dbutil.DB, repo *types.Repo, commitID api.CommitID) (links []*Resolver, err error) {
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
	db dbutil.DB,
	repo *types.Repo,
) (phabRepo *types.PhabricatorRepo, links *protocol.RepoLinks, serviceType string) {
	span, ctx := ot.StartSpanFromContext(ctx, "externallink.linksForRepository")
	defer span.Finish()
	span.SetTag("Repo", repo.Name)
	span.SetTag("ExternalRepo", repo.ExternalRepo)

	var err error
	phabRepo, err = database.Phabricator(db).GetByName(ctx, repo.Name)
	if err != nil && !errcode.IsNotFound(err) {
		ext.Error.Set(span, true)
		span.SetTag("phabErr", err.Error())
	}

	typ, _ := extsvc.ParseServiceType(repo.ExternalRepo.ServiceType)
	switch typ {
	case extsvc.TypeGitHub:
		ghrepo := repo.Metadata.(*github.Repository)
		links = &protocol.RepoLinks{
			Root:   ghrepo.URL,
			Tree:   pathAppend(ghrepo.URL, "/tree/{rev}/{path}"),
			Blob:   pathAppend(ghrepo.URL, "/blob/{rev}/{path}"),
			Commit: pathAppend(ghrepo.URL, "/commit/{commit}"),
		}
	case extsvc.TypeGitLab:
		proj := repo.Metadata.(*gitlab.Project)
		links = &protocol.RepoLinks{
			Root:   proj.WebURL,
			Tree:   pathAppend(proj.WebURL, "/tree/{rev}/{path}"),
			Blob:   pathAppend(proj.WebURL, "/blob/{rev}/{path}"),
			Commit: pathAppend(proj.WebURL, "/commit/{commit}"),
		}
	case extsvc.TypeBitbucketServer:
		repo := repo.Metadata.(*bitbucketserver.Repo)
		if len(repo.Links.Self) == 0 {
			break
		}

		href := repo.Links.Self[0].Href
		root := strings.TrimSuffix(href, "/browse")
		links = &protocol.RepoLinks{
			Root:   href,
			Tree:   pathAppend(root, "/browse/{path}?at={rev}"),
			Blob:   pathAppend(root, "/browse/{path}?at={rev}"),
			Commit: pathAppend(root, "/commits/{commit}"),
		}
	case extsvc.TypeAWSCodeCommit:
		repo := repo.Metadata.(*awscodecommit.Repository)
		if repo.ARN == "" {
			break
		}

		splittedARN := strings.Split(strings.TrimPrefix(repo.ARN, "arn:aws:codecommit:"), ":")
		if len(splittedARN) == 0 {
			break
		}
		region := splittedARN[0]
		webURL := fmt.Sprintf(
			"https://%s.console.aws.amazon.com/codesuite/codecommit/repositories/%s",
			region,
			repo.Name,
		)
		links = &protocol.RepoLinks{
			Root:   webURL + "/browse",
			Tree:   webURL + "/browse/{rev}/--/{path}",
			Blob:   webURL + "/browse/{rev}/--/{path}",
			Commit: webURL + "/commit/{commit}",
		}
	}

	return phabRepo, links, repo.ExternalRepo.ServiceType
}

func pathAppend(base, p string) string {
	return strings.TrimRight(base, "/") + p
}
