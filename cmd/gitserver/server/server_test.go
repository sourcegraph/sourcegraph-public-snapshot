package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/mutablelimiter"
)

type Test struct {
	Name             string
	Request          *http.Request
	ExpectedCode     int
	ExpectedBody     string
	ExpectedTrailers http.Header
}

func TestRequest(t *testing.T) {
	tests := []Test{
		{
			Name:         "Command",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/gorilla/mux", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusOK,
			ExpectedBody: "teststdout",
			ExpectedTrailers: http.Header{
				"X-Exec-Error":       {""},
				"X-Exec-Exit-Status": {"42"},
				"X-Exec-Stderr":      {"teststderr"},
			},
		},
		{
			Name:         "CommandWithURL",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "my-mux", "url": "https://github.com/gorilla/mux.git", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusOK,
			ExpectedBody: "teststdout",
			ExpectedTrailers: http.Header{
				"X-Exec-Error":       {""},
				"X-Exec-Exit-Status": {"42"},
				"X-Exec-Stderr":      {"teststderr"},
			},
		},
		{
			Name:         "NonexistingRepo",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/gorilla/doesnotexist", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":false}`,
		},
		{
			Name:         "NonexistingRepoWithURL",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "my-doesnotexist", "url": "https://github.com/gorilla/doesntexist.git", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":false}`,
		},
		{
			Name:         "UnclonedRepo",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/nicksnyder/go-i18n", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":true}`,
		},
		{
			Name:         "UnclonedRepoWithURL",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "my-go-i18n", "url": "https://github.com/nicksnyder/go-i18n.git", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":true}`,
		},
		{
			Name:         "Error",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/gorilla/mux", "args": ["testerror"]}`)),
			ExpectedCode: http.StatusOK,
			ExpectedTrailers: http.Header{
				"X-Exec-Error":       {"testerror"},
				"X-Exec-Exit-Status": {"0"},
				"X-Exec-Stderr":      {""},
			},
		},
		{
			Name:         "EmptyBody",
			Request:      httptest.NewRequest("POST", "/exec", nil),
			ExpectedCode: http.StatusBadRequest,
			ExpectedBody: `EOF`,
		},
		{
			Name:         "EmptyInput",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader("{}")),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":false}`,
		},
	}

	s := &Server{ReposDir: "/testroot", skipCloneForTests: true}
	h := s.Handler()

	repoCloned = func(dir string) bool {
		return dir == "/testroot/github.com/gorilla/mux" || dir == "/testroot/my-mux"
	}

	testRepoExists = func(ctx context.Context, url string) error {
		if url == "https://github.com/nicksnyder/go-i18n.git" {
			return nil
		}
		return errors.New("not cloneable")
	}
	defer func() {
		testRepoExists = nil
	}()

	runCommandMock = func(ctx context.Context, cmd *exec.Cmd) (error, int) {
		switch cmd.Args[1] {
		case "testcommand":
			cmd.Stdout.Write([]byte("teststdout"))
			cmd.Stderr.Write([]byte("teststderr"))
			return nil, 42
		case "testerror":
			return errors.New("testerror"), 0
		}
		return nil, 0
	}
	defer func() { runCommandMock = nil }()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.ResponseRecorder{Body: new(bytes.Buffer)}
			h.ServeHTTP(&w, test.Request)

			res := w.Result()
			if res.StatusCode != test.ExpectedCode {
				t.Errorf("wrong status: expected %d, got %d", test.ExpectedCode, w.Code)
			}

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(body)) != test.ExpectedBody {
				t.Errorf("wrong body: expected %q, got %q", test.ExpectedBody, string(body))
			}

			for k, v := range test.ExpectedTrailers {
				if got := res.Trailer.Get(k); got != v[0] {
					t.Errorf("wrong trailer %q: expected %q, got %q", k, v[0], got)
				}
			}
		})
	}
}

func BenchmarkQuickRevParseHead_packed_refs(b *testing.B) {
	dir, err := ioutil.TempDir("", "gitserver_test")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// This simulates the most amount of work quickRevParseHead has to do, and
	// is also the most common in prod. That is where the final rev is in
	// packed-refs.
	err = ioutil.WriteFile(filepath.Join(dir, "HEAD"), []byte("ref: refs/heads/master\n"), 0600)
	if err != nil {
		b.Fatal(err)
	}
	// in prod the kubernetes repo has a packed-refs file that is 62446 lines
	// long. Simulate something like that with everything except master
	masterRev := "4d5092a09bca95e0153c423d76ef62d4fcd168ec"
	{
		f, err := os.Create(filepath.Join(dir, "packed-refs"))
		if err != nil {
			b.Fatal(err)
		}
		writeRef := func(refBase string, num int) {
			_, err := fmt.Fprintf(f, "%016x%016x%08x %s-%d\n", rand.Uint64(), rand.Uint64(), rand.Uint32(), refBase, num)
			if err != nil {
				b.Fatal(err)
			}
		}
		for i := 0; i < 32; i++ {
			writeRef("refs/heads/feature-branch", i)
		}
		_, err = fmt.Fprintf(f, "%s refs/heads/master\n", masterRev)
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < 10000; i++ {
			// actual format is refs/pull/${i}/head, but doesn't actually
			// matter for testing
			writeRef("refs/pull/head", i)
			writeRef("refs/pull/merge", i)
		}
		err = f.Close()
		if err != nil {
			b.Fatal(err)
		}
	}

	// Exclude setup
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rev, err := quickRevParseHead(dir)
		if err != nil {
			b.Fatal(err)
		}
		if rev != masterRev {
			b.Fatal("unexpected rev: ", rev)
		}
	}

	// Exclude cleanup (defers)
	b.StopTimer()
}

