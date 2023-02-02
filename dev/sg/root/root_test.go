package root

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCreateSGHome(t *testing.T) {
	testHome := os.TempDir()
	actualHome, err := createSGHome(testHome)
	defer func() {
		os.Remove(actualHome)
	}()

	if err != nil {
		t.Fatalf("error creating SG Home dir(.sourcegraph) at %q: %q", testHome, err)
	}

	wantedHome := filepath.Join(testHome, ".sourcegraph")
	_, err = os.Stat(wantedHome)
	if err != nil {
		t.Errorf("failed to stat SG Home %q. Expected directory to be created\n", err)
	}
}

func TestWalkGitIgnoreFunc(t *testing.T) {

	tests := []struct {
		name            string
		gitIgnore       string
		additionalLines []string
		expectedFiles   []string
	}{
		{
			name: "empty gitignore + no additional lines",
			expectedFiles: []string{
				"foo.txt",

				"bar",
				"bar/baz.txt",

				".git",
				".git/qux.txt",

				".gitignore",
			},
		},

		{

			name: "gitignore: ignore baz.txt only",
			gitIgnore: `
bar/baz.txt
`,
			expectedFiles: []string{
				"foo.txt",

				"bar",

				".git",
				".git/qux.txt",

				".gitignore",
			},
		},

		{

			name: "gitignore: ignore bar folder entirely",
			gitIgnore: `
bar
`,
			expectedFiles: []string{
				"foo.txt",

				".git",
				".git/qux.txt",

				".gitignore",
			},
		},

		{

			name: "gitignore: ignore bar folder entirely / additional lines: ignore .git",
			gitIgnore: `
bar
`,
			additionalLines: []string{".git"},
			expectedFiles: []string{
				"foo.txt",

				".gitignore",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()

			// Setup: create file layout that looks like the following
			// 	- foo.txt
			// 	- bar/baz.txt
			// 	- .git/qux.txt
			// 	- .gitignore

			for _, f := range []struct {
				name     string
				contents string
			}{
				{name: "foo.txt", contents: "foo"},
				{name: "bar/baz.txt", contents: "baz"},
				{name: ".git/qux.txt", contents: "qux"},
				{name: ".gitignore", contents: test.gitIgnore},
			} {

				fileName := filepath.Join(root, f.name)

				dir := filepath.Dir(fileName)
				err := os.MkdirAll(dir, 0777)
				if err != nil {
					t.Fatalf("failed to create directory %q: %q", dir, err)
				}

				err = os.WriteFile(fileName, []byte(f.contents), 0777)
				if err != nil {
					t.Fatalf("failed to create file %q: %q", fileName, err)
				}
			}

			// Setup: prepare walkFunction that will record the names of all files and folders
			// that are visited.
			var actualFiles []string

			gatherWalkFn := func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(root, path)
				if err != nil {
					t.Fatalf("failed to calculate relative path for %q: %q", path, err)
				}

				if relPath == "." {
					// don't bother including the root directory
					return nil
				}

				actualFiles = append(actualFiles, relPath)
				return nil
			}

			// Test: call walkDir with the skipGitIgnoreWalkFunc wrapper
			gitignorePath := filepath.Join(root, ".gitignore")
			err := filepath.WalkDir(root, skipGitIgnoreWalkFunc(gatherWalkFn, gitignorePath, test.additionalLines...))
			if err != nil {
				t.Fatalf("failed to walk directory %q: %q", root, err)
			}

			// Examine: sort the actual and expected files so that we can compare them
			// to see if we recorded the set of files that we expected
			sort.Strings(actualFiles)
			sort.Strings(test.expectedFiles)

			if diff := cmp.Diff(test.expectedFiles, actualFiles); diff != "" {
				t.Errorf("unexpected files (-want +got):\n%s", diff)
			}
		})
	}
}
