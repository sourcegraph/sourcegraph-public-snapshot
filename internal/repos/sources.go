package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// A Sourcer converts the given ExternalService to a Source whose yielded Repos
// should be synced.
type Sourcer func(context.Context, *types.ExternalService) (Source, error)

// NewSourcer returns a Sourcer that converts the given ExternalService
// into a Source that uses the provided httpcli.Factory to create the
// http.Clients needed to contact the respective upstream code host APIs.
//
// The provided decorator functions will be applied to the Source.
func NewSourcer(logger log.Logger, db database.DB, cf *httpcli.Factory, decs ...func(Source) Source) Sourcer {
	return func(ctx context.Context, svc *types.ExternalService) (Source, error) {
		src, err := NewSource(ctx, logger.Scoped("source"), db, svc, cf)
		if err != nil {
			return nil, err
		}

		for _, dec := range decs {
			src = dec(src)
		}

		return src, nil
	}
}

// NewSource returns a repository yielding Source from the given ExternalService configuration.
func NewSource(ctx context.Context, logger log.Logger, db database.DB, svc *types.ExternalService, cf *httpcli.Factory) (Source, error) {
	switch strings.ToUpper(svc.Kind) {
	case extsvc.KindGitHub:
		return NewGitHubSource(ctx, logger.Scoped("GithubSource"), db, svc, cf)
	case extsvc.KindGitLab:
		return NewGitLabSource(ctx, logger.Scoped("GitLabSource"), svc, cf)
	case extsvc.KindAzureDevOps:
		return NewAzureDevOpsSource(ctx, logger.Scoped("AzureDevOpsSource"), svc, cf)
	case extsvc.KindGerrit:
		return NewGerritSource(ctx, svc, cf)
	case extsvc.KindBitbucketServer:
		return NewBitbucketServerSource(ctx, logger.Scoped("BitbucketServerSource"), svc, cf)
	case extsvc.KindBitbucketCloud:
		return NewBitbucketCloudSource(ctx, logger.Scoped("BitbucketCloudSource"), svc, cf)
	case extsvc.KindGitolite:
		return NewGitoliteSource(ctx, svc, cf)
	case extsvc.KindPhabricator:
		return NewPhabricatorSource(ctx, logger.Scoped("PhabricatorSource"), svc, cf)
	case extsvc.KindAWSCodeCommit:
		return NewAWSCodeCommitSource(ctx, svc, cf)
	case extsvc.KindPerforce:
		return NewPerforceSource(ctx, svc)
	case extsvc.KindGoPackages:
		return NewGoPackagesSource(ctx, svc, cf)
	case extsvc.KindJVMPackages:
		// JVM doesn't need a client factory because we use coursier.
		return NewJVMPackagesSource(ctx, svc)
	case extsvc.KindPagure:
		return NewPagureSource(ctx, svc, cf)
	case extsvc.KindNpmPackages:
		return NewNpmPackagesSource(ctx, svc, cf)
	case extsvc.KindPythonPackages:
		return NewPythonPackagesSource(ctx, svc, cf)
	case extsvc.KindRustPackages:
		return NewRustPackagesSource(ctx, svc, cf)
	case extsvc.KindRubyPackages:
		return NewRubyPackagesSource(ctx, svc, cf)
	case extsvc.KindOther:
		return NewOtherSource(ctx, svc, cf, logger.Scoped("OtherSource"))
	default:
		return nil, errors.Newf("cannot create source for kind %q", svc.Kind)
	}
}

// A Source yields repositories to be stored and analysed by Sourcegraph.
// Successive calls to its ListRepos method may yield different results.
type Source interface {
	// ListRepos sends all the repos a source yields over the passed in channel
	// as SourceResults
	ListRepos(context.Context, chan SourceResult)
	// CheckConnection returns an error if the Source service is not reachable
	// or available to serve requests. The error is descriptive and can be displayed
	// to the user.
	CheckConnection(context.Context) error
	// ExternalServices returns the ExternalServices for the Source.
	ExternalServices() types.ExternalServices
}

// RepoGetter captures the optional GetRepo method of a Source. It's used on
// sourcegraph.com to lazily sync individual repos and to lazily sync dependency
// repos on any customer instance.
type RepoGetter interface {
	GetRepo(context.Context, string) (*types.Repo, error)
}

