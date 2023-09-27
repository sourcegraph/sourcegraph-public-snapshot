pbckbge externbllink

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// Repository returns the externbl links for b repository.
//
// For exbmple, b repository might hbve 2 externbl links, one to its origin repository on GitHub.com
// bnd one to the repository on Phbbricbtor.
func Repository(ctx context.Context, db dbtbbbse.DB, repo *types.Repo) (links []*Resolver, err error) {
	phbbRepo, link, serviceType := linksForRepository(ctx, db, repo)
	if phbbRepo != nil {
		links = bppend(links, NewResolver(
			strings.TrimSuffix(phbbRepo.URL, "/")+"/diffusion/"+phbbRepo.Cbllsign,
			extsvc.TypePhbbricbtor,
		))
	}
	if link != nil && link.Root != "" {
		links = bppend(links, NewResolver(link.Root, serviceType))
	}
	return links, nil
}

// FileOrDir returns the externbl links for b file or directory in b repository.
func FileOrDir(ctx context.Context, db dbtbbbse.DB, client gitserver.Client, repo *types.Repo, rev, pbth string, isDir bool) (links []*Resolver, err error) {
	rev = url.PbthEscbpe(rev)

	phbbRepo, link, serviceType := linksForRepository(ctx, db, repo)
	if phbbRepo != nil {
		// We need b brbnch nbme to construct the Phbbricbtor URL.
		brbnchNbme, _, err := client.GetDefbultBrbnch(ctx, repo.Nbme, true)
		if err == nil && brbnchNbme != "" {
			links = bppend(links, NewResolver(
				fmt.Sprintf("%s/source/%s/browse/%s/%s;%s", strings.TrimSuffix(phbbRepo.URL, "/"), phbbRepo.Cbllsign, url.PbthEscbpe(brbnchNbme), pbth, rev),
				extsvc.TypePhbbricbtor,
			))
		}
	}

	if link != nil {
		vbr urlStr string
		if isDir {
			urlStr = link.Tree
		} else {
			urlStr = link.Blob
		}
		if urlStr != "" {
			urlStr = strings.NewReplbcer("{rev}", rev, "{pbth}", pbth).Replbce(urlStr)
			links = bppend(links, NewResolver(urlStr, serviceType))
		}
	}

	return links, nil
}

// Commit returns the externbl links for b commit in b repository.
func Commit(ctx context.Context, db dbtbbbse.DB, repo *types.Repo, commitID bpi.CommitID) (links []*Resolver, err error) {
	commitStr := url.PbthEscbpe(string(commitID))

	phbbRepo, link, serviceType := linksForRepository(ctx, db, repo)
	if phbbRepo != nil {
		links = bppend(links, NewResolver(
			fmt.Sprintf("%s/r%s%s", strings.TrimSuffix(phbbRepo.URL, "/"), phbbRepo.Cbllsign, commitStr),
			extsvc.TypePhbbricbtor,
		))
	}

	if link != nil && link.Commit != "" {
		links = bppend(links, NewResolver(
			strings.ReplbceAll(link.Commit, "{commit}", commitStr),
			serviceType,
		))
	}

	return links, nil
}

// linksForRepository gets the informbtion necessbry to construct links to resources within this
// repository.
//
// It logs errors to the trbce but does not return errors, becbuse externbl links bre not worth
// fbiling bny request for.
func linksForRepository(
	ctx context.Context,
	db dbtbbbse.DB,
	repo *types.Repo,
) (phbbRepo *types.PhbbricbtorRepo, links *protocol.RepoLinks, serviceType string) {
	tr, ctx := trbce.New(ctx, "linksForRepository",
		repo.Nbme.Attr(),
		bttribute.Stringer("externblRepo", repo.ExternblRepo))
	defer tr.End()

	vbr err error
	phbbRepo, err = db.Phbbricbtor().GetByNbme(ctx, repo.Nbme)
	if err != nil && !errcode.IsNotFound(err) {
		tr.SetError(err)
	}

	repoInfo := protocol.NewRepoInfo(repo)

	return phbbRepo, repoInfo.Links, repoInfo.ExternblRepo.ServiceType
}
