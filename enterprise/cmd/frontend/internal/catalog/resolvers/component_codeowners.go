package resolvers

import (
	"context"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type codeOwnerData struct {
	Owner     string
	FileCount int // count of owned files
}

func (r *componentResolver) CodeOwners(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.CodeOwnerConnectionResolver, error) {
	slocs, err := r.sourceSetResolver(ctx)
	if err != nil {
		return nil, err
	}
	return slocs.CodeOwners(ctx, args)
}

func (r *rootResolver) GitTreeEntryCodeOwners(ctx context.Context, treeEntry *gql.GitTreeEntryResolver, args *graphqlutil.ConnectionArgs) (gql.CodeOwnerConnectionResolver, error) {
	return sourceSetResolverFromTreeEntry(treeEntry, r.db).CodeOwners(ctx, args)
}

func (r *sourceSetResolver) CodeOwners(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.CodeOwnerConnectionResolver, error) {
	allFiles, err := r.allFiles(ctx)
	if err != nil {
		return nil, err
	}

	// Read all owners files.
	var possibleCodeownersFiles []fileInfo

	// All repositories that contain the component's source locations might have an owners file at
	// the root.
	rootFilenames := []string{".github/CODEOWNERS", "CODENOTIFY"} // root files
	for _, sloc := range r.slocs {
		for _, rootFilename := range rootFilenames {
			possibleCodeownersFiles = append(possibleCodeownersFiles, fileInfo{
				FileInfo: gql.CreateFileInfo(rootFilename, false),
				repo:     sloc.repoName,
				commit:   sloc.commitID,
			})
		}
	}

	isCodeownersFile := func(fullPath string) bool {
		name := path.Base(fullPath)
		return name == "CODEOWNERS" || name == "CODENOTIFY"
	}
	for _, f := range allFiles {
		if isCodeownersFile(f.Name()) {
			possibleCodeownersFiles = append(possibleCodeownersFiles, f)
		}
	}
	var (
		mu            sync.Mutex
		codeowners    codeownersComputer
		codeownersErr error
		wg            sync.WaitGroup
	)
	for _, f := range possibleCodeownersFiles {
		if codeowners.has(f.repo, f.commit, f.Name()) {
			continue
		}
		wg.Add(1)
		go func(f fileInfo) {
			defer wg.Done()

			// TODO(sqs): doesnt check perms? SECURITY
			data, err := git.ReadFile(ctx, f.repo, f.commit, f.Name(), 0, authz.DefaultSubRepoPermsChecker)
			if os.IsNotExist(err) {
				// TODO(sqs): this is probably one of the rootFilenames we tried, but check to make
				// sure it is.
				return
			}
			if err != nil {
				mu.Lock()
				if codeownersErr == nil {
					codeownersErr = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if err := codeowners.add(f.repo, f.commit, f.Name(), data); err != nil {
				if codeownersErr == nil {
					codeownersErr = err
				}
			}
		}(f)
	}
	wg.Wait()
	if codeownersErr != nil {
		return nil, codeownersErr
	}

	var (
		byOwner        = map[string]*codeOwnerData{}
		totalFileCount int
	)
	for _, f := range allFiles {
		if f.IsDir() || isCodeownersFile(f.Name()) {
			continue
		}

		totalFileCount++
		owners := codeowners.get(f.repo, f.commit, f.Name())
		for _, owner := range owners {
			od := byOwner[owner]
			if od == nil {
				od = &codeOwnerData{Owner: owner}
				byOwner[owner] = od
			}
			od.FileCount++
		}
	}

	edges := make([]gql.ComponentCodeOwnerEdgeResolver, 0, len(byOwner))
	for _, od := range byOwner {
		edges = append(edges, &componentCodeOwnerEdgeResolver{
			db:             r.db,
			data:           od,
			totalFileCount: totalFileCount,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		ei, ej := edges[i], edges[j]
		if ei.FileCount() != ej.FileCount() {
			return ei.FileCount() > ej.FileCount()
		}
		return ei.Node().Email() < ej.Node().Email()
	})

	return &codeOwnerConnectionResolver{edges: edges, first: args.First}, nil
}

type componentCodeOwnerEdgeResolver struct {
	db             database.DB
	data           *codeOwnerData
	totalFileCount int
}

func (r *componentCodeOwnerEdgeResolver) Node() *gql.PersonResolver {
	return gql.NewPersonResolver(r.db, strings.TrimPrefix(r.data.Owner, "@"), strings.TrimPrefix(r.data.Owner, "@")+"@sourcegraph.com", false)
}
func (r *componentCodeOwnerEdgeResolver) FileCount() int32 { return int32(r.data.FileCount) }
func (r *componentCodeOwnerEdgeResolver) FileProportion() float64 {
	return float64(r.data.FileCount) / float64(r.totalFileCount)
}

type codeOwnerConnectionResolver struct {
	edges []gql.ComponentCodeOwnerEdgeResolver
	first *int32
}

func (r *codeOwnerConnectionResolver) Edges() []gql.ComponentCodeOwnerEdgeResolver {
	if r.first != nil && len(r.edges) > int(*r.first) {
		return r.edges[:int(*r.first)]
	}
	return r.edges
}

func (r *codeOwnerConnectionResolver) TotalCount() int32 {
	return int32(len(r.edges))
}

func (r *codeOwnerConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(r.first != nil && int(*r.first) < len(r.edges))
}