type DependenciesServiceSource interface {
	Source
	SetDependenciesService(depsSvc *dependencies.Service)
}

// WithDependenciesService returns a decorator used in NewSourcer that calls SetDB on
// Sources that can be upgraded to it.
func WithDependenciesService(depsSvc *dependencies.Service) func(Source) Source {
	return func(src Source) Source {
		if s, ok := src.(DependenciesServiceSource); ok {
			s.SetDependenciesService(depsSvc)
			return s
		}
		return src
	}
}

// A UserSource is a source that can use a custom authenticator (such as one
// contained in a user credential) to interact with the code host, rather than
// global credentials.
type UserSource interface {
	// WithAuthenticator returns a copy of the original Source configured to use
	// the given authenticator, provided that authenticator type is supported by
	// the code host.
	WithAuthenticator(auth.Authenticator) (Source, error)
	// ValidateAuthenticator validates the currently set authenticator is usable.
	// Returns an error, when validating the Authenticator yielded an error.
	ValidateAuthenticator(ctx context.Context) error
}

type AffiliatedRepositorySource interface {
	AffiliatedRepositories(ctx context.Context) ([]types.CodeHostRepository, error)
}

// A VersionSource is a source that can query the version of the code host.
type VersionSource interface {
	Version(context.Context) (string, error)
}

// UnsupportedAuthenticatorError is returned by WithAuthenticator if the
// authenticator isn't supported on that code host.
type UnsupportedAuthenticatorError struct {
	have   string
	source string
}

func (e UnsupportedAuthenticatorError) Error() string {
	return fmt.Sprintf("authenticator type unsupported for %s sources: %s", e.source, e.have)
}

func newUnsupportedAuthenticatorError(source string, a auth.Authenticator) UnsupportedAuthenticatorError {
	return UnsupportedAuthenticatorError{
		have:   fmt.Sprintf("%T", a),
		source: source,
	}
}

// A SourceResult is sent by a Source over a channel for each repository it
// yields when listing repositories
type SourceResult struct {
	// Source points to the Source that produced this result
	Source Source
	// Repo is the repository that was listed by the Source
	Repo *types.Repo
	// Err is only set in case the Source ran into an error when listing repositories
	Err error
}

type SourceError struct {
	Err    error
	ExtSvc *types.ExternalService
}

func (s *SourceError) Error() string {
	var e errors.MultiError
	if errors.As(s.Err, &e) {
		// Create new Error with custom formatter. Do not mutate otherwise can
		// race with other callers of Error.
		return sourceErrorFormatFunc(e.Errors())
	}
	return s.Err.Error()
}

func (s *SourceError) Cause() error {
	return s.Err
}

func sourceErrorFormatFunc(es []error) string {
	if len(es) == 1 {
		return es[0].Error()
	}

	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(es), strings.Join(points, "\n\t"))
}

// ListAll calls ListRepos on the given Source and collects the SourceResults
// the Source sends over a channel into a slice of *types.Repo and a single error
func ListAll(ctx context.Context, src Source) ([]*types.Repo, error) {
	results := make(chan SourceResult)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	var (
		repos []*types.Repo
		errs  error
	)

	for res := range results {
		if res.Err != nil {
			for _, extSvc := range res.Source.ExternalServices() {
				errs = errors.Append(errs, &SourceError{Err: res.Err, ExtSvc: extSvc})
			}
			continue
		}
		repos = append(repos, res.Repo)
	}

	return repos, errs
}

// searchRepositories calls SearchRepositories on the given DiscoverableSource and collects the SourceResults
// the Source sends over a channel into a slice of *types.Repo and a single error
func searchRepositories(ctx context.Context, src DiscoverableSource, query string, first int, excludeRepos []string) ([]*types.Repo, error) {
	results := make(chan SourceResult)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		src.SearchRepositories(ctx, query, first, excludeRepos, results)
		close(results)
	}()

	var (
		repos []*types.Repo
		errs  error
	)

	for res := range results {
		if res.Err != nil {
			for _, extSvc := range res.Source.ExternalServices() {
				errs = errors.Append(errs, &SourceError{Err: res.Err, ExtSvc: extSvc})
			}
			continue
		}
		repos = append(repos, res.Repo)
	}

	return repos, errs
}
