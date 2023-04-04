package repos

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFakeSourcer returns a Sourcer which always returns the given error and source,
// ignoring the given external services.
func NewFakeSourcer(err error, src Source) Sourcer {
	return func(ctc context.Context, svc *types.ExternalService) (Source, error) {
		if err != nil {
			return nil, &SourceError{Err: err, ExtSvc: svc}
		}
		return src, nil
	}
}

// FakeSource is a fake implementation of Source to be used in tests.
type FakeSource struct {
	svc           *types.ExternalService
	repos         []*types.Repo
	err           error
	connectionErr error

	// ListRepos will send on this channel if it's not nil and wait on the channel
	// again before quitting. This can help with testing certain concurrent situation
	// in tests.
	lockChan chan struct{}
}

// NewFakeSource returns an instance of FakeSource with the given urn, error
// and repos.
func NewFakeSource(svc *types.ExternalService, err error, rs ...*types.Repo) *FakeSource {
	return &FakeSource{svc: svc, err: err, repos: rs}
}

// InitLockChan creates a non nil lock channel and returns it
func (s *FakeSource) InitLockChan() chan struct{} {
	s.lockChan = make(chan struct{})
	return s.lockChan
}

func (s *FakeSource) Unavailable() *FakeSource {
	s.connectionErr = errors.New("fake source unavailable")
	return s
}

func (s *FakeSource) CheckConnection(ctx context.Context) error {
	return s.connectionErr
}

// ListRepos returns the Repos that FakeSource was instantiated with
// as well as the error, if any.
func (s *FakeSource) ListRepos(ctx context.Context, results chan SourceResult) {
	if s.lockChan != nil {
		s.lockChan <- struct{}{}
		<-s.lockChan
	}

	if s.err != nil {
		results <- SourceResult{Source: s, Err: s.err}
		return
	}

	for _, r := range s.repos {
		results <- SourceResult{Source: s, Repo: r.With(typestest.Opt.RepoSources(s.svc.URN()))}
	}
}

func (s *FakeSource) GetRepo(ctx context.Context, name string) (*types.Repo, error) {
	for _, r := range s.repos {
		if strings.HasSuffix(string(r.Name), name) {
			return r, s.err
		}
	}

	if s.err == nil {
		return nil, errors.New("not found")
	}

	return nil, s.err
}

// ExternalServices returns a singleton slice containing the external service.
func (s *FakeSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *FakeDiscoverableSource) ListNamespaces(ctx context.Context, results chan SourceNamespaceResult) {
	if s.lockChan != nil {
		s.lockChan <- struct{}{}
		<-s.lockChan
	}

	if s.err != nil {
		results <- SourceNamespaceResult{Source: s, Err: s.err}
		return
	}

	for _, n := range s.namespaces {
		results <- SourceNamespaceResult{Source: s, Namespace: &types.ExternalServiceNamespace{ID: n.ID, Name: n.Name, ExternalID: n.ExternalID}}
	}
}

func (s *FakeDiscoverableSource) SearchRepositories(ctx context.Context, query string, first int, excludedRepos []string, results chan SourceResult) {
	if s.lockChan != nil {
		s.lockChan <- struct{}{}
		<-s.lockChan
	}

	if s.err != nil {
		results <- SourceResult{Source: s, Err: s.err}
		return
	}

	for _, r := range s.repos {
		results <- SourceResult{Source: s, Repo: r}
	}
}

// FakeDiscoverableSource is a fake implementation of DiscoverableSource to be used in tests.
type FakeDiscoverableSource struct {
	*FakeSource
	namespaces []*types.ExternalServiceNamespace
}

// NewFakeDiscoverableSource returns an instance of FakeDiscoverableSource with the given namespaces.
func NewFakeDiscoverableSource(fs *FakeSource, connectionErr bool, ns ...*types.ExternalServiceNamespace) *FakeDiscoverableSource {
	if connectionErr {
		fs.Unavailable()
	}
	return &FakeDiscoverableSource{FakeSource: fs, namespaces: ns}
}
