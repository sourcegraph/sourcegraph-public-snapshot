package resolvers

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"pgregory.net/rapid"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/pbt"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func malformedAfterCursorGenerator() *rapid.Generator[*string] {
	return rapid.Custom(func(t *rapid.T) *string {
		val := codenav.UsagesCursor{}
		val.CursorType = "nonsense"
		bytes, err := json.Marshal(val)
		require.NoError(t, err)
		if rapid.Bool().Draw(t, "") {
			return pointers.Ptr(string(bytes))
		}
		if rapid.Bool().Draw(t, "") {
			// Wrong encoding
			return pointers.Ptr(base64.URLEncoding.EncodeToString(bytes))
		}
		return pointers.Ptr(base64.StdEncoding.EncodeToString(bytes))
	})
}

func wellFormedAfterCursorGenerator() *rapid.Generator[*string] {
	cursorTypeGen := rapid.OneOf(
		rapid.Just(codenav.CursorTypeDefinitions),
		rapid.Just(codenav.CursorTypeReferences),
		rapid.Just(codenav.CursorTypeImplementations),
		rapid.Just(codenav.CursorTypePrototypes),
	)
	return rapid.Custom(func(t *rapid.T) *string {
		val := codenav.UsagesCursor{}
		val.CursorType = cursorTypeGen.Draw(t, "cursortype")
		bytes, err := json.Marshal(val)
		require.NoError(t, err)
		s := base64.StdEncoding.EncodeToString(bytes)
		return &s
	})
}

func afterCursorGenerator(rng *rand.Rand) *rapid.Generator[*string] {
	return pbt.WithProbabilities[*string](rng, []pbt.GeneratorChoice[*string]{
		{
			Chance: 0.5,
			Value:  rapid.Just[*string](nil),
		},
		{
			Chance: 0.45,
			Value:  wellFormedAfterCursorGenerator(),
		},
		{
			Chance: 0.05,
			Value:  malformedAfterCursorGenerator(),
		},
	})
}

func usagesFilterGenerator() *rapid.Generator[*UsagesFilter] {
	repoGen := rapid.Ptr(rapid.Make[RepositoryFilter](), true)
	return rapid.OneOf(
		rapid.Just[*UsagesFilter](nil),
		rapid.Custom(func(t *rapid.T) *UsagesFilter {
			notFilter := rapid.Deferred(usagesFilterGenerator).Draw(t, "")
			return &UsagesFilter{notFilter, repoGen.Draw(t, "repo")}
		}))
}

func usagesForSymbolArgsGenerator(rng *rand.Rand) *rapid.Generator[*UsagesForSymbolArgs] {
	revisionGen := pbt.WithProbabilities[string](rng, []pbt.GeneratorChoice[string]{
		{
			Chance: 0.5,
			Value:  rapid.Map(pbt.CommitID(), func(id api.CommitID) string { return string(id) }),
		},
		{
			Chance: 0.20,
			Value:  rapid.Just("HEAD"),
		},
		{
			Chance: 0.05,
			Value:  rapid.Just("mybranch"),
		},
		{
			Chance: 0.25,
			Value:  rapid.Just(""),
		},
	})
	symbolGen := rapid.Make[*SymbolComparator]()
	rangeGen := rapid.Make[RangeInput]()
	cursorGen := afterCursorGenerator(rng)
	filterGen := usagesFilterGenerator()
	firstGen := rapid.Ptr(rapid.Int32Range(-1, 250), true)
	return rapid.Custom(func(t *rapid.T) *UsagesForSymbolArgs {
		val := UsagesForSymbolArgs{
			symbolGen.Draw(t, "symbolComparator"),
			rangeGen.Draw(t, "range"),
			filterGen.Draw(t, "usagesFilter"),
			firstGen.Draw(t, "first"),
			cursorGen.Draw(t, "after"),
		}
		if val.Range.Revision != nil {
			val.Range.Revision = pointers.Ptr(revisionGen.Draw(t, "revision"))
		}
		val.After = cursorGen.Draw(t, "after")
		return &val
	})
}

type fakeRepoStore struct {
	repos  map[api.RepoName]*types.Repo
	nextID int
}

func (s *fakeRepoStore) GetByName(name api.RepoName) (*types.Repo, error) {
	if s.repos != nil {
		if val, ok := s.repos[name]; ok {
			return val, nil
		}
	}
	return nil, &database.RepoNotFoundErr{}
}

func (s *fakeRepoStore) Insert(name api.RepoName) *types.Repo {
	if s.repos == nil {
		s.repos = make(map[api.RepoName]*types.Repo)
	}
	repo := types.Repo{
		ID:   api.RepoID(s.nextID),
		Name: name,
	}
	s.nextID += 1
	s.repos[name] = &repo
	return &repo
}

type fakeGitserverClient struct {
	commits map[api.RepoName]map[string]api.CommitID
}

func (c *fakeGitserverClient) ResolveRevision(repo api.RepoName, rev string) (api.CommitID, error) {
	if c.commits != nil {
		if m, ok := c.commits[repo]; ok {
			if commitID, ok := m[rev]; ok {
				return commitID, nil
			}
			return "", &gitdomain.RevisionNotFoundError{}
		}
	}
	return "", &gitdomain.RepoNotExistError{}
}

func (c *fakeGitserverClient) Insert(repo api.RepoName, rev string) api.CommitID {
	if c.commits == nil {
		c.commits = make(map[api.RepoName]map[string]api.CommitID)
	}
	if _, ok := c.commits[repo]; !ok {
		c.commits[repo] = make(map[string]api.CommitID)
	}
	if commitID, err := api.NewCommitID(rev); err == nil {
		c.commits[repo][rev] = commitID
		return commitID
	}
	bytes := sha1.Sum([]byte(string(repo) + rev))
	commitID := api.CommitID(bytes[:])
	c.commits[repo][rev] = commitID
	return commitID
}

func TestResolve(t *testing.T) {
	mockRepoStore := dbmocks.NewStrictMockRepoStore()
	mockGitserverClient := gitserver.NewStrictMockClient()

	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64().Draw(t, "seed")
		rng := rand.New(rand.NewSource(seed))
		p90Bool := pbt.Bool(rng, 0.9)

		args := usagesForSymbolArgsGenerator(rng).Draw(t, "args")
		// NOTE: For some reason, moving the p90Bool.Draw operations inside the hook
		// functions triggers a bug inside rapid with a buffer overrun or running
		// out of entropy bits.
		repoStoreImpl := fakeRepoStore{}
		gitserverClientImpl := fakeGitserverClient{}

		mockRepoStore.GetByNameFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
			if repo, err := repoStoreImpl.GetByName(name); err == nil {
				return repo, nil
			}
			if p90Bool.Draw(t, "repo") {
				return repoStoreImpl.Insert(name), nil
			} else {
				return nil, &database.RepoNotFoundErr{}
			}
		})

		mockGitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, revision string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
			if commitID, err := gitserverClientImpl.ResolveRevision(repo, revision); err == nil {
				return commitID, nil
			}
			if p90Bool.Draw(t, "commit") {
				return gitserverClientImpl.Insert(repo, revision), nil
			} else {
				return "", &gitdomain.RevisionNotFoundError{}
			}
		})

		require.NotPanics(t, func() {
			_, _ = args.Resolve(context.Background(), mockRepoStore, mockGitserverClient, 100)
		})
	})
}