func BenchmarkQuickRevParseHead_unpacked_refs(b *testing.B) {
	dir, err := ioutil.TempDir("", "gitserver_test")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// This simulates the usual case for a repo that HEAD is often
	// updated. The master ref will be unpacked.
	masterRev := "4d5092a09bca95e0153c423d76ef62d4fcd168ec"
	files := map[string]string{
		"HEAD":              "ref: refs/heads/master\n",
		"refs/heads/master": masterRev + "\n",
	}
	for path, content := range files {
		path = filepath.Join(dir, path)
		err := os.MkdirAll(filepath.Dir(path), 0700)
		if err != nil {
			b.Fatal(err)
		}
		err = ioutil.WriteFile(path, []byte(content), 0600)
		if err != nil {
			b.Fatal(err)
		}
	}

	// Exclude setup
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rev, err := quickRevParseHead(dir)
		if err != nil {
			b.Fatal(err)
		}
		if rev != masterRev {
			b.Fatal("unexpected rev: ", rev)
		}
	}

	// Exclude cleanup (defers)
	b.StopTimer()
}

func TestUrlRedactor(t *testing.T) {
	testCases := []struct {
		url      string
		message  string
		redacted string
	}{
		{
			url:      "http://token@github.com/foo/bar/",
			message:  "fatal: repository 'http://token@github.com/foo/bar/' not found",
			redacted: "fatal: repository 'http://<redacted>@github.com/foo/bar/' not found",
		},
		{
			url:      "http://user:password@github.com/foo/bar/",
			message:  "fatal: repository 'http://user:password@github.com/foo/bar/' not found",
			redacted: "fatal: repository 'http://<redacted>:<redacted>@github.com/foo/bar/' not found",
		},
		{
			url:      "http://token@github.com///repo//nick/",
			message:  "fatal: repository 'http://token@github.com/foo/bar/' not found",
			redacted: "fatal: repository 'http://<redacted>@github.com/foo/bar/' not found",
		},
		{
			url:      "token",
			message:  "fatal: repository 'http://token' not found",
			redacted: "fatal: repository 'http://<redacted>' not found",
		},
		{
			url:      "git@github.com:foo/bar",
			message:  "fatal: repository 'git@github.com:foo/bar' not found",
			redacted: "fatal: repository '<redacted>' not found",
		},
	}
	for _, testCase := range testCases {
		if actual := newURLRedactor(testCase.url).redact(testCase.message); actual != testCase.redacted {
			t.Errorf("newUrlRedactor(%q).redact(%q) got %q; want %q", testCase.url, testCase.message, actual, testCase.redacted)
		}
	}
}

func TestCloneRepo(t *testing.T) {
	remote, cleanup1 := tmpDir(t)
	defer cleanup1()

	repo := remote
	cmd := func(name string, arg ...string) string {
		t.Helper()
		c := exec.Command(name, arg...)
		c.Dir = repo
		c.Env = []string{
			"GIT_COMMITTER_NAME=a",
			"GIT_COMMITTER_EMAIL=a@a.com",
			"GIT_AUTHOR_NAME=a",
			"GIT_AUTHOR_EMAIL=a@a.com",
		}
		b, err := c.Output()
		if err != nil {
			t.Fatalf("%s %s failed: %s", name, strings.Join(arg, " "), err)
		}
		return string(b)
	}

	// Setup a repo with a commit so we can see if we can clone it.
	cmd("git", "init", ".")
	cmd("sh", "-c", "echo hello world > hello.txt")
	cmd("git", "add", "hello.txt")
	cmd("git", "commit", "-m", "hello")
	wantCommit := cmd("git", "rev-parse", "HEAD")

	reposDir, cleanup2 := tmpDir(t)
	defer cleanup2()

	s := &Server{
		ReposDir:         reposDir,
		ctx:              context.Background(),
		locker:           &RepositoryLocker{},
		cloneLimiter:     mutablelimiter.New(1),
		cloneableLimiter: mutablelimiter.New(1),
	}
	_, err := s.cloneRepo(context.Background(), "example.com/foo/bar", remote, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Wait until the clone is done. Please do not use this code snippet
	// outside of a test. We only know this works since our test only starts
	// one clone and will have nothing else attempt to lock.
	dst := filepath.Join(s.ReposDir, "example.com/foo/bar")
	for i := 0; i < 1000; i++ {
		_, cloning := s.locker.Status(dst)
		if !cloning {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	repo = dst
	gotCommit := cmd("git", "rev-parse", "HEAD")
	if wantCommit != gotCommit {
		t.Fatal("failed to clone")
	}

	// Test blocking with a failure (already exists since we didn't specify overwrite)
	_, err = s.cloneRepo(context.Background(), "example.com/foo/bar", remote, &cloneOptions{Block: true})
	if !os.IsExist(errors.Cause(err)) {
		t.Fatalf("expected clone repo to fail with already exists: %s", err)
	}

	// Test blocking with overwrite. First add random file to GIT_DIR. If the
	// file is missing after cloning we know the directory was replaced
	mkFiles(t, dst, ".git/HELLO")
	_, err = s.cloneRepo(context.Background(), "example.com/foo/bar", remote, &cloneOptions{Block: true, Overwrite: true})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".git/HELLO")); !os.IsNotExist(err) {
		t.Fatalf("expected clone to be overwritten: %s", err)
	}

	repo = dst
	gotCommit = cmd("git", "rev-parse", "HEAD")
	if wantCommit != gotCommit {
		t.Fatal("failed to clone")
	}
}
