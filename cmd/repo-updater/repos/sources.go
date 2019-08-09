package repos

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
)

// A Sourcer converts the given ExternalServices to Sources
// whose yielded Repos should be synced.
type Sourcer func(...*ExternalService) (Sources, error)

// NewSourcer returns a Sourcer that converts the given ExternalServices
// into Sources that use the provided httpcli.Factory to create the
// http.Clients needed to contact the respective upstream code host APIs.
//
// Deleted external services are ignored.
//
// The provided decorator functions will be applied to each Source.
func NewSourcer(cf *httpcli.Factory, decs ...func(Source) Source) Sourcer {
	return func(svcs ...*ExternalService) (Sources, error) {
		srcs := make([]Source, 0, len(svcs))
		errs := new(multierror.Error)

		for _, svc := range svcs {
			if svc.IsDeleted() {
				continue
			}

			src, err := NewSource(svc, cf)
			if err != nil {
				errs = multierror.Append(errs, err)
				continue
			}

			for _, dec := range decs {
				src = dec(src)
			}

			srcs = append(srcs, src)
		}

		return srcs, errs.ErrorOrNil()
	}
}

// NewSource returns a repository yielding Source from the given ExternalService configuration.
func NewSource(svc *ExternalService, cf *httpcli.Factory) (Source, error) {
	switch strings.ToLower(svc.Kind) {
	case "github":
		return NewGithubSource(svc, cf)
	case "gitlab":
		return NewGitLabSource(svc, cf)
	case "bitbucketserver":
		return NewBitbucketServerSource(svc, cf)
	case "bitbucketcloud":
		return NewBitbucketCloudSource(svc, cf)
	case "gitolite":
		return NewGitoliteSource(svc, cf)
	case "phabricator":
		return NewPhabricatorSource(svc, cf)
	case "awscodecommit":
		return NewAWSCodeCommitSource(svc, cf)
	case "other":
		return NewOtherSource(svc)
	default:
		panic(fmt.Sprintf("source not implemented for external service kind %q", svc.Kind))
	}
}

// sourceTimeout is the default timeout to use on Source.ListRepos
const sourceTimeout = 30 * time.Minute

// A Source yields repositories to be stored and analysed by Sourcegraph.
// Successive calls to its ListRepos method may yield different results.
type Source interface {
	// ListRepos returns all the repos a source yields.
	ListRepos(context.Context) ([]*Repo, error)
	// ExternalServices returns the ExternalServices for the Source.
	ExternalServices() ExternalServices
}

// Sources is a list of Sources that implements the Source interface.
type Sources []Source

// ListRepos lists all the repos of all the sources and returns the
// aggregate result.
func (srcs Sources) ListRepos(ctx context.Context) ([]*Repo, error) {
	if len(srcs) == 0 {
		return nil, nil
	}

	type result struct {
		src   Source
		repos []*Repo
		err   error
	}

	// Group sources by external service kind so that we execute requests
	// serially to each code host. This is to comply with abuse rate limits of GitHub,
	// but we do it for any source to be conservative.
	// See https://developer.github.com/v3/guides/best-practices-for-integrators/#dealing-with-abuse-rate-limits)

	var wg sync.WaitGroup
	ch := make(chan result)
	for _, sources := range group(srcs) {
		wg.Add(1)
		go func(sources Sources) {
			defer wg.Done()
			for _, src := range sources {
				if repos, err := src.ListRepos(ctx); err != nil {
					ch <- result{src: src, err: err}
				} else {
					ch <- result{src: src, repos: repos}
				}
			}
		}(sources)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var repos []*Repo
	errs := new(multierror.Error)

	for r := range ch {
		if r.err != nil {
			errs = multierror.Append(errs, r.err)
		} else {
			repos = append(repos, r.repos...)
		}
	}

	return repos, errs.ErrorOrNil()
}

// ExternalServices returns the ExternalServices from the given Sources.
func (srcs Sources) ExternalServices() ExternalServices {
	es := make(ExternalServices, 0, len(srcs))
	for _, src := range srcs {
		es = append(es, src.ExternalServices()...)
	}
	return es
}

type multiSource interface {
	Sources() []Source
}

// Sources returns the underlying Sources.
func (srcs Sources) Sources() []Source { return srcs }

func group(srcs []Source) map[string]Sources {
	groups := make(map[string]Sources)

	for _, src := range srcs {
		if ms, ok := src.(multiSource); ok {
			for kind, ss := range group(ms.Sources()) {
				groups[kind] = append(groups[kind], ss...)
			}
		} else if es := src.ExternalServices(); len(es) > 1 {
			err := errors.Errorf("Source %#v has many external services and isn't a multiSource", src)
			panic(err)
		} else {
			kind := es[0].Kind
			groups[kind] = append(groups[kind], src)
		}
	}

	return groups
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_501(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
