package git_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git/gittest"
)

func TestGetObject(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"echo x > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo           gitserver.Repo
		objectName     string
		wantOID        string
		wantObjectType git.ObjectType
	}{
		"basic": {
			repo:           gittest.MakeGitRepository(t, gitCommands...),
			objectName:     "e86b31b62399cfc86199e8b6e21a35e76d0e8b5e^{tree}",
			wantOID:        "a1dffc7a64c0b2d395484bf452e9aeb1da3a18f2",
			wantObjectType: git.ObjectTypeTree,
		},
	}

	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			oid, objectType, err := git.GetObject(ctx, test.repo, test.objectName)
			if err != nil {
				t.Fatal(err)
			}
			if oid.String() != test.wantOID {
				t.Errorf("got OID %q, want %q", oid, test.wantOID)
			}
			if objectType != test.wantObjectType {
				t.Errorf("got object type %q, want %q", objectType, test.wantObjectType)
			}
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_946(size int) error {
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
