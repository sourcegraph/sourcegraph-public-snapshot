package resolvers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	pathpkg "path"
	"sort"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (r *componentResolver) Branches(ctx context.Context, args *gql.GitRefConnectionArgs) (gql.GitRefConnectionResolver, error) {
	slocs, err := r.sourceLocationSetResolver(ctx)
	if err != nil {
		return nil, err
	}
	return slocs.Branches(ctx, args)
}

func (r *rootResolver) GitTreeEntryBranches(ctx context.Context, treeEntry *gql.GitTreeEntryResolver, args *gql.GitRefConnectionArgs) (gql.GitRefConnectionResolver, error) {
	return sourceLocationSetResolverFromTreeEntry(treeEntry, r.db).Branches(ctx, args)
}

func (r *sourceLocationSetResolver) Branches(ctx context.Context, args *gql.GitRefConnectionArgs) (gql.GitRefConnectionResolver, error) {
	v := gql.GitRefTypeBranch
	args.Type = &v

	// TODO(sqs): consolidate impl with that in package graphqlbackend (repository_git_refs.go)?
	// That impl is actually different and potentially faster: it first lists all branch names, then
	// looks up per-branch to see if it matches. We'd need to add a path filter to that GQL field's
	// args and implement that to make it usable here, but that seems doable. (Then we would NOT use
	// the diff search path as we do here currently.)

	// TODO(sqs): This will miss branches whose commit is very old. Instead, the impl should be
	// changed to iterate through all branches and find commits that affect the component's files.

	var allRefs []*gql.GitRefResolver
	for sloc, paths := range groupSourceLocationsByRepo(r.slocs) {
		matches, err := getBranchesForRepo(ctx, sloc.repoName, joinPathPrefixRegexps(paths))
		if err != nil {
			return nil, err
		}

		ignoreRef := func(refName string) bool {
			return strings.HasPrefix(refName, "refs/heads/main-dry-run/") || strings.HasPrefix(refName, "refs/heads/docker-images/") || strings.HasPrefix(refName, "refs/heads/docker-images-patch/") || strings.HasPrefix(refName, "refs/heads/docker-images-patch-notest/")
		}

		type refData struct {
			*gql.GitRefResolver
			index int // support sorting by the original order returned from `git log`
		}
		refsByName := map[string]refData{}
		for i, m := range matches {
			for _, sourceRef := range m.SourceRefs {
				if ignoreRef(sourceRef) {
					continue
				}
				if _, ok := refsByName[sourceRef]; !ok {
					refsByName[sourceRef] = refData{
						GitRefResolver: gql.NewGitRefResolver(sloc.repo, sourceRef, gql.GitObjectID(m.Oid)),
						index:          i,
					}
				}
			}
		}

		refs := make([]*gql.GitRefResolver, 0, len(refsByName))
		for _, refResolver := range refsByName {
			refs = append(refs, refResolver.GitRefResolver)
		}
		sort.Slice(refs, func(i, j int) bool {
			return refsByName[refs[i].Name()].index < refsByName[refs[j].Name()].index
		})
		allRefs = append(allRefs, refs...)
	}

	return gql.NewGitRefConnectionResolver(args.First, allRefs), nil
}

func getBranchesForRepo(ctx context.Context, repo api.RepoName, pathRegexp string) ([]protocol.CommitMatch, error) {
	type cacheEntry struct {
		Data []protocol.CommitMatch
	}
	cachePath := func(repoName api.RepoName, pathRegexp string) string {
		dir := "/tmp/sqs-wip-cache/getBranchesForRepo/" + pathpkg.Base(string(repoName))
		_ = os.MkdirAll(dir, 0700)

		b, err := json.Marshal([]interface{}{repoName, pathRegexp})
		if err != nil {
			panic(err)
		}
		h := sha256.Sum256(b)
		name := hex.EncodeToString(h[:])
		return pathpkg.Join(dir, name)
	}
	get := func(path string) (cacheEntry, bool) {
		b, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return cacheEntry{}, false
		}
		if err != nil {
			panic(err)
		}
		var v cacheEntry
		if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&v); err != nil {
			panic(err)
		}
		return v, true
	}
	set := func(path string, data cacheEntry) {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(data); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(path, buf.Bytes(), 0600); err != nil {
			panic(err)
		}
	}

	getBranchesForRepoUncached := func(ctx context.Context, repo api.RepoName, pathRegexp string) ([]protocol.CommitMatch, error) {
		req := &protocol.SearchRequest{
			Repo: repo,
			Revisions: []protocol.RevisionSpecifier{
				{RevSpec: "^HEAD"},
				{RefGlob: "refs/heads/"},
			},
			Limit: 10, // TODO(sqs)
			Query: &protocol.DiffModifiesFile{Expr: pathRegexp},
		}

		var allMatches []protocol.CommitMatch
		onMatches := func(matches []protocol.CommitMatch) {
			allMatches = append(allMatches, matches...)
		}
		if _, err := gitserver.DefaultClient.Search(ctx, req, onMatches); err != nil {
			return nil, err
		}
		return allMatches, nil
	}
	v, ok := get(cachePath(repo, pathRegexp))
	if ok {
		return v.Data, nil
	}

	allMatches, err := getBranchesForRepoUncached(ctx, repo, pathRegexp)
	if err == nil {
		set(cachePath(repo, pathRegexp), cacheEntry{Data: allMatches})
	}
	return allMatches, err
}
