package graphqlbackend

import (
	"context"
	"fmt"
	"regexp/syntax" //nolint:depguard // using the grafana fork of regexp clashes with zoekt, which uses the std regexp/syntax.
	"sync"
	"time"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"
	"github.com/sourcegraph/zoekt/stream"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchzoekt "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

func (r *RepositoryResolver) TextSearchIndex() *repositoryTextSearchIndexResolver {
	return &repositoryTextSearchIndexResolver{
		repo:   r,
		client: search.Indexed(),
	}
}

type repositoryTextSearchIndexResolver struct {
	repo   *RepositoryResolver
	client zoekt.Streamer

	once  sync.Once
	entry *zoekt.RepoListEntry
	err   error
}

func (r *repositoryTextSearchIndexResolver) resolve(ctx context.Context) (*zoekt.RepoListEntry, error) {
	r.once.Do(func() {
		q := zoektquery.NewSingleBranchesRepos("HEAD", uint32(r.repo.IDInt32()))
		repoList, err := r.client.List(ctx, q, nil)
		if err != nil {
			r.err = err
			return
		}
		// During rebalancing we have a repo on more than one shard. Pick the
		// newest one since that will be the winner.
		var latest time.Time
		for _, entry := range repoList.Repos {
			if t := entry.IndexMetadata.IndexTime; t.After(latest) {
				r.entry = entry
				latest = t
			}
		}
	})
	return r.entry, r.err
}

func (r *repositoryTextSearchIndexResolver) Repository() *RepositoryResolver { return r.repo }

func (r *repositoryTextSearchIndexResolver) Status(ctx context.Context) (*repositoryTextSearchIndexStatus, error) {
	entry, err := r.resolve(ctx)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return &repositoryTextSearchIndexStatus{entry: *entry}, nil
}

func (r *repositoryTextSearchIndexResolver) Host(ctx context.Context) (*repositoryIndexserverHostResolver, error) {
	// We don't want to let the user wait for too long. If the socket
	// connection is working, 500ms should be generous.
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	host, err := searchzoekt.GetIndexserverHost(ctx, r.repo.RepoName())
	if err != nil {
		return nil, nil
	}
	return &repositoryIndexserverHostResolver{
		host,
	}, nil
}

type repositoryIndexserverHostResolver struct {
	host searchzoekt.Host
}

func (r *repositoryIndexserverHostResolver) Name(ctx context.Context) string {
	return r.host.Name
}

type repositoryTextSearchIndexStatus struct {
	entry zoekt.RepoListEntry
}

func (r *repositoryTextSearchIndexStatus) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.entry.IndexMetadata.IndexTime}
}

func (r *repositoryTextSearchIndexStatus) ContentByteSize() BigInt {
	return BigInt(r.entry.Stats.ContentBytes)
}

func (r *repositoryTextSearchIndexStatus) ContentFilesCount() int32 {
	return int32(r.entry.Stats.Documents)
}

func (r *repositoryTextSearchIndexStatus) IndexByteSize() int32 {
	return int32(r.entry.Stats.IndexBytes)
}

func (r *repositoryTextSearchIndexStatus) IndexShardsCount() int32 {
	return int32(r.entry.Stats.Shards)
}

func (r *repositoryTextSearchIndexStatus) NewLinesCount() int32 {
	return int32(r.entry.Stats.NewLinesCount)
}

func (r *repositoryTextSearchIndexStatus) DefaultBranchNewLinesCount() int32 {
	return int32(r.entry.Stats.DefaultBranchNewLinesCount)
}

func (r *repositoryTextSearchIndexStatus) OtherBranchesNewLinesCount() int32 {
	return int32(r.entry.Stats.OtherBranchesNewLinesCount)
}

