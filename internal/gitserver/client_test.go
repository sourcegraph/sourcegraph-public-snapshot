package gitserver_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/migration"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestClient_ListCloned(t *testing.T) {
	addrs := []string{"gitserver-0", "gitserver-1"}
	cli := gitserver.NewTestClient(
		httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.String() {
			case "http://gitserver-0/list?cloned":
				return &http.Response{
					Body: io.NopCloser(bytes.NewBufferString(`["repo0-a", "repo0-b"]`)),
				}, nil
			case "http://gitserver-1/list?cloned":
				return &http.Response{
					Body: io.NopCloser(bytes.NewBufferString(`["repo1-a", "repo1-b"]`)),
				}, nil
			default:
				return nil, errors.Errorf("unexpected url: %s", r.URL.String())
			}
		}),
		database.NewMockDB(),
		addrs,
	)

	want := []string{"repo0-a", "repo1-a", "repo1-b"}
	got, err := cli.ListCloned(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	sort.Strings(want)
	if !cmp.Equal(want, got, cmpopts.EquateEmpty()) {
		t.Errorf("mismatch for (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestClient_RequestRepoMigrate(t *testing.T) {
	repo := api.RepoName("github.com/sourcegraph/sourcegraph")
	addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080"}

	expected := "http://172.16.8.2:8080"

	cli := gitserver.NewTestClient(
		httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.String() {
			// Ensure that the request was received by the "expected" gitserver instance - where
			// expected is the gitserver instance according to the Rendezvous hashing scheme.
			// For anything else apart from this we return an error.
			case expected + "/repo-update":
				var req protocol.RepoUpdateRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				if err != nil {
					t.Fatal(err)
				}
				if req.CloneFromShard != "http://172.16.8.1:8080" {
					t.Fatalf("expected cloneFromShard to be \"http://172.16.8.1:8080\", got %q", req.CloneFromShard)
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString("{}")),
				}, nil
			default:
				return nil, errors.Newf("unexpected URL: %q", r.URL.String())
			}
		}),
		database.NewMockDB(),
		addrs,
	)

	_, err := cli.RequestRepoMigrate(context.Background(), repo, "172.16.8.1:8080", "172.16.8.2:8080")
	if err != nil {
		t.Fatalf("expected URL %q, but got err %q", expected, err)
	}
}

func TestClient_Remove(t *testing.T) {
	repo := api.RepoName("github.com/sourcegraph/sourcegraph")
	addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080"}

	expected := "http://172.16.8.1:8080"

	cli := gitserver.NewTestClient(
		httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.String() {
			// Ensure that the request was received by the "expected" gitserver instance - where
			// expected is the gitserver instance according to the Rendezvous hashing scheme.
			// For anything else apart from this we return an error.
			case expected + "/delete":
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString("{}")),
				}, nil
			default:
				return nil, errors.Newf("unexpected URL: %q", r.URL.String())
			}
		}),
		database.NewMockDB(),
		addrs,
	)

	err := cli.Remove(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected URL %q, but got err %q", expected, err)
	}

	err = cli.RemoveFrom(context.Background(), repo, "172.16.8.1:8080")
	if err != nil {
		t.Fatalf("expected URL %q, but got err %q", expected, err)
	}
}

