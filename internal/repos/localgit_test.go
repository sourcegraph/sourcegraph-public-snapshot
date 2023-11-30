package repos

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestLocalGitSource_ListRepos(t *testing.T) {
	configs := []struct {
		pattern string
		group   string
		repos   []string
		folders []string
	}{
		{
			pattern: "projects/*",
			repos:   []string{"projects/a", "projects/b", "projects/c.bare"},
			folders: []string{"not-a-repo"},
		},
		{
			pattern: "work/*",
			group:   "work",
			repos:   []string{"work/a", "work/b", "work/c.bare"},
		},
		{
			pattern: "work*",
			repos:   []string{"work-project", "work-project2", "not-a-work-project"},
		},
		{
			pattern: "nested/*/*",
			repos:   []string{"nested/work/project", "nested/other-work/other-project"},
			folders: []string{"nested/work/not-a-project"},
		},
		{
			pattern: "single-repo",
			repos:   []string{"single-repo"},
		},
		{
			pattern: "with space",
			repos:   []string{"with space"},
		},
		{
			pattern: "no-match/*",
			repos:   []string{"single-repo"},
		},
	}

	repoPatterns := []*schema.LocalGitRepoPattern{}
	roots := []string{}

	for _, config := range configs {
		root := gitInitRepos(t, config.repos...)
		roots = append(roots, root)
		repoPatterns = append(repoPatterns, &schema.LocalGitRepoPattern{Pattern: filepath.Join(root, config.pattern), Group: config.group})
		for _, folder := range config.folders {
			if err := os.MkdirAll(filepath.Join(root, folder), 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}

	ctx := context.Background()

	svc := typestest.MakeExternalService(t, extsvc.VariantLocalGit, &schema.LocalGitExternalService{
		Repos: repoPatterns,
	})

	src, err := NewLocalGitSource(ctx, logtest.Scoped(t), svc)
	if err != nil {
		t.Fatal(err)
	}

	repos, err := ListAll(ctx, src)
	if err != nil {
		t.Fatal(err)
	}

	sort.SliceStable(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	// We need to replace the temporary folder, which changes between runs, with something static
	root_placeholder := "~root~"
	for _, repo := range repos {
		for _, root := range roots {
			if strings.Contains(repo.URI, root) {
				repo.URI = strings.Replace(repo.URI, root, root_placeholder, 1)
				repo.ExternalRepo.ID = strings.Replace(repo.ExternalRepo.ID, root, root_placeholder, 1)
				repo.ExternalRepo.ServiceID = strings.Replace(repo.ExternalRepo.ServiceID, root, root_placeholder, 1)
				for k := range repo.Sources {
					repo.Sources[k].CloneURL = strings.Replace(repo.Sources[k].CloneURL, root, root_placeholder, 1)
				}
				repo.Metadata.(*extsvc.LocalGitMetadata).AbsRepoPath = strings.Replace(repo.Metadata.(*extsvc.LocalGitMetadata).AbsRepoPath, root, root_placeholder, 1)
				break
			}
		}
	}

	testutil.AssertGolden(t, filepath.Join("testdata", "sources", t.Name()), Update(t.Name()), repos)
}

func Test_convertGitCloneURLToCodebaseName(t *testing.T) {
	testCases := []struct {
		cloneURL string
		expect   autogold.Value
	}{
		{"", autogold.Expect("")},
		{"https://github.com/sourcegraph/handbook", autogold.Expect("github.com/sourcegraph/handbook")},
		{"https://github.com/sourcegraph/handbook.git", autogold.Expect("github.com/sourcegraph/handbook")},
		{"git@github.com:sourcegraph/handbook", autogold.Expect("github.com/sourcegraph/handbook")},
		{"github:sourcegraph/handbook", autogold.Expect("github.com/sourcegraph/handbook")},

		// Note: this "git@github.com:/sourcegraph/handbook" URL format comes from the following
		// on Taylor's laptop:
		//
		//  git clone https://github.com/sourcegraph/handbook handbook-https
		//  cd handbook-https/
		//  git remote get-url origin
		//
		// No clue why an HTTPS URL gets translated into a git@github.com format (or why it has a leading slash)
		// but this exists in the wild so we should handle it ;)
		{"git@github.com:/sourcegraph/handbook", autogold.Expect("github.com/sourcegraph/handbook")},
	}
	for _, tc := range testCases {
		t.Run(tc.cloneURL, func(t *testing.T) {
			got := convertGitCloneURLToCodebaseName(tc.cloneURL)
			tc.expect.Equal(t, got)
		})
	}
}

func gitInitBare(t *testing.T, path string) {
	if err := exec.Command("git", "init", "--bare", path).Run(); err != nil {
		t.Fatal(err)
	}
}

func gitInit(t *testing.T, path string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func gitInitRepos(t *testing.T, names ...string) string {
	root := t.TempDir()
	root = filepath.Join(root, "repos-root")

	for _, name := range names {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}

		if strings.HasSuffix(p, ".bare") {
			gitInitBare(t, p)
		} else {
			gitInit(t, p)
		}
	}

	return root
}