func (r *repositoryTextSearchIndexResolver) Refs(ctx context.Context) ([]*repositoryTextSearchIndexedRef, error) {
	// We assume that the default branch for enabled repositories is always configured to be indexed.
	//
	// TODO(sqs): support configuring which branches should be indexed (add'l branches, not default branch, etc.).
	repoResolver := r.repo
	defaultBranchRef, err := repoResolver.DefaultBranch(ctx)
	if err != nil {
		return nil, err
	}
	if defaultBranchRef == nil {
		return []*repositoryTextSearchIndexedRef{}, nil
	}
	refNames := []string{defaultBranchRef.name}

	refs := make([]*repositoryTextSearchIndexedRef, len(refNames))
	for i, refName := range refNames {
		refs[i] = &repositoryTextSearchIndexedRef{ref: &GitRefResolver{name: refName, repo: repoResolver}}
	}
	refByName := func(name string) *repositoryTextSearchIndexedRef {
		possibleRefNames := []string{"refs/heads/" + name, "refs/tags/" + name}
		for _, ref := range possibleRefNames {
			if _, err := repoResolver.gitserverClient.ResolveRevision(ctx, repoResolver.RepoName(), ref); err == nil {
				name = ref
				break
			}
		}
		for _, ref := range refs {
			if ref.ref.name == name {
				return ref
			}
		}

		// If Zoekt reports it has another indexed branch, include that.
		newRef := &repositoryTextSearchIndexedRef{ref: &GitRefResolver{name: name, repo: repoResolver}}
		refs = append(refs, newRef)
		return newRef
	}

	entry, err := r.resolve(ctx)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		for _, branch := range entry.Repository.Branches {
			name := branch.Name
			if branch.Name == "HEAD" {
				name = defaultBranchRef.name
			}
			ref := refByName(name)
			ref.indexedCommit = GitObjectID(branch.Version)
			ref.skippedIndexed = &skippedIndexedResolver{
				repo:   r.repo,
				branch: branch.Name,
				client: r.client,
			}
		}
	}
	return refs, nil
}

type repositoryTextSearchIndexedRef struct {
	ref            *GitRefResolver
	indexedCommit  GitObjectID
	skippedIndexed *skippedIndexedResolver
}

func (r *repositoryTextSearchIndexedRef) Ref() *GitRefResolver { return r.ref }
func (r *repositoryTextSearchIndexedRef) Indexed() bool        { return r.indexedCommit != "" }

func (r *repositoryTextSearchIndexedRef) Current(ctx context.Context) (bool, error) {
	if r.indexedCommit == "" {
		return false, nil
	}

	commit, err := r.ref.Target().Commit(ctx)
	if err != nil {
		return false, err
	}
	return commit.oid == r.indexedCommit, nil
}

func (r *repositoryTextSearchIndexedRef) IndexedCommit() *gitObject {
	if r.indexedCommit == "" {
		return nil
	}
	return &gitObject{repo: r.ref.repo, oid: r.indexedCommit, typ: GitObjectTypeCommit}
}

func (r *repositoryTextSearchIndexedRef) SkippedIndexed() *skippedIndexedResolver {
	return r.skippedIndexed
}

type skippedIndexedResolver struct {
	repo   *RepositoryResolver
	branch string

	client zoekt.Streamer
}

func (r *skippedIndexedResolver) Count(ctx context.Context) (BigInt, error) {
	// During indexing, Zoekt may decide to skip a document for various reasons. If
	// a document is skipped, Zoekt replaces the content of the skipped document
	// with "NOT-INDEXED: <reason>"
	expr, err := syntax.Parse("^NOT-INDEXED: ", syntax.Perl)
	if err != nil {
		return 0, err
	}

	q := &zoektquery.And{Children: []zoektquery.Q{
		&zoektquery.Regexp{Regexp: expr, Content: true, CaseSensitive: true},
		zoektquery.NewSingleBranchesRepos(r.branch, uint32(r.repo.IDInt32())),
	}}

	var stats zoekt.Stats
	if err := r.client.StreamSearch(
		ctx,
		q,
		&zoekt.SearchOptions{},
		stream.SenderFunc(func(sr *zoekt.SearchResult) {
			stats.Add(sr.Stats)
		}),
	); err != nil {
		return 0, err
	}

	return BigInt(stats.FileCount), nil
}

func (r *skippedIndexedResolver) Query() string {
	// Adding select:file renders the results as path match instead of content
	// match. This is important because the indexed content (NOT-INDEXED: <reason>)
	// is different from the on-disk content served by gitserver which leads to
	// broken highlighting and problems with rendering content of binary files.
	return fmt.Sprintf("r:^%s$@%s type:file select:file index:only patternType:regexp ^NOT-INDEXED:", regexp.QuoteMeta(r.repo.Name()), r.branch)
}
