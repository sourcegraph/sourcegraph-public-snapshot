package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/xlang"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

func searchReferencesInRepos(ctx context.Context, args *search.Args) (res []*searchResultResolver, common *searchResultsCommon, err error) {
	refValues, negatedRefValues := args.Query.StringValues(query.FieldRef)
	if len(negatedRefValues) != 0 {
		return nil, nil, errors.New("not supported: negated references queries (-ref:)")
	}
	if len(refValues) != 1 {
		return nil, nil, errors.New("search query must have at most 1 ref: value")
	}
	if len(args.Query.Values(query.FieldDefault)) > 0 {
		return nil, nil, errors.New("not yet supported: combining references search query (ref:) and text search patterns")
	}

	var symbol lspext.SymbolDescriptor // the symbol that the ref: field refers to
	if err := json.Unmarshal([]byte(refValues[0]), &symbol); err != nil {
		return nil, nil, errors.Wrap(err, "parsing ref: value")
	}
	var hints map[string]interface{} // hints for speeding up workspace/xreferences
	if hintValues, _ := args.Query.StringValues(query.FieldHints); len(hintValues) > 0 {
		if err := json.Unmarshal([]byte(hintValues[0]), &hints); err != nil {
			return nil, nil, errors.Wrap(err, "parsing hints: value")
		}
	}
	var language string
	if langValues, _ := args.Query.StringValues(query.FieldLang); len(langValues) == 0 {
		return nil, nil, errors.New("references search query must have a lang: value (such as lang:go)")
	} else if len(langValues) >= 2 {
		return nil, nil, errors.New("not supported: multiple lang: values in references search")
	} else {
		language = langValues[0]
	}

	tr, ctx := trace.New(ctx, "searchReferencesInRepos", fmt.Sprintf("language: %s, symbol: %+v, numRepoRevs: %d", language, symbol, len(args.Repos)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// Speed up references search by consulting the global_dep index to see who else
	// references this symbol (avoiding the need to search packages in repositories that
	// don't reference it). Only do this if no repository filters are specified in the query.
	if len(args.Query.Values(query.FieldRepo)) == 0 && len(args.Query.Values(query.FieldRepoGroup)) == 0 {
		pkgDescriptor, ok := xlang.SymbolPackageDescriptor(symbol, language)
		if ok {
			// NOTE: This clobbers the package's version, which may not be desirable in
			// the future when we want to offer precisely versioned searches (e.g., find
			// me all references to function foo in package bar at version 1.2.3).
			pkgDescriptor, ok = xlang.PackageIdentifier(pkgDescriptor, language)
		}
		if ok {
			dependents, err := db.GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
				Language: language,
				DepData:  pkgDescriptor,
			})
			if err != nil {
				return nil, nil, err
			}
			repoIsDependent := func(repo api.RepoID) bool {
				for _, d := range dependents {
					if d.RepoID == repo {
						return true
					}
				}
				return false
			}

			// Only search in the intersection of args.repos and dependents.RepoID. (Also
			// include repos that aren't yet indexed, since for those their absence from
			// dependents.RepoID doesn't imply they aren't a dependent.)
			tmp := *args
			args = &tmp
			keepRepos := args.Repos[:0]
			for _, repo := range args.Repos {
				if repo.Repo.IndexedRevision == nil || repoIsDependent(repo.Repo.ID) {
					keepRepos = append(keepRepos, repo)
				}
			}
			args.Repos = keepRepos
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg            sync.WaitGroup
		mu            sync.Mutex
		unflattened   [][]*fileMatchResolver
		flattenedSize int
	)

	// addMatches assumes the caller holds mu.
	addMatches := func(matches []*fileMatchResolver) {
		if len(matches) > 0 {
			common.resultCount += int32(len(matches))
			sort.Slice(matches, func(i, j int) bool {
				a, b := matches[i].uri, matches[j].uri
				return a > b
			})
			unflattened = append(unflattened, matches)
			flattenedSize += len(matches)

			// Stop searching once we have found enough matches. This does
			// lead to potentially unstable result ordering, but is worth
			// it for the performance benefit.
			if flattenedSize > int(args.Pattern.FileMatchLimit) {
				tr.LazyPrintf("cancel due to result size: %d > %d", flattenedSize, args.Pattern.FileMatchLimit)
				common.limitHit = true
				cancel()
			}
		}
	}

	common = &searchResultsCommon{}
	for _, repoRev := range args.Repos {
		if len(repoRev.Revs) == 0 {
			return nil, common, nil // no revs to search
		}
		if len(repoRev.Revs) >= 2 {
			return nil, common, errMultipleRevsNotSupported
		}

		wg.Add(1)
		go func(repoRev search.RepositoryRevisions) {
			defer wg.Done()
			rev := repoRev.RevSpecs()[0] // TODO(sqs): search multiple revs
			matches, repoLimitHit, searchErr := searchReferencesInRepo(ctx, repoRev.Repo, repoRev.GitserverRepo, rev, language, symbol, hints, args.Pattern)
			if searchErr != nil {
				tr.LogFields(otlog.String("repo", string(repoRev.Repo.URI)), otlog.String("searchErr", searchErr.Error()), otlog.Bool("timeout", errcode.IsTimeout(searchErr)), otlog.Bool("temporary", errcode.IsTemporary(searchErr)))
			}
			mu.Lock()
			defer mu.Unlock()
			if ctx.Err() == nil {
				common.searched = append(common.searched, repoRev.Repo)
			}
			// non-diff search reports timeout through searchErr, so pass false for timedOut
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, false, searchErr); fatalErr != nil {
				if ctx.Err() != nil {
					// Our request has been canceled, we can just ignore
					// searchReferencesInRepo for this repo. We only check this condition
					// here since handleRepoSearchResult handles deadlines
					// exceeded differently to canceled.
					return
				}
				err = errors.Wrapf(searchErr, "failed to search %s", repoRev.String())
				tr.LazyPrintf("cancel due to error: %v", err)
				cancel()
			}
			addMatches(matches)
		}(*repoRev)
	}

	wg.Wait()
	if err != nil {
		return nil, common, err
	}

	flattened := flattenFileMatches(unflattened, int(args.Pattern.FileMatchLimit))
	return fileMatchesToSearchResults(flattened), common, nil
}

