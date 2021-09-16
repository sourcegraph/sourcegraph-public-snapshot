package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// A Sourcer converts the given ExternalService to a Source whose yielded Repos
// should be synced.
type Sourcer func(*types.ExternalService) (Source, error)

// NewSourcer returns a Sourcer that converts the given ExternalService
// into a Source that uses the provided httpcli.Factory to create the
// http.Clients needed to contact the respective upstream code host APIs.
//
// The provided decorator functions will be applied to the Source.
func NewSourcer(cf *httpcli.Factory, decs ...func(Source) Source) Sourcer {
	return func(svc *types.ExternalService) (Source, error) {
		src, err := NewSource(svc, cf)
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
func NewSource(svc *types.ExternalService, cf *httpcli.Factory) (Source, error) {
	switch strings.ToUpper(svc.Kind) {
	case extsvc.KindGitHub:
		return NewGithubSource(svc, cf)
	case extsvc.KindGitLab:
		return NewGitLabSource(svc, cf)
	case extsvc.KindBitbucketServer:
		return NewBitbucketServerSource(svc, cf)
	case extsvc.KindBitbucketCloud:
		return NewBitbucketCloudSource(svc, cf)
	case extsvc.KindGitolite:
		return NewGitoliteSource(svc, cf)
	case extsvc.KindPhabricator:
		return NewPhabricatorSource(svc, cf)
	case extsvc.KindAWSCodeCommit:
		return NewAWSCodeCommitSource(svc, cf)
	case extsvc.KindPerforce:
		return NewPerforceSource(svc)
	case extsvc.KindJVMPackages:
		return NewJVMPackagesSource(svc)
	case extsvc.KindOther:
		return NewOtherSource(svc, cf)
	default:
		return nil, fmt.Errorf("cannot create source for kind %q", svc.Kind)
	}
}

// A Source yields repositories to be stored and analysed by Sourcegraph.
// Successive calls to its ListRepos method may yield different results.
type Source interface {
	// ListRepos sends all the repos a source yields over the passed in channel
	// as SourceResults
	ListRepos(context.Context, chan SourceResult)
	// ExternalServices returns the ExternalServices for the Source.
	ExternalServices() types.ExternalServices
}

// RepoGetter captures the optional GetRepo method of a Source. It's used only
// on sourcegraph.com to lazily sync individual repos.
type RepoGetter interface {
	GetRepo(context.Context, string) (*types.Repo, error)
}

type DBSource interface {
	Source
	SetDB(dbutil.DB)
}

// WithDB returns a decorator used in NewSourcer that calls SetDB on Sources that
// can be upgraded to it.
func WithDB(db dbutil.DB) func(Source) Source {
	return func(src Source) Source {
		if s, ok := src.(DBSource); ok {
			s.SetDB(db)
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
	var e *multierror.Error
	if errors.As(s.Err, &e) {
		// Create new Error with custom formatter. Do not mutate otherwise can
		// race with other callers of Error.
		return (&multierror.Error{
			Errors:      e.Errors,
			ErrorFormat: sourceErrorFormatFunc,
		}).Error()
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

// listAll calls ListRepos on the given Source and collects the SourceResults
// the Source sends over a channel into a slice of *types.Repo and a single error
func listAll(ctx context.Context, src Source) ([]*types.Repo, error) {
	results := make(chan SourceResult)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	var (
		repos []*types.Repo
		errs  *multierror.Error
	)

	for res := range results {
		if res.Err != nil {
			for _, extSvc := range res.Source.ExternalServices() {
				errs = multierror.Append(errs, &SourceError{Err: res.Err, ExtSvc: extSvc})
			}
			continue
		}
		repos = append(repos, res.Repo)
	}

	return repos, errs.ErrorOrNil()
}
