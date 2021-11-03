package git

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestRead(t *testing.T) {
	t.Parallel()

	const wantData = "abcd\n"
	repo := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	ctx := context.Background()

	tests := map[string]struct {
		file     string
		maxBytes int64
		checkFn  func(*testing.T, error, []byte)
	}{
		"all": {
			file: "file1",
			checkFn: func(t *testing.T, err error, data []byte) {
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != wantData {
					t.Errorf("got %q, want %q", data, wantData)
				}
			},
		},

		"nonexistent": {
			file: "filexyz",
			checkFn: func(t *testing.T, err error, data []byte) {
				if err == nil {
					t.Fatal("err == nil")
				}
				if !os.IsNotExist(err) {
					t.Fatalf("got err %v, want os.IsNotExist", err)
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name+"-ReadFile", func(t *testing.T) {
			data, err := ReadFile(ctx, repo, commitID, test.file, test.maxBytes)
			test.checkFn(t, err, data)
		})
		t.Run(name+"-GetFileReader", func(t *testing.T) {
			rc, err := NewFileReader(ctx, repo, commitID, test.file)
			if err != nil {
				t.Fatal(err)
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			test.checkFn(t, err, data)
		})
	}

	t.Run("maxBytes", func(t *testing.T) {
		data, err := ReadFile(ctx, repo, commitID, "file1", 3)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != wantData[:3] {
			t.Errorf("got %q, want %q", data, wantData)
		}
	})
}

func TestReadEnforcesSubRepoPermissions(t *testing.T) {
	// Cannot run in parallel since we need a custom gitserver.DefaultClient
	oldClient := gitserver.DefaultClient
	t.Cleanup(func() {
		gitserver.DefaultClient = oldClient
	})

	newClient := gitserver.NewClient(oldClient.HTTPClient)
	newClient.Addrs = oldClient.Addrs
	mc := authz.NewMockSubRepoPermissionChecker()
	mc.CurrentUserPermissionsFunc.SetDefaultHook(func(ctx context.Context, content authz.RepoContent) (authz.Perms, error) {
		if content.Path == "file1" {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	newClient.SubRepoPermissionsChecker = mc
	gitserver.DefaultClient = newClient

	repo := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	ctx := context.Background()

	tests := map[string]struct {
		file    string
		allowed bool
		wantErr bool
	}{
		"allowed": {
			file:    "file1",
			wantErr: false,
		},
		"not allowed": {
			file:    "notallowed",
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name+"-ReadFile", func(t *testing.T) {
			_, err := ReadFile(ctx, repo, commitID, test.file, 0)
			if !test.wantErr && err != nil {
				t.Fatal(err)
			}
			if err == nil && test.wantErr {
				t.Fatal("wanted an error")
			}
		})
		t.Run(name+"-GetFileReader", func(t *testing.T) {
			rc, err := NewFileReader(ctx, repo, commitID, test.file)
			if !test.wantErr && err != nil {
				t.Fatal(err)
			}
			if err == nil && test.wantErr {
				t.Fatal("wanted an error")
			}
			if err != nil {
				return
			}
			defer rc.Close()
			_, err = io.ReadAll(rc)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
