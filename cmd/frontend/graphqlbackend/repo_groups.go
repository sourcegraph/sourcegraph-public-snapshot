package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type repoGroup struct {
	name         string
	repositories []api.RepoName
}

func (g repoGroup) Name() string { return g.name }

func (g repoGroup) Repositories() []string { return repoNamesToStrings(g.repositories) }

func (r *schemaResolver) RepoGroups(ctx context.Context) ([]*repoGroup, error) {
	groupsByName, err := resolveRepoGroups(ctx)
	if err != nil {
		return nil, err
	}

	groups := make([]*repoGroup, 0, len(groupsByName))
	for name, repos := range groupsByName {
		repoPaths := make([]api.RepoName, len(repos))
		for i, repo := range repos {
			repoPaths[i] = repo.Name
		}
		groups = append(groups, &repoGroup{
			name:         name,
			repositories: repoPaths,
		})
	}
	return groups, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_174(size int) error {
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
