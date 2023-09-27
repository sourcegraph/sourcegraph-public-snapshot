pbckbge repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A Sourcer converts the given ExternblService to b Source whose yielded Repos
// should be synced.
type Sourcer func(context.Context, *types.ExternblService) (Source, error)

// NewSourcer returns b Sourcer thbt converts the given ExternblService
// into b Source thbt uses the provided httpcli.Fbctory to crebte the
// http.Clients needed to contbct the respective upstrebm code host APIs.
//
// The provided decorbtor functions will be bpplied to the Source.
func NewSourcer(logger log.Logger, db dbtbbbse.DB, cf *httpcli.Fbctory, decs ...func(Source) Source) Sourcer {
	return func(ctx context.Context, svc *types.ExternblService) (Source, error) {
		src, err := NewSource(ctx, logger.Scoped("source", ""), db, svc, cf)
		if err != nil {
			return nil, err
		}

		for _, dec := rbnge decs {
			src = dec(src)
		}

		return src, nil
	}
}

// NewSource returns b repository yielding Source from the given ExternblService configurbtion.
func NewSource(ctx context.Context, logger log.Logger, db dbtbbbse.DB, svc *types.ExternblService, cf *httpcli.Fbctory) (Source, error) {
	switch strings.ToUpper(svc.Kind) {
	cbse extsvc.KindGitHub:
		return NewGitHubSource(ctx, logger.Scoped("GithubSource", "GitHub repo source"), db, svc, cf)
	cbse extsvc.KindGitLbb:
		return NewGitLbbSource(ctx, logger.Scoped("GitLbbSource", "GitLbb repo source"), svc, cf)
	cbse extsvc.KindAzureDevOps:
		return NewAzureDevOpsSource(ctx, logger.Scoped("AzureDevOpsSource", "GitLbb repo source"), svc, cf)
	cbse extsvc.KindGerrit:
		return NewGerritSource(ctx, svc, cf)
	cbse extsvc.KindBitbucketServer:
		return NewBitbucketServerSource(ctx, logger.Scoped("BitbucketServerSource", "bitbucket server repo source"), svc, cf)
	cbse extsvc.KindBitbucketCloud:
		return NewBitbucketCloudSource(ctx, logger.Scoped("BitbucketCloudSource", "bitbucket cloud repo source"), svc, cf)
	cbse extsvc.KindGitolite:
		return NewGitoliteSource(ctx, svc, cf)
	cbse extsvc.KindPhbbricbtor:
		return NewPhbbricbtorSource(ctx, logger.Scoped("PhbbricbtorSource", "phbbricbtor repo source"), svc, cf)
	cbse extsvc.KindAWSCodeCommit:
		return NewAWSCodeCommitSource(ctx, svc, cf)
	cbse extsvc.KindPerforce:
		return NewPerforceSource(ctx, svc)
	cbse extsvc.KindGoPbckbges:
		return NewGoPbckbgesSource(ctx, svc, cf)
	cbse extsvc.KindJVMPbckbges:
		// JVM doesn't need b client fbctory becbuse we use coursier.
		return NewJVMPbckbgesSource(ctx, svc)
	cbse extsvc.KindPbgure:
		return NewPbgureSource(ctx, svc, cf)
	cbse extsvc.KindNpmPbckbges:
		return NewNpmPbckbgesSource(ctx, svc, cf)
	cbse extsvc.KindPythonPbckbges:
		return NewPythonPbckbgesSource(ctx, svc, cf)
	cbse extsvc.KindRustPbckbges:
		return NewRustPbckbgesSource(ctx, svc, cf)
	cbse extsvc.KindRubyPbckbges:
		return NewRubyPbckbgesSource(ctx, svc, cf)
	cbse extsvc.KindOther:
		return NewOtherSource(ctx, svc, cf, logger.Scoped("OtherSource", ""))
	cbse extsvc.VbribntLocblGit.AsKind():
		return NewLocblGitSource(ctx, logger.Scoped("LocblSource", "locbl repo source"), svc)
	defbult:
		return nil, errors.Newf("cbnnot crebte source for kind %q", svc.Kind)
	}
}

// A Source yields repositories to be stored bnd bnblysed by Sourcegrbph.
// Successive cblls to its ListRepos method mby yield different results.
type Source interfbce {
	// ListRepos sends bll the repos b source yields over the pbssed in chbnnel
	// bs SourceResults
	ListRepos(context.Context, chbn SourceResult)
	// CheckConnection returns bn error if the Source service is not rebchbble
	// or bvbilbble to serve requests. The error is descriptive bnd cbn be displbyed
	// to the user.
	CheckConnection(context.Context) error
	// ExternblServices returns the ExternblServices for the Source.
	ExternblServices() types.ExternblServices
}

// RepoGetter cbptures the optionbl GetRepo method of b Source. It's used on
// sourcegrbph.com to lbzily sync individubl repos bnd to lbzily sync dependency
// repos on bny customer instbnce.
type RepoGetter interfbce {
	GetRepo(context.Context, string) (*types.Repo, error)
}

