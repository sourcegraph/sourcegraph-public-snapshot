package resolvers

import (
	"context"
	"io/fs"
	"path"
	"sort"
	"sync"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type ownerData struct {
	Owner     string
	FileCount int // count of owned files
}

func (r *catalogComponentResolver) Owners(ctx context.Context) (*[]gql.CatalogEntityOwnerEdgeResolver, error) {
	var allEntries []fs.FileInfo
	for _, sourcePath := range r.component.SourcePaths {
		// TODO(sqs): doesnt check perms? SECURITY
		entries, err := git.ReadDir(ctx, r.component.SourceRepo, r.component.SourceCommit, sourcePath, true)
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, entries...)
	}

	// Read all owners files.
	possibleCodeownersFiles := []string{".github/CODEOWNERS", "CODENOTIFY"} // root files
	isCodeownersFile := func(fullPath string) bool {
		name := path.Base(fullPath)
		return name == "CODEOWNERS" || name == "CODENOTIFY"
	}
	for _, e := range allEntries {
		if isCodeownersFile(e.Name()) {
			possibleCodeownersFiles = append(possibleCodeownersFiles, e.Name())
		}
	}
	var (
		mu            sync.Mutex
		codeowners    codeownersComputer
		codeownersErr error
		wg            sync.WaitGroup
	)
	for _, p := range possibleCodeownersFiles {
		if codeowners.has(r.component.SourceRepo, r.component.SourceCommit, p) {
			continue
		}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			// TODO(sqs): doesnt check perms? SECURITY
			data, err := git.ReadFile(ctx, r.component.SourceRepo, r.component.SourceCommit, p, 0)
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
			if err := codeowners.add(r.component.SourceRepo, r.component.SourceCommit, p, data); err != nil {
				if codeownersErr == nil {
					codeownersErr = err
				}
			}
		}(p)
	}
	wg.Wait()
	if codeownersErr != nil {
		return nil, codeownersErr
	}

	var (
		byOwner        = map[string]*ownerData{}
		totalFileCount int
	)
	for _, e := range allEntries {
		if e.IsDir() || isCodeownersFile(e.Name()) {
			continue
		}

		totalFileCount++
		owners := codeowners.get(r.component.SourceRepo, r.component.SourceCommit, e.Name())
		for _, owner := range owners {
			od := byOwner[owner]
			if od == nil {
				od = &ownerData{Owner: owner}
				byOwner[owner] = od
			}
			od.FileCount++
		}
	}

	edges := make([]gql.CatalogEntityOwnerEdgeResolver, 0, len(byOwner))
	for _, od := range byOwner {
		edges = append(edges, &catalogEntityOwnerEdgeResolver{
			db:             r.db,
			data:           od,
			totalFileCount: totalFileCount,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].FileCount() > edges[j].FileCount()
	})

	return &edges, nil
}

type catalogEntityOwnerEdgeResolver struct {
	db             database.DB
	data           *ownerData
	totalFileCount int
}

func (r *catalogEntityOwnerEdgeResolver) Node() string     { return r.data.Owner }
func (r *catalogEntityOwnerEdgeResolver) FileCount() int32 { return int32(r.data.FileCount) }
func (r *catalogEntityOwnerEdgeResolver) FileProportion() float64 {
	return float64(r.data.FileCount) / float64(r.totalFileCount)
}