func TestClient_Archive(t *testing.T) {
	root := gitserver.CreateRepoDir(t)

	tests := map[api.RepoName]struct {
		remote string
		want   map[string]string
		err    error
	}{
		"simple": {
			remote: createSimpleGitRepo(t, root),
			want: map[string]string{
				"dir1/":      "",
				"dir1/file1": "infile1",
				"file 2":     "infile2",
			},
		},
		"repo-with-dotgit-dir": {
			remote: createRepoWithDotGitDir(t, root),
			want:   map[string]string{"file1": "hello\n", ".git/mydir/file2": "milton\n", ".git/mydir/": "", ".git/": ""},
		},
		"not-found": {
			err: errors.New("repository does not exist: not-found"),
		},
	}

	srv := httptest.NewServer((&server.Server{
		ReposDir: filepath.Join(root, "repos"),
		GetRemoteURLFunc: func(_ context.Context, name api.RepoName) (string, error) {
			testData := tests[name]
			if testData.remote != "" {
				return testData.remote, nil
			}
			return "", errors.Errorf("no remote for %s", name)
		},
		GetVCSSyncer: func(ctx context.Context, name api.RepoName) (server.VCSSyncer, error) {
			return &server.GitRepoSyncer{}, nil
		},
	}).Handler())
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	addrs := []string{u.Host}
	cli := gitserver.NewTestClient(&http.Client{}, database.NewMockDB(), addrs)

	ctx := context.Background()
	for name, test := range tests {
		t.Run(string(name), func(t *testing.T) {
			if test.remote != "" {
				if _, err := cli.RequestRepoUpdate(ctx, name, 0); err != nil {
					t.Fatal(err)
				}
			}

			rc, err := cli.Archive(ctx, name, gitserver.ArchiveOptions{Treeish: "HEAD", Format: "zip"})
			if have, want := fmt.Sprint(err), fmt.Sprint(test.err); have != want {
				t.Errorf("archive: have err %v, want %v", have, want)
			}

			if rc == nil {
				return
			}

			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				t.Fatal(err)
			}
			zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}

			got := map[string]string{}
			for _, f := range zr.File {
				r, err := f.Open()
				if err != nil {
					t.Errorf("failed to open %q because %s", f.Name, err)
					continue
				}
				contents, err := io.ReadAll(r)
				_ = r.Close()
				if err != nil {
					t.Errorf("Read(%q): %s", f.Name, err)
					continue
				}
				got[f.Name] = string(contents)
			}

			if !cmp.Equal(test.want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(test.want, got))
			}
		})
	}
}