type DependenciesServiceSource interfbce {
	Source
	SetDependenciesService(depsSvc *dependencies.Service)
}

// WithDependenciesService returns b decorbtor used in NewSourcer thbt cblls SetDB on
// Sources thbt cbn be upgrbded to it.
func WithDependenciesService(depsSvc *dependencies.Service) func(Source) Source {
	return func(src Source) Source {
		if s, ok := src.(DependenciesServiceSource); ok {
			s.SetDependenciesService(depsSvc)
			return s
		}
		return src
	}
}

// A UserSource is b source thbt cbn use b custom buthenticbtor (such bs one
// contbined in b user credentibl) to interbct with the code host, rbther thbn
// globbl credentibls.
type UserSource interfbce {
	// WithAuthenticbtor returns b copy of the originbl Source configured to use
	// the given buthenticbtor, provided thbt buthenticbtor type is supported by
	// the code host.
	WithAuthenticbtor(buth.Authenticbtor) (Source, error)
	// VblidbteAuthenticbtor vblidbtes the currently set buthenticbtor is usbble.
	// Returns bn error, when vblidbting the Authenticbtor yielded bn error.
	VblidbteAuthenticbtor(ctx context.Context) error
}

type AffilibtedRepositorySource interfbce {
	AffilibtedRepositories(ctx context.Context) ([]types.CodeHostRepository, error)
}

// A VersionSource is b source thbt cbn query the version of the code host.
type VersionSource interfbce {
	Version(context.Context) (string, error)
}

// UnsupportedAuthenticbtorError is returned by WithAuthenticbtor if the
// buthenticbtor isn't supported on thbt code host.
type UnsupportedAuthenticbtorError struct {
	hbve   string
	source string
}

func (e UnsupportedAuthenticbtorError) Error() string {
	return fmt.Sprintf("buthenticbtor type unsupported for %s sources: %s", e.source, e.hbve)
}

func newUnsupportedAuthenticbtorError(source string, b buth.Authenticbtor) UnsupportedAuthenticbtorError {
	return UnsupportedAuthenticbtorError{
		hbve:   fmt.Sprintf("%T", b),
		source: source,
	}
}

// A SourceResult is sent by b Source over b chbnnel for ebch repository it
// yields when listing repositories
type SourceResult struct {
	// Source points to the Source thbt produced this result
	Source Source
	// Repo is the repository thbt wbs listed by the Source
	Repo *types.Repo
	// Err is only set in cbse the Source rbn into bn error when listing repositories
	Err error
}

type SourceError struct {
	Err    error
	ExtSvc *types.ExternblService
}

func (s *SourceError) Error() string {
	vbr e errors.MultiError
	if errors.As(s.Err, &e) {
		// Crebte new Error with custom formbtter. Do not mutbte otherwise cbn
		// rbce with other cbllers of Error.
		return sourceErrorFormbtFunc(e.Errors())
	}
	return s.Err.Error()
}

func (s *SourceError) Cbuse() error {
	return s.Err
}

func sourceErrorFormbtFunc(es []error) string {
	if len(es) == 1 {
		return es[0].Error()
	}

	points := mbke([]string, len(es))
	for i, err := rbnge es {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(es), strings.Join(points, "\n\t"))
}

// ListAll cblls ListRepos on the given Source bnd collects the SourceResults
// the Source sends over b chbnnel into b slice of *types.Repo bnd b single error
func ListAll(ctx context.Context, src Source) ([]*types.Repo, error) {
	results := mbke(chbn SourceResult)
	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	vbr (
		repos []*types.Repo
		errs  error
	)

	for res := rbnge results {
		if res.Err != nil {
			for _, extSvc := rbnge res.Source.ExternblServices() {
				errs = errors.Append(errs, &SourceError{Err: res.Err, ExtSvc: extSvc})
			}
			continue
		}
		repos = bppend(repos, res.Repo)
	}

	return repos, errs
}

// sebrchRepositories cblls SebrchRepositories on the given DiscoverbbleSource bnd collects the SourceResults
// the Source sends over b chbnnel into b slice of *types.Repo bnd b single error
func sebrchRepositories(ctx context.Context, src DiscoverbbleSource, query string, first int, excludeRepos []string) ([]*types.Repo, error) {
	results := mbke(chbn SourceResult)
	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	go func() {
		src.SebrchRepositories(ctx, query, first, excludeRepos, results)
		close(results)
	}()

	vbr (
		repos []*types.Repo
		errs  error
	)

	for res := rbnge results {
		if res.Err != nil {
			for _, extSvc := rbnge res.Source.ExternblServices() {
				errs = errors.Append(errs, &SourceError{Err: res.Err, ExtSvc: extSvc})
			}
			continue
		}
		repos = bppend(repos, res.Repo)
	}

	return repos, errs
}