var mockSearchReferencesInRepo func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev, language string, symbol lspext.SymbolDescriptor, hints map[string]interface{}, query *search.PatternInfo) (matches []*fileMatchResolver, limitHit bool, err error)

func searchReferencesInRepo(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev, language string, symbol lspext.SymbolDescriptor, hints map[string]interface{}, query *search.PatternInfo) (matches []*fileMatchResolver, limitHit bool, err error) {
	if mockSearchReferencesInRepo != nil {
		return mockSearchReferencesInRepo(ctx, repo, gitserverRepo, rev, language, symbol, hints, query)
	}

	commit, err := git.ResolveRevision(ctx, gitserverRepo, nil, rev, nil)
	if err != nil {
		return nil, false, err
	}

	// We expect references search to be slow in many cases.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	refs, err := backend.LangServer.WorkspaceXReferences(ctx, repo, commit, language, lspext.WorkspaceReferencesParams{
		Query: symbol,
		Hints: hints,
	})
	if err != nil {
		return nil, false, err
	}

	// Group refs by file.
	refsByFile := map[lsp.DocumentURI][]lsp.Location{}
	var files []lsp.DocumentURI
	for _, ref := range refs {
		if _, seen := refsByFile[ref.Reference.URI]; !seen {
			files = append(files, ref.Reference.URI)
		}
		refsByFile[ref.Reference.URI] = append(refsByFile[ref.Reference.URI], ref.Reference)
	}
	sort.Slice(files, func(i, j int) bool { return files[i] < files[j] })

	maxLineMatches := 25
	if len(files) > int(query.FileMatchLimit) {
		files = files[:int(query.FileMatchLimit)]
		limitHit = true
	}

	matches = make([]*fileMatchResolver, len(files))
	for i, file := range files {
		fileRefs := refsByFile[file]
		if len(fileRefs) > maxLineMatches {
			fileRefs = fileRefs[:maxLineMatches]
		}
		lineMatches := make([]*lineMatch, 0, len(fileRefs))
		for _, fr := range fileRefs {
			lineMatches = append(lineMatches, &lineMatch{
				JLineNumber:       int32(fr.Range.Start.Line),
				JOffsetAndLengths: [][2]int32{{int32(fr.Range.Start.Character), int32(fr.Range.End.Character - fr.Range.Start.Character)}},
			})
		}
		uri, err := uri.Parse(string(file))
		if err != nil {
			return nil, false, err
		}
		matches[i] = &fileMatchResolver{
			JPath:        uri.FilePath(),
			JLineMatches: lineMatches,
			uri:          string(file),
			repo:         repo,
			commitID:     commit,
		}
	}
	return matches, limitHit, nil
}