func createRepoWithDotGitDir(t *testing.T, root string) string {
	t.Helper()
	b64 := func(s string) string {
		t.Helper()
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	dir := filepath.Join(root, "remotes", "repo-with-dot-git-dir")

	// This repo was synthesized by hand to contain a file whose path is `.git/mydir/file2` (the Git
	// CLI will not let you create a file with a `.git` path component).
	//
	// The synthesized bad commit is:
	//
	// commit aa600fc517ea6546f31ae8198beb1932f13b0e4c (HEAD -> master)
	// Author: Quinn Slack <qslack@qslack.com>
	// 	Date:   Tue Jun 5 16:17:20 2018 -0700
	//
	// wip
	//
	// diff --git a/.git/mydir/file2 b/.git/mydir/file2
	// new file mode 100644
	// index 0000000..82b919c
	// --- /dev/null
	// +++ b/.git/mydir/file2
	// @@ -0,0 +1 @@
	// +milton
	files := map[string]string{
		"config": `
[core]
repositoryformatversion=0
filemode=true
`,
		"HEAD":              `ref: refs/heads/master`,
		"refs/heads/master": `aa600fc517ea6546f31ae8198beb1932f13b0e4c`,
		"objects/e7/9c5e8f964493290a409888d5413a737e8e5dd5": b64("eAFLyslPUrBgyMzLLMlMzOECACgtBOw="),
		"objects/ce/013625030ba8dba906f756967f9e9ca394464a": b64("eAFLyslPUjBjyEjNycnnAgAdxQQU"),
		"objects/82/b919c9c565d162c564286d9d6a2497931be47e": b64("eAFLyslPUjBnyM3MKcnP4wIAIw8ElA=="),
		"objects/e5/231c1d547df839dce09809e43608fe6c537682": b64("eAErKUpNVTAzYTAxAAIFvfTMEgbb8lmsKdJ+zz7ukeMOulcqZqOllmloYGBmYqKQlpmTashwjtFMlZl7xe2VbN/DptXPm7N4ipsXACOoGDo="),
		"objects/da/5ecc846359eaf23e8abe907b3125fdd7abdbc0": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWJo2il58mjqxaSjKRq5c7NUpk+WflIHABZRD2I="),
		"objects/d0/01d287018593691c36042e1c8089fde7415296": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWQ4x2imysy94vZKtu9h0+rnzVk8xc0LAP2TDiQ="),
		"objects/b4/009ecbf1eba01c5279f25840e2afc0d15f5005": b64("eAGdjdsJAjEQRf1OFdOAMpPN5gEitiBWEJIRBzcJu2b7N2IHfh24nMtJrRTpQA4PfWOGjEhZe4fk5zDZQGmyaDRT8ujDI7MzNOtgVdz7s21w26VWuC8xveC8vr+8/nBKrVxgyF4bJBfgiA5RjXUEO/9xVVKlS1zUB/JxNbA="),
		"objects/3d/779a05641b4ee6f1bc1e0b52de75163c2a2669": b64("eAErKUpNVTA2YjAxAAKF3MqUzCKGW3FnWpIjX32y69o3odpQ9e/11bcPAAAipRGQ"),
		"objects/aa/600fc517ea6546f31ae8198beb1932f13b0e4c": b64("eAGdjlkKAjEQBf3OKfoCSmfpLCDiFcQTZDodHHQWxwxe3xFv4FfBKx4UT8PQNzDa7doiAkLGataFXCg12lRYMEVM4qzHWMUz2eCjUXNeZGzQOdwkd1VLl1EzmZCqoehQTK6MRVMlRFJ5bbdpgcvajyNcH5nvcHy+vjz/cOBpOIEmE41D7xD2GBDVtm6BTf64qnc/qw9c4UKS"),
		"objects/e6/9de29bb2d1d6434b8b29ae775ad8c2e48c5391": b64("eAFLyslPUjBgAAAJsAHw"),
	}
	for name, data := range files {
		name = filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func createSimpleGitRepo(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "remotes", "simple")

	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	for _, cmd := range []string{
		"git init",
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t 200601021704.05 dir1 dir1/file1",
		"git add dir1/file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --date=2014-05-06T19:20:21Z 'file 2' || touch -t 201405062120.21 'file 2'",
		"git add 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
		"git branch test-ref HEAD~1",
		"git branch test-nested-ref test-ref",
	} {
		c := exec.Command("bash", "-c", `GIT_CONFIG_GLOBAL="" GIT_CONFIG_SYSTEM="" `+cmd)
		c.Dir = dir
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}

	return dir
}

func TestAddrForRepo(t *testing.T) {
	addrs := []string{"gitserver-1", "gitserver-2", "gitserver-3"}
	pinned := map[string]string{
		"repo2": "gitserver-1",
	}

	testCases := []struct {
		name string
		repo api.RepoName
		want string
	}{
		{
			name: "repo1",
			repo: api.RepoName("repo1"),
			want: "gitserver-3",
		},
		{
			name: "check we normalise",
			repo: api.RepoName("repo1.git"),
			want: "gitserver-3",
		},
		{
			name: "another repo",
			repo: api.RepoName("github.com/sourcegraph/sourcegraph.git"),
			want: "gitserver-2",
		},
		{
			name: "pinned repo", // different server address that the hashing function would normally yield
			repo: api.RepoName("repo2"),
			want: "gitserver-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := gitserver.AddrForRepo(context.Background(), "gitserver", database.NewMockDB(), tc.repo, gitserver.GitServerAddresses{
				Addresses:     addrs,
				PinnedServers: pinned,
			})
			if err != nil {
				t.Fatal("Error during getting gitserver address")
			}
			if got != tc.want {
				t.Fatalf("Want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestAddrForRepoFromDB(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// enable feature flag
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGitserverClientLookupTable: true,
				// Set the rate at 100% for this test
				GitserverClientLookupTableRate: 100,
			},
		},
	})
	defer conf.Mock(nil)

	// enable dotCom mode
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(false)

	addrs := []string{"gitserver-1", "gitserver-2", "gitserver-3"}
	pinned := map[string]string{
		"github.com/sourcegraph/repo2": "gitserver-1",
	}

	testCases := []struct {
		name           string
		repo           api.RepoName
		want           string
		dotComDisabled bool
		dbNotCalled    bool
	}{
		{
			name: "repo1",
			repo: api.RepoName("github.com/sourcegraph/repo1"),
			want: "gitserver-2",
		},
		{
			name: "normalisation",
			repo: api.RepoName("github.com/sourcegraph/repo1.git"),
			want: "gitserver-2",
		},
		{
			name: "repo not in the DB",
			repo: api.RepoName("github.com/sourcegraph/sourcegraph.git"),
			want: "gitserver-2",
		},
		{
			name:        "pinned repo", // different server address that the hashing function would normally yield
			repo:        api.RepoName("github.com/sourcegraph/repo2"),
			want:        "gitserver-1",
			dbNotCalled: true,
		},
		{
			name:           "repo1",
			repo:           api.RepoName("github.com/sourcegraph/repo1"),
			want:           "gitserver-2",
			dotComDisabled: true,
			dbNotCalled:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.dotComDisabled {
				envvar.MockSourcegraphDotComMode(false)
			} else {
				envvar.MockSourcegraphDotComMode(true)
			}

			db := database.NewMockDB()

			store := database.NewMockGitserverRepoStore()
			store.GetByNameFunc.SetDefaultReturn(&types.GitserverRepo{
				RepoID:        1,
				ShardID:       "gitserver-1",
				CloneStatus:   types.CloneStatusNotCloned,
				RepoSizeBytes: 100,
			}, nil)
			db.GitserverReposFunc.SetDefaultReturn(store)

			got, err := gitserver.AddrForRepo(context.Background(), "gitserver", db, tc.repo, gitserver.GitServerAddresses{
				Addresses:     addrs,
				PinnedServers: pinned,
			})
			if err != nil {
				t.Fatal("Error during getting gitserver address")
			}
			if got != tc.want {
				t.Fatalf("Want %q, got %q", tc.want, got)
			}
			if tc.dbNotCalled && len(db.GitserverReposFunc.History()) > 0 {
				t.Fatalf("Should not have called the database")
			} else if !tc.dbNotCalled && len(db.GitserverReposFunc.History()) == 0 {
				t.Fatalf("Should have called the database")
			}
		})
	}
}

func TestAddrForRepoFromDBRates(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	gitserver.AddrForRepoCounter = 0

	// enable dotCom mode
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(false)

	addrs := []string{"gitserver-1", "gitserver-2", "gitserver-3"}
	pinned := map[string]string{
		"github.com/sourcegraph/repo2": "gitserver-1",
	}

	defer conf.Mock(nil)

	// enable feature flag with 50% rate
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGitserverClientLookupTable: true,
				GitserverClientLookupTableRate:   50,
			},
		},
	})

	check := func(repo string, dbCalled bool) {
		t.Helper()

		db := database.NewMockDB()
		store := database.NewMockGitserverRepoStore()
		store.GetByNameFunc.SetDefaultReturn(&types.GitserverRepo{
			RepoID:        1,
			ShardID:       "gitserver-1",
			CloneStatus:   types.CloneStatusNotCloned,
			RepoSizeBytes: 100,
		}, nil)
		db.GitserverReposFunc.SetDefaultReturn(store)

		_, err := gitserver.AddrForRepo(context.Background(), "gitserver", db, api.RepoName(repo), gitserver.GitServerAddresses{
			Addresses:     addrs,
			PinnedServers: pinned,
		})
		if err != nil {
			t.Fatal("Error during getting gitserver address")
		}
		if dbCalled && len(db.GitserverReposFunc.History()) == 0 {
			t.Fatalf("Should have called the database")
		} else if !dbCalled && len(db.GitserverReposFunc.History()) > 0 {
			t.Fatalf("Should not have called the database")
		}
	}

	// first request should be served from the hash
	// Rate: 50%, Counter: 1, Mod: 100 / 50 = 2, Result: 1 % 2 = 1
	check("github.com/sourcegraph/repo1", false)

	// second request should be served from the DB
	// Rate: 50%, Counter: 2, Mod: 100 / 50 = 2, Result: 2 % 2 = 0
	check("github.com/sourcegraph/repo1", true)

	// third request should be served from the hash
	// Rate: 50%, Counter: 3, Mod: 100 / 50 = 2, Result: 3 % 2 = 1
	check("github.com/sourcegraph/repo1", false)

	// Let's change the rate to 30%
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGitserverClientLookupTable: true,
				GitserverClientLookupTableRate:   30,
			},
		},
	})

	// next request should be served from the hash
	// Rate: 30%, Counter: 4, Mod: 100 / 30 = 3, Result: 4 % 3 = 1
	check("github.com/sourcegraph/repo1", false)

	// next request should be served from the hash
	// Rate: 30%, Counter: 5, Mod: 100 / 30 = 3, Result: 5 % 3 = 2
	check("github.com/sourcegraph/repo1", false)

	// next request should be served from the DB
	// Rate: 30%, Counter: 6, Mod: 100 / 30 = 3, Result: 6 % 3 = 0
	check("github.com/sourcegraph/repo1", true)
}

func TestRendezvousAddrForRepo(t *testing.T) {
	addrs := []string{"gitserver-1", "gitserver-2", "gitserver-3"}

	testCases := []struct {
		name string
		repo api.RepoName
		want string
	}{
		{
			name: "repo1",
			repo: api.RepoName("repo1"),
			want: "gitserver-1",
		},
		{
			name: "check we normalise",
			repo: api.RepoName("repo1.git"),
			want: "gitserver-1",
		},
		{
			name: "another repo",
			repo: api.RepoName("github.com/sourcegraph/sourcegraph.git"),
			want: "gitserver-3",
		},
		{
			name: "yet another repo",
			repo: api.RepoName("gitlab.com/foo/bar"),
			want: "gitserver-2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := gitserver.RendezvousAddrForRepo(tc.repo, addrs)
			if got != tc.want {
				t.Fatalf("Want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestClient_P4Exec(t *testing.T) {
	_ = gitserver.CreateRepoDir(t)
	tests := []struct {
		name     string
		host     string
		user     string
		password string
		args     []string
		handler  http.HandlerFunc
		wantBody string
		wantErr  string
	}{
		{
			name:     "check request body",
			host:     "ssl:111.222.333.444:1666",
			user:     "admin",
			password: "pa$$word",
			args:     []string{"protects"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				wantBody := `{"p4port":"ssl:111.222.333.444:1666","p4user":"admin","p4passwd":"pa$$word","args":["protects"]}`
				if diff := cmp.Diff(wantBody, string(body)); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("example output"))
			},
			wantBody: "example output",
			wantErr:  "<nil>",
		},
		{
			name: "error response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("example error"))
			},
			wantErr: "unexpected status code: 400 - example error",
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			u, _ := url.Parse(server.URL)
			addrs := []string{u.Host}
			cli := gitserver.NewTestClient(&http.Client{}, database.NewMockDB(), addrs)

			rc, _, err := cli.P4Exec(ctx, test.host, test.user, test.password, test.args...)
			if diff := cmp.Diff(test.wantErr, fmt.Sprintf("%v", err)); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}

			var body []byte
			if rc != nil {
				defer func() { _ = rc.Close() }()

				body, err = io.ReadAll(rc)
				if err != nil {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(test.wantBody, string(body)); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ResolveRevisions(t *testing.T) {
	root := t.TempDir()
	remote := createSimpleGitRepo(t, root)
	// These hashes should be stable since we set the timestamps
	// when creating the commits.
	hash1 := "b6602ca96bdc0ab647278577a3c6edcb8fe18fb0"
	hash2 := "c5151eceb40d5e625716589b745248e1a6c6228d"

	tests := []struct {
		input []protocol.RevisionSpecifier
		want  []string
		err   error
	}{{
		input: []protocol.RevisionSpecifier{{}},
		want:  []string{hash2},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "HEAD"}},
		want:  []string{hash2},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "HEAD~1"}},
		want:  []string{hash1},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-ref"}},
		want:  []string{hash1},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-nested-ref"}},
		want:  []string{hash1},
	}, {
		input: []protocol.RevisionSpecifier{{RefGlob: "refs/heads/test-*"}},
		want:  []string{hash1, hash1}, // two hashes because to refs point to that hash
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-fake-ref"}},
		err:   &gitdomain.RevisionNotFoundError{Repo: api.RepoName(remote), Spec: "test-fake-ref"},
	}}

	srv := httptest.NewServer((&server.Server{
		ReposDir: filepath.Join(root, "repos"),
		GetRemoteURLFunc: func(_ context.Context, name api.RepoName) (string, error) {
			return remote, nil
		},
		GetVCSSyncer: func(ctx context.Context, name api.RepoName) (server.VCSSyncer, error) {
			return &server.GitRepoSyncer{}, nil
		},
	}).Handler())
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	addrs := []string{u.Host}
	cli := gitserver.NewTestClient(&http.Client{}, database.NewMockDB(), addrs)

	ctx := context.Background()
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			_, err := cli.RequestRepoUpdate(ctx, api.RepoName(remote), 0)
			require.NoError(t, err)

			got, err := cli.ResolveRevisions(ctx, api.RepoName(remote), test.input)
			if test.err != nil {
				require.Equal(t, test.err, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}

}

func TestClient_AddrForRepo_UsesConfToRead_PinnedRepos(t *testing.T) {
	ctx := context.Background()
	client := gitserver.NewTestClient(&http.Client{}, database.NewMockDB(), []string{"gitserver1", "gitserver2"})
	setPinnedRepos(map[string]string{
		"repo1": "gitserver2",
	})

	addr, err := client.AddrForRepo(ctx, "repo1")
	if err != nil {
		t.Fatal("Error during getting gitserver address")
	}
	require.Equal(t, "gitserver2", addr)

	// simulate config change - site admin manually changes the pinned repo config
	setPinnedRepos(map[string]string{
		"repo1": "gitserver1",
	})

	addr, err = client.AddrForRepo(ctx, "repo1")
	if err != nil {
		t.Fatal("Error during getting gitserver address")
	}
	require.Equal(t, "gitserver1", addr)
}

func setPinnedRepos(pinned map[string]string) {
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			GitServerPinnedRepos: pinned,
		},
	}})
}

func TestClient_AddrForRepo_Rendezvous(t *testing.T) {
	ctx := context.Background()
	client := gitserver.NewTestClient(&http.Client{}, database.NewMockDB(), []string{"gitserver1", "gitserver2"})

	tests := []struct {
		name     string
		repoName api.RepoName
		cursor   string
		wantAddr string
	}{
		{
			name:     "Rendezvous hashing is not used before migration",
			repoName: api.RepoName("repoA"),
			cursor:   "",
			wantAddr: "gitserver1",
		},
		{
			name:     "Rendezvous hashing is not used for not yet migrated repos",
			repoName: api.RepoName("repoA"),
			cursor:   "repo",
			wantAddr: "gitserver1",
		},
		{
			name:     "Rendezvous hashing is used for already migrated repos",
			repoName: api.RepoName("repoA"),
			cursor:   "repoZ",
			wantAddr: "gitserver2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			migration.MigrationMocks.GetCursor = func(ctx context.Context, db dbutil.DB) (string, error) {
				return tc.cursor, nil
			}
			defer migration.ResetMigrationMocks()

			addr, err := client.AddrForRepo(ctx, tc.repoName)
			if err != nil {
				t.Fatal("Error during getting gitserver address")
			}
			require.Equal(t, tc.wantAddr, addr)
		})
	}
}

func TestClient_BatchLog(t *testing.T) {
	addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080", "172.16.8.3:8080"}

	cli := gitserver.NewTestClient(
		httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
			var req protocol.BatchLogRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}

			var results []protocol.BatchLogResult
			for _, repoCommit := range req.RepoCommits {
				results = append(results, protocol.BatchLogResult{
					RepoCommit:    repoCommit,
					CommandOutput: fmt.Sprintf("out<%s: %s@%s>", r.URL.String(), repoCommit.Repo, repoCommit.CommitID),
					CommandError:  "",
				})
			}

			encoded, _ := json.Marshal(protocol.BatchLogResponse{Results: results})
			body := io.NopCloser(strings.NewReader(strings.TrimSpace(string(encoded))))
			return &http.Response{StatusCode: 200, Body: body}, nil
		}),
		database.NewMockDB(),
		addrs,
	)

	opts := gitserver.BatchLogOptions{
		RepoCommits: []api.RepoCommit{
			{Repo: api.RepoName("github.com/test/foo"), CommitID: api.CommitID("deadbeef01")},
			{Repo: api.RepoName("github.com/test/bar"), CommitID: api.CommitID("deadbeef02")},
			{Repo: api.RepoName("github.com/test/baz"), CommitID: api.CommitID("deadbeef03")},
			{Repo: api.RepoName("github.com/test/bonk"), CommitID: api.CommitID("deadbeef04")},
			{Repo: api.RepoName("github.com/test/quux"), CommitID: api.CommitID("deadbeef05")},
			{Repo: api.RepoName("github.com/test/honk"), CommitID: api.CommitID("deadbeef06")},
			{Repo: api.RepoName("github.com/test/xyzzy"), CommitID: api.CommitID("deadbeef07")},
			{Repo: api.RepoName("github.com/test/lorem"), CommitID: api.CommitID("deadbeef08")},
			{Repo: api.RepoName("github.com/test/ipsum"), CommitID: api.CommitID("deadbeef09")},
			{Repo: api.RepoName("github.com/test/fnord"), CommitID: api.CommitID("deadbeef10")},
		},
		Format: "--format=test",
	}

	results := map[api.RepoCommit]gitserver.RawBatchLogResult{}
	var mu sync.Mutex

	if err := cli.BatchLog(context.Background(), opts, func(repoCommit api.RepoCommit, gitLogResult gitserver.RawBatchLogResult) error {
		mu.Lock()
		defer mu.Unlock()

		results[repoCommit] = gitLogResult
		return nil
	}); err != nil {
		t.Fatalf("unexpected error performing batch log: %s", err)
	}

	expectedResults := map[api.RepoCommit]gitserver.RawBatchLogResult{
		// Shard 1
		{Repo: "github.com/test/baz", CommitID: "deadbeef03"}:  {Stdout: "out<http://172.16.8.1:8080/batch-log: github.com/test/baz@deadbeef03>"},
		{Repo: "github.com/test/quux", CommitID: "deadbeef05"}: {Stdout: "out<http://172.16.8.1:8080/batch-log: github.com/test/quux@deadbeef05>"},
		{Repo: "github.com/test/honk", CommitID: "deadbeef06"}: {Stdout: "out<http://172.16.8.1:8080/batch-log: github.com/test/honk@deadbeef06>"},

		// Shard 2
		{Repo: "github.com/test/bar", CommitID: "deadbeef02"}:   {Stdout: "out<http://172.16.8.2:8080/batch-log: github.com/test/bar@deadbeef02>"},
		{Repo: "github.com/test/xyzzy", CommitID: "deadbeef07"}: {Stdout: "out<http://172.16.8.2:8080/batch-log: github.com/test/xyzzy@deadbeef07>"},

		// Shard 3
		{Repo: "github.com/test/foo", CommitID: "deadbeef01"}:   {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/foo@deadbeef01>"},
		{Repo: "github.com/test/bonk", CommitID: "deadbeef04"}:  {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/bonk@deadbeef04>"},
		{Repo: "github.com/test/lorem", CommitID: "deadbeef08"}: {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/lorem@deadbeef08>"},
		{Repo: "github.com/test/ipsum", CommitID: "deadbeef09"}: {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/ipsum@deadbeef09>"},
		{Repo: "github.com/test/fnord", CommitID: "deadbeef10"}: {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/fnord@deadbeef10>"},
	}
	if diff := cmp.Diff(expectedResults, results); diff != "" {
		t.Errorf("unexpected results (-want +got):\n%s", diff)
	}
}

func TestLocalGitCommand(t *testing.T) {
	// creating a repo with 1 committed file
	root := gitserver.CreateRepoDir(t)

	for _, cmd := range []string{
		"git init",
		"echo -n infile1 > file1",
		"touch --date=2006-01-02T15:04:05Z file1 || touch -t 200601021704.05 file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	} {
		c := exec.Command("bash", "-c", `GIT_CONFIG_GLOBAL="" GIT_CONFIG_SYSTEM="" `+cmd)
		c.Dir = root
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}

	ctx := context.Background()
	command := gitserver.NewLocalGitCommand(api.RepoName(filepath.Base(root)), "log")
	command.ReposDir = filepath.Dir(root)

	stdout, stderr, err := command.DividedOutput(ctx)
	if err != nil {
		t.Fatalf("Local git command run failed. Command: %q Error:\n\n%s", command, err)
	}
	if len(stderr) > 0 {
		t.Fatalf("Local git command run failed. Command: %q Error:\n\n%s", command, stderr)
	}

	stringOutput := string(stdout)
	if !strings.Contains(stringOutput, "commit1") {
		t.Fatalf("No commit message in git log output. Output: %s", stringOutput)
	}
	if command.ExitStatus() != 0 {
		t.Fatalf("Local git command finished with non-zero status. Status: %d", command.ExitStatus())
	}
}
