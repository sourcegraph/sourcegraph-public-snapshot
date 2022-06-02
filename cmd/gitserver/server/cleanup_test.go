package server

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

const (
	testRepoA = "testrepo-A"
	testRepoC = "testrepo-C"
)

func (s *Server) testSetup(t *testing.T) {
	t.Helper()
	s.Handler() // Handler as a side-effect sets up Server
	db := dbtest.NewDB(t)
	s.DB = database.NewDB(db)
	s.Hostname = "gitserver-0"
}

func TestCleanup_computeStats(t *testing.T) {
	root := t.TempDir()

	for _, name := range []string{"a", "b/d", "c"} {
		p := path.Join(root, name, ".git")
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "--bare", "init", p)
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}

	want := protocol.ReposStats{
		UpdatedAt: time.Now(),

		// This may be different in practice, but the way we setup the tests
		// we only have .git dirs to measure so this is correct.
		GitDirBytes: dirSize(root),
	}

	// We run cleanupRepos because we want to test as a side-effect it creates
	// the correct file in the correct place.
	s := &Server{ReposDir: root,
		Logger: logtest.Scoped(t),
	}
	s.testSetup(t)

	if _, err := s.DB.ExecContext(context.Background(), `
insert into repo(id, name) values (1, 'a'), (2, 'b/d'), (3, 'c');
update gitserver_repos set shard_id = 1;
update gitserver_repos set repo_size_bytes = 228 where repo_id = 3;
`); err != nil {
		t.Fatalf("unexpected error while inserting test data: %s", err)
	}

	s.cleanupRepos([]string{"gitserver-0"})

	for i := 1; i <= 3; i++ {
		repo, err := s.DB.GitserverRepos().GetByID(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if repo.RepoSizeBytes == 0 {
			t.Fatal("repo_size_bytes is not updated")
		}
	}

	// we hardcode the name here so the tests break if someone changes the
	// value of reposStatsName. We don't want it to change without good reason
	// since it will temporarily break the repo-stats endpoint.
	b, err := os.ReadFile(filepath.Join(root, "repos-stats.json"))
	if err != nil {
		t.Fatal(err)
	}

	var got protocol.ReposStats
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	if got.UpdatedAt.Before(want.UpdatedAt) {
		t.Fatal("want should have been computed after we called cleanupRepos")
	}
	if got.UpdatedAt.After(time.Now()) {
		t.Fatal("want.UpdatedAt is in the future")
	}
	got.UpdatedAt = want.UpdatedAt

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("mismatch for (-want +got):\n%s", d)
	}
}

func TestCleanupInactive(t *testing.T) {
	root := t.TempDir()

	repoA := path.Join(root, testRepoA, ".git")
	cmd := exec.Command("git", "--bare", "init", repoA)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoC := path.Join(root, testRepoC, ".git")
	if err := os.MkdirAll(repoC, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	s := &Server{ReposDir: root,
		Logger: logtest.Scoped(t),
	}
	s.testSetup(t)
	s.cleanupRepos([]string{"gitserver-0"})

	if _, err := os.Stat(repoA); os.IsNotExist(err) {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stat(repoC); err == nil {
		t.Error("expected corrupt repoC to be removed during clean up")
	}
}

func TestCleanupWrongShard_forced(t *testing.T) {
	root := t.TempDir()
	// should be allocated to shard gitserver-1
	testRepoD := "testrepo-D"

	repoA := path.Join(root, testRepoA, ".git")
	cmd := exec.Command("git", "--bare", "init", repoA)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoD := path.Join(root, testRepoD, ".git")
	cmdD := exec.Command("git", "--bare", "init", repoD)
	if err := cmdD.Run(); err != nil {
		t.Fatal(err)
	}

	s := &Server{ReposDir: root,
		Logger: logtest.Scoped(t),
	}
	s.testSetup(t)
	// force cleanup of repos
	wrongShardReposDeleteLimit = 10
	s.cleanupRepos([]string{"gitserver-0", "gitserver-1"})

	if _, err := os.Stat(repoA); os.IsNotExist(err) {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stat(repoD); !os.IsNotExist(err) {
		t.Error("expected repoD assigned to different shard to be removed during clean up")
	}
}

func TestCleanupWrongShard_wrongShardName(t *testing.T) {
	root := t.TempDir()
	// should be allocated to shard gitserver-1
	testRepoD := "testrepo-D"

	repoA := path.Join(root, testRepoA, ".git")
	cmd := exec.Command("git", "--bare", "init", repoA)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoD := path.Join(root, testRepoD, ".git")
	cmdD := exec.Command("git", "--bare", "init", repoD)
	if err := cmdD.Run(); err != nil {
		t.Fatal(err)
	}

	s := &Server{ReposDir: root,
		Logger: logtest.Scoped(t),
	}
	s.testSetup(t)
	s.Hostname = "does-not-exist"
	// force cleanup of repos
	wrongShardReposDeleteLimit = 10
	s.cleanupRepos([]string{"gitserver-0", "gitserver-1"})

	if _, err := os.Stat(repoA); err != nil {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stat(repoD); err != nil {
		t.Error("expected repoD assigned to different shard not to be removed")
	}
}

func TestCleanupWrongShard_cleanupDisabled(t *testing.T) {
	root := t.TempDir()
	// should be allocated to shard gitserver-1
	testRepoD := "testrepo-D"

	repoA := path.Join(root, testRepoA, ".git")
	cmd := exec.Command("git", "--bare", "init", repoA)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoD := path.Join(root, testRepoD, ".git")
	cmdD := exec.Command("git", "--bare", "init", repoD)
	if err := cmdD.Run(); err != nil {
		t.Fatal(err)
	}

	s := &Server{ReposDir: root,
		Logger: logtest.Scoped(t),
	}
	s.testSetup(t)
	wrongShardReposDeleteLimit = -1
	s.cleanupRepos([]string{"gitserver-0", "gitserver-1"})

	if _, err := os.Stat(repoA); os.IsNotExist(err) {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stat(repoD); err != nil {
		t.Error("expected repoD assigned to different shard not to be removed", err)
	}
}

// Note that the exact values (e.g. 50 commits) below are related to git's
// internal heuristics regarding whether or not to invoke `git gc --auto`.
//
// They are stable today, but may become flaky in the future if/when the
// relevant internal magic numbers and transformations change.
func TestGitGCAuto(t *testing.T) {
	// Create a test repository with detectable garbage that GC can prune.
	wd := t.TempDir()
	repo := filepath.Join(wd, "garbage-repo")
	runCmd(t, wd, "git", "init", repo)

	// First we need to generate a moderate number of commits.
	for i := 0; i < 50; i++ {
		runCmd(t, repo, "sh", "-c", "echo 1 >> file1")
		runCmd(t, repo, "git", "add", "file1")
		runCmd(t, repo, "git", "commit", "-m", "file1")
	}

	// Now on a second branch, we do the same thing.
	runCmd(t, repo, "git", "checkout", "-b", "secondary")
	for i := 0; i < 50; i++ {
		runCmd(t, repo, "sh", "-c", "echo 2 >> file2")
		runCmd(t, repo, "git", "add", "file2")
		runCmd(t, repo, "git", "commit", "-m", "file2")
	}

	// Bring everything back together in one branch.
	runCmd(t, repo, "git", "checkout", "master")
	runCmd(t, repo, "git", "merge", "secondary")

	// Now create a bare repo like gitserver expects
	root := t.TempDir()
	wdRepo := repo
	repo = filepath.Join(root, "garbage-repo")
	runCmd(t, root, "git", "clone", "--bare", wdRepo, filepath.Join(repo, ".git"))

	// `git count-objects -v` can indicate objects, packs, etc.
	// We'll run this before and after to verify that an action
	// was taken by `git gc --auto`.
	countObjects := func() string {
		t.Helper()
		return runCmd(t, repo, "git", "count-objects", "-v")
	}

	// Verify that we have GC-able objects in the repository.
	if strings.Contains(countObjects(), "count: 0") {
		t.Fatalf("expected git to report objects but none found")
	}

	// Handler must be invoked for Server side-effects.
	s := &Server{ReposDir: root,
		Logger: logtest.Scoped(t),
	}
	s.testSetup(t)
	s.cleanupRepos([]string{"gitserver-0"})

	// Verify that there are no more GC-able objects in the repository.
	if !strings.Contains(countObjects(), "count: 0") {
		t.Fatalf("expected git to report no objects, but found some")
	}
}

func TestCleanupExpired(t *testing.T) {
	root := t.TempDir()

	repoNew := path.Join(root, "repo-new", ".git")
	repoOld := path.Join(root, "repo-old", ".git")
	repoGCNew := path.Join(root, "repo-gc-new", ".git")
	repoGCOld := path.Join(root, "repo-gc-old", ".git")
	repoBoom := path.Join(root, "repo-boom", ".git")
	repoCorrupt := path.Join(root, "repo-corrupt", ".git")
	repoNonBare := path.Join(root, "repo-non-bare", ".git")
	repoPerforce := path.Join(root, "repo-perforce", ".git")
	repoPerforceGCOld := path.Join(root, "repo-perforce-gc-old", ".git")
	repoRemoteURLScrub := path.Join(root, "repo-remote-url-scrub", ".git")
	remote := path.Join(root, "remote", ".git")
	for _, path := range []string{
		repoNew, repoOld,
		repoGCNew, repoGCOld,
		repoBoom, repoCorrupt,
		repoPerforce, repoPerforceGCOld,
		repoRemoteURLScrub,
		remote,
	} {
		cmd := exec.Command("git", "--bare", "init", path)
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}

	if err := exec.Command("git", "init", filepath.Dir(repoNonBare)).Run(); err != nil {
		t.Fatal(err)
	}

	getRemoteURL := func(ctx context.Context, name api.RepoName) (string, error) {
		if name == "repo-boom" {
			return "", errors.Errorf("boom")
		}
		return remote, nil
	}

	modTime := func(path string) time.Time {
		t.Helper()
		fi, err := os.Stat(filepath.Join(path, "HEAD"))
		if err != nil {
			t.Fatal(err)
		}
		return fi.ModTime()
	}
	recloneTime := func(path string) time.Time {
		t.Helper()
		ts, err := getRecloneTime(GitDir(path))
		if err != nil {
			t.Fatal(err)
		}
		return ts
	}

	writeFile(t, filepath.Join(repoGCNew, "gc.log"), []byte("warning: There are too many unreachable loose objects; run 'git prune' to remove them."))
	writeFile(t, filepath.Join(repoGCOld, "gc.log"), []byte("warning: There are too many unreachable loose objects; run 'git prune' to remove them."))

	for path, delta := range map[string]time.Duration{
		repoOld:           2 * repoTTL,
		repoGCOld:         2 * repoTTLGC,
		repoBoom:          2 * repoTTL,
		repoCorrupt:       repoTTLGC / 2, // should only trigger corrupt, not old
		repoPerforce:      2 * repoTTL,
		repoPerforceGCOld: 2 * repoTTLGC,
	} {
		ts := time.Now().Add(-delta)
		if err := setRecloneTime(GitDir(path), ts); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(filepath.Join(path, "HEAD"), ts, ts); err != nil {
			t.Fatal(err)
		}
	}
	if err := gitConfigSet(GitDir(repoCorrupt), gitConfigMaybeCorrupt, "1"); err != nil {
		t.Fatal(err)
	}
	if err := setRepositoryType(GitDir(repoPerforce), "perforce"); err != nil {
		t.Fatal(err)
	}
	if err := setRepositoryType(GitDir(repoPerforceGCOld), "perforce"); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "-C", repoRemoteURLScrub, "remote", "add", "origin", "http://hello:world@boom.com/").Run(); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	repoNewTime := modTime(repoNew)
	repoOldTime := modTime(repoOld)
	repoGCNewTime := modTime(repoGCNew)
	repoGCOldTime := modTime(repoGCOld)
	repoCorruptTime := modTime(repoBoom)
	repoPerforceTime := modTime(repoPerforce)
	repoPerforceGCOldTime := modTime(repoPerforceGCOld)
	repoBoomTime := modTime(repoBoom)
	repoBoomRecloneTime := recloneTime(repoBoom)

	if _, err := os.Stat(repoNonBare); err != nil {
		t.Fatal(err)
	}

	s := &Server{
		Logger:           logtest.Scoped(t),
		ReposDir:         root,
		GetRemoteURLFunc: getRemoteURL,
		GetVCSSyncer: func(ctx context.Context, name api.RepoName) (VCSSyncer, error) {
			return &GitRepoSyncer{}, nil
		},
	}
	s.testSetup(t)
	s.cleanupRepos([]string{"gitserver-0"})

	// repos that shouldn't be re-cloned
	if repoNewTime.Before(modTime(repoNew)) {
		t.Error("expected repoNew to not be modified")
	}
	if repoGCNewTime.Before(modTime(repoGCNew)) {
		t.Error("expected repoGCNew to not be modified")
	}
	if repoPerforceTime.Before(modTime(repoPerforce)) {
		t.Error("expected repoPerforce to not be modified")
	}
	if repoPerforceGCOldTime.Before(modTime(repoPerforceGCOld)) {
		t.Error("expected repoPerforceGCOld to not be modified")
	}

	// repos that should be recloned
	if !repoOldTime.Before(modTime(repoOld)) {
		t.Error("expected repoOld to be recloned during clean up")
	}
	if !repoGCOldTime.Before(modTime(repoGCOld)) {
		t.Error("expected repoGCOld to be recloned during clean up")
	}
	if !repoCorruptTime.Before(modTime(repoCorrupt)) {
		t.Error("expected repoCorrupt to be recloned during clean up")
	}

	// repos that fail to clone need to have recloneTime updated
	if repoBoomTime.Before(modTime(repoBoom)) {
		t.Fatal("expected repoBoom to fail to re-clone due to hardcoding getRemoteURL failure")
	}
	if !repoBoomRecloneTime.Before(recloneTime(repoBoom)) {
		t.Error("expected repoBoom reclone time to be updated")
	}
	if !now.After(recloneTime(repoBoom)) {
		t.Error("expected repoBoom reclone time to be updated to not now")
	}

	// we scrubbed remote URL
	if out, err := exec.Command("git", "-C", repoRemoteURLScrub, "remote", "-v").Output(); len(out) > 0 {
		t.Fatalf("expected no output from git remote after URL scrubbing, got: %s", out)
	} else if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(repoNonBare); err == nil {
		t.Fatal("non-bare repo was not removed")
	}
}

// TestCleanupOldLocks checks whether cleanupRepos removes stale lock files. It
// does not check whether each job in cleanupRepos finishes successfully, nor
// does it check if other files or directories have been created.
func TestCleanupOldLocks(t *testing.T) {
	type file struct {
		name        string
		age         time.Duration
		wantRemoved bool
	}

	cases := []struct {
		name  string
		files []file
	}{
		{
			name: "fresh_config_lock",
			files: []file{
				{
					name: "config.lock",
				},
			},
		},
		{
			name: "stale_config_lock",
			files: []file{
				{
					name:        "config.lock",
					age:         time.Hour,
					wantRemoved: true,
				},
			},
		},
		{
			name: "fresh_packed",
			files: []file{
				{
					name: "packed-refs.lock",
				},
			},
		},
		{
			name: "stale_packed",
			files: []file{
				{
					name:        "packed-refs.lock",
					age:         2 * time.Hour,
					wantRemoved: true,
				},
			},
		},
		{
			name: "fresh_commit-graph_lock",
			files: []file{
				{
					name: "objects/info/commit-graph.lock",
				},
			},
		},
		{
			name: "stale_commit-graph_lock",
			files: []file{
				{
					name:        "objects/info/commit-graph.lock",
					age:         2 * time.Hour,
					wantRemoved: true,
				},
			},
		},
		{
			name: "refs_lock",
			files: []file{
				{
					name: "refs/heads/fresh",
				},
				{
					name: "refs/heads/fresh.lock",
				},
				{
					name: "refs/heads/stale",
				},
				{
					name:        "refs/heads/stale.lock",
					age:         2 * time.Hour,
					wantRemoved: true,
				},
			},
		},
	}

	root := t.TempDir()

	// initialize git directories and place files
	for _, c := range cases {
		cmd := exec.Command("git", "--bare", "init", c.name+"/.git")
		cmd.Dir = root
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
		dir := GitDir(filepath.Join(root, c.name, ".git"))
		for _, f := range c.files {
			writeFile(t, dir.Path(f.name), nil)
			if f.age == 0 {
				continue
			}
			err := os.Chtimes(dir.Path(f.name), time.Now().Add(-f.age), time.Now().Add(-f.age))
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	s := &Server{ReposDir: root, Logger: logtest.Scoped(t)}
	s.testSetup(t)
	s.cleanupRepos([]string{"gitserver-0"})

	isRemoved := func(path string) bool {
		_, err := os.Stat(path)
		return errors.Is(err, fs.ErrNotExist)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := GitDir(filepath.Join(root, c.name, ".git"))
			for _, f := range c.files {
				if f.wantRemoved != isRemoved(dir.Path(f.name)) {
					t.Fatalf("%s should have been removed", f.name)
				}
			}
		})
	}
}

func TestSetupAndClearTmp(t *testing.T) {
	root := t.TempDir()

	s := &Server{ReposDir: root, Logger: logtest.Scoped(t)}

	// All non .git paths should become .git
	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"example.org/repo/.git/HEAD",

		// Needs to be deleted
		".tmp/foo",
		".tmp/baz/bam",

		// Older tmp cleanups that failed
		".tmp-old123/foo",
	)

	tmp, err := s.SetupAndClearTmp()
	if err != nil {
		t.Fatal(err)
	}

	// Straight after cleaning .tmp should be empty
	assertPaths(t, filepath.Join(root, ".tmp"), ".")

	// tmp should exist
	if info, err := os.Stat(tmp); err != nil {
		t.Fatal(err)
	} else if !info.IsDir() {
		t.Fatal("tmpdir is not a dir")
	}

	// tmp should be on the same mount as root, ie root is parent.
	if filepath.Dir(tmp) != root {
		t.Fatalf("tmp is not under root: tmp=%s root=%s", tmp, root)
	}

	// Wait until async cleaning is done
	for i := 0; i < 1000; i++ {
		found := false
		files, err := os.ReadDir(s.ReposDir)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			found = found || strings.HasPrefix(f.Name(), ".tmp-old")
		}
		if !found {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Only files should be the repo files
	assertPaths(t, root,
		"github.com/foo/baz/.git/HEAD",
		"example.org/repo/.git/HEAD",
		".tmp",
	)
}

func TestSetupAndClearTmp_Empty(t *testing.T) {
	root := t.TempDir()

	s := &Server{ReposDir: root, Logger: logtest.Scoped(t)}

	_, err := s.SetupAndClearTmp()
	if err != nil {
		t.Fatal(err)
	}

	// No files, just the empty .tmp dir should exist
	assertPaths(t, root, ".tmp")
}

func TestRemoveRepoDirectory(t *testing.T) {
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"github.com/foo/survivor/.git/HEAD",
		"github.com/bam/bam/.git/HEAD",
		"example.com/repo/.git/HEAD",
	)

	// Set them up in the DB
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := database.NewDB(dbtest.NewDB(t))

	idMapping := make(map[api.RepoName]api.RepoID)

	// Set them all as cloned in the DB
	for _, r := range []string{
		"github.com/foo/baz",
		"github.com/foo/survivor",
		"github.com/bam/bam",
		"example.com/repo",
	} {
		repo := &types.Repo{
			Name: api.RepoName(r),
		}
		if err := db.Repos().Create(ctx, repo); err != nil {
			t.Fatal(err)
		}
		if err := db.GitserverRepos().Upsert(ctx, &types.GitserverRepo{
			RepoID:      repo.ID,
			ShardID:     "test",
			CloneStatus: types.CloneStatusCloned,
		}); err != nil {
			t.Fatal(err)
		}
		idMapping[repo.Name] = repo.ID
	}

	s := &Server{
		Logger:   logtest.Scoped(t),
		ReposDir: root,
		DB:       db,
		ctx:      ctx,
	}

	// Remove everything but github.com/foo/survivor
	for _, d := range []string{
		"github.com/foo/baz/.git",
		"github.com/bam/bam/.git",
		"example.com/repo/.git",
	} {
		if err := s.removeRepoDirectory(GitDir(filepath.Join(root, d))); err != nil {
			t.Fatalf("failed to remove %s: %s", d, err)
		}
	}

	// Removing them a second time is safe
	for _, d := range []string{
		"github.com/foo/baz/.git",
		"github.com/bam/bam/.git",
		"example.com/repo/.git",
	} {
		if err := s.removeRepoDirectory(GitDir(filepath.Join(root, d))); err != nil {
			t.Fatalf("failed to remove %s: %s", d, err)
		}
	}

	assertPaths(t, root,
		"github.com/foo/survivor/.git/HEAD",
		".tmp",
	)

	for _, tc := range []struct {
		name   api.RepoName
		status types.CloneStatus
	}{
		{"github.com/foo/baz", types.CloneStatusNotCloned},
		{"github.com/bam/bam", types.CloneStatusNotCloned},
		{"example.com/repo", types.CloneStatusNotCloned},
		{"github.com/foo/survivor", types.CloneStatusCloned},
	} {
		id, ok := idMapping[tc.name]
		if !ok {
			t.Fatal("id mapping not found")
		}
		r, err := db.GitserverRepos().GetByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if r.CloneStatus != tc.status {
			t.Errorf("Want %q, got %q for %q", tc.status, r.CloneStatus, tc.name)
		}
	}
}

func TestRemoveRepoDirectory_Empty(t *testing.T) {
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
	)
	s := &Server{
		Logger:   logtest.Scoped(t),
		ReposDir: root,
	}

	if err := s.removeRepoDirectory(GitDir(filepath.Join(root, "github.com/foo/baz/.git"))); err != nil {
		t.Fatal(err)
	}

	assertPaths(t, root,
		".tmp",
	)
}

func TestHowManyBytesToFree(t *testing.T) {
	const G = 1024 * 1024 * 1024
	s := &Server{
		Logger:             logtest.Scoped(t),
		DesiredPercentFree: 10,
	}

	tcs := []struct {
		name      string
		diskSize  uint64
		bytesFree uint64
		want      int64
	}{
		{
			name:      "if there is already enough space, no space is freed",
			diskSize:  10 * G,
			bytesFree: 1.5 * G,
			want:      0,
		},
		{
			name:      "if there is exactly enough space, no space is freed",
			diskSize:  10 * G,
			bytesFree: 1 * G,
			want:      0,
		},
		{
			name:      "if there not enough space, some space is freed",
			diskSize:  10 * G,
			bytesFree: 0.5 * G,
			want:      int64(0.5 * G),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			s.DiskSizer = &fakeDiskSizer{
				diskSize:  tc.diskSize,
				bytesFree: tc.bytesFree,
			}
			b, err := s.howManyBytesToFree()
			if err != nil {
				t.Fatal(err)
			}
			if b != tc.want {
				t.Errorf("s.howManyBytesToFree(...) is %v, want 0", b)
			}
		})
	}
}

type fakeDiskSizer struct {
	bytesFree uint64
	diskSize  uint64
}

func (f *fakeDiskSizer) BytesFreeOnDisk(mountPoint string) (uint64, error) {
	return f.bytesFree, nil
}

func (f *fakeDiskSizer) DiskSizeBytes(mountPoint string) (uint64, error) {
	return f.diskSize, nil
}

func mkFiles(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(p)), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, p), nil)
	}
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	err := os.WriteFile(path, content, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

// assertPaths checks that all paths under want exist. It excludes non-empty directories
func assertPaths(t *testing.T, root string, want ...string) {
	t.Helper()
	notfound := make(map[string]struct{})
	for _, p := range want {
		notfound[p] = struct{}{}
	}
	var unwanted []string
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if empty, err := isEmptyDir(path); err != nil {
				t.Fatal(err)
			} else if !empty {
				return nil
			}
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if _, ok := notfound[rel]; ok {
			delete(notfound, rel)
		} else {
			unwanted = append(unwanted, rel)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(notfound) > 0 {
		var paths []string
		for p := range notfound {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		t.Errorf("did not find expected paths: %s", strings.Join(paths, " "))
	}
	if len(unwanted) > 0 {
		sort.Strings(unwanted)
		t.Errorf("found unexpected paths: %s", strings.Join(unwanted, " "))
	}
}

func isEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func TestFreeUpSpace(t *testing.T) {
	t.Run("no error if no space requested and no repos", func(t *testing.T) {
		s := &Server{DiskSizer: &fakeDiskSizer{}, Logger: logtest.Scoped(t)}
		if err := s.freeUpSpace(0); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("error if space requested and no repos", func(t *testing.T) {
		s := &Server{DiskSizer: &fakeDiskSizer{}, Logger: logtest.Scoped(t)}
		if err := s.freeUpSpace(1); err == nil {
			t.Fatal("want error")
		}
	})
	t.Run("oldest repo gets removed to free up space", func(t *testing.T) {
		// Set up.
		rd := t.TempDir()

		r1 := filepath.Join(rd, "repo1")
		r2 := filepath.Join(rd, "repo2")
		if err := makeFakeRepo(r1, 1000); err != nil {
			t.Fatal(err)
		}
		if err := makeFakeRepo(r2, 1000); err != nil {
			t.Fatal(err)
		}
		// Force the modification time of r2 to be after that of r1.
		fi1, err := os.Stat(r1)
		if err != nil {
			t.Fatal(err)
		}
		mtime2 := fi1.ModTime().Add(time.Second)
		if err := os.Chtimes(r2, time.Now(), mtime2); err != nil {
			t.Fatal(err)
		}

		// Run.
		s := Server{
			Logger:    logtest.Scoped(t),
			ReposDir:  rd,
			DiskSizer: &fakeDiskSizer{},
		}
		if err := s.freeUpSpace(1000); err != nil {
			t.Fatal(err)
		}

		// Check.
		assertPaths(t, rd,
			".tmp",
			"repo2/.git/HEAD",
			"repo2/.git/space_eater")
		rds := dirSize(rd)
		wantSize := int64(1000)
		if rds > wantSize {
			t.Errorf("repo dir size is %d, want no more than %d", rds, wantSize)
		}
	})
}

func makeFakeRepo(d string, sizeBytes int) error {
	gd := filepath.Join(d, ".git")
	if err := os.MkdirAll(gd, 0700); err != nil {
		return errors.Wrap(err, "creating .git dir and any parents")
	}
	if err := os.WriteFile(filepath.Join(gd, "HEAD"), nil, 0666); err != nil {
		return errors.Wrap(err, "creating HEAD file")
	}
	if err := os.WriteFile(filepath.Join(gd, "space_eater"), make([]byte, sizeBytes), 0666); err != nil {
		return errors.Wrapf(err, "writing to space_eater file")
	}
	return nil
}

func TestMaybeCorruptStderrRe(t *testing.T) {
	bad := []string{
		"error: packfile .git/objects/pack/pack-a.pack does not match index",
		"error: Could not read d24d09b8bc5d1ea2c3aa24455f4578db6aa3afda\n",
		`error: short SHA1 1325 is ambiguous
error: Could not read d24d09b8bc5d1ea2c3aa24455f4578db6aa3afda`,
		`unrelated
error: Could not read d24d09b8bc5d1ea2c3aa24455f4578db6aa3afda`,
		"\n\nerror: Could not read d24d09b8bc5d1ea2c3aa24455f4578db6aa3afda",
	}
	good := []string{
		"",
		"error: short SHA1 1325 is ambiguous",
		"error: object 156639577dd2ea91cdd53b25352648387d985743 is a blob, not a commit",
		"error: object 45043b3ff0440f4d7937f8c68f8fb2881759edef is a tree, not a commit",
	}
	for _, stderr := range bad {
		if !maybeCorruptStderrRe.MatchString(stderr) {
			t.Errorf("should contain corrupt line:\n%s", stderr)
		}
	}
	for _, stderr := range good {
		if maybeCorruptStderrRe.MatchString(stderr) {
			t.Errorf("should not contain corrupt line:\n%s", stderr)
		}
	}
}

func TestJitterDuration(t *testing.T) {
	f := func(key string) bool {
		d := jitterDuration(key, repoTTLGC/4)
		return 0 <= d && d < repoTTLGC/4
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func prepareEmptyGitRepo(t *testing.T, dir string) GitDir {
	t.Helper()
	cmd := exec.Command("git", "init", ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("execution error: %v, output %s", err, out)
	}
	return GitDir(filepath.Join(dir, ".git"))
}

func TestTooManyLooseObjects(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepareEmptyGitRepo(t, dir)

	// create sentinel object folder
	if err := os.MkdirAll(gitDir.Path("objects", "17"), fs.ModePerm); err != nil {
		t.Fatal(err)
	}

	touch := func(name string) error {
		file, err := os.Create(gitDir.Path("objects", "17", name))
		if err != nil {
			return err
		}
		return file.Close()
	}

	limit := 2 * 256 // 2 objects per folder

	cases := []struct {
		name string
		file string
		want bool
	}{
		{
			name: "empty",
			file: "",
			want: false,
		},
		{
			name: "1 file",
			file: "abc1",
			want: false,
		},
		{
			name: "ignore files with non-hexadecimal names",
			file: "abcxyz123",
			want: false,
		},
		{
			name: "2 files",
			file: "abc2",
			want: false,
		},
		{
			name: "3 files (too many)",
			file: "abc3",
			want: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.file != "" {
				err := touch(c.file)
				if err != nil {
					t.Fatal(err)
				}
			}
			tooManyLO, err := tooManyLooseObjects(gitDir, limit)
			if err != nil {
				t.Fatal(err)
			}
			if tooManyLO != c.want {
				t.Fatalf("want %t, got %t\n", c.want, tooManyLO)
			}
		})
	}
}

func TestTooManyLooseObjectsMissingSentinelDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepareEmptyGitRepo(t, dir)

	_, err := tooManyLooseObjects(gitDir, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHasBitmap(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepareEmptyGitRepo(t, dir)

	t.Run("empty git repo", func(t *testing.T) {
		hasBm, err := hasBitmap(gitDir)
		if err != nil {
			t.Fatal(err)
		}
		if hasBm {
			t.Fatalf("expected no bitmap file for an empty git repository")
		}
	})

	t.Run("repo with bitmap", func(t *testing.T) {
		script := `echo acont > afile
git add afile
git commit -am amsg
git repack -d -l -A --write-bitmap
`
		cmd := exec.Command("/bin/sh", "-euxc", script)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("out=%s, err=%s", out, err)
		}
		hasBm, err := hasBitmap(gitDir)
		if err != nil {
			t.Fatal(err)
		}
		if !hasBm {
			t.Fatalf("expected bitmap file after running git repack -d -l -A --write-bitmap")
		}
	})
}

func TestTooManyPackFiles(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepareEmptyGitRepo(t, dir)

	newPackFile := func(name string) error {
		file, err := os.Create(gitDir.Path("objects", "pack", name))
		if err != nil {
			return err
		}
		return file.Close()
	}

	packLimit := 1

	cases := []struct {
		name string
		file string
		want bool
	}{
		{
			name: "empty",
			file: "",
			want: false,
		},
		{
			name: "1 pack",
			file: "a.pack",
			want: false,
		},
		{
			name: "2 packs",
			file: "b.pack",
			want: true,
		},
		{
			name: "2 packs, with 1 keep file",
			file: "b.keep",
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.file != "" {
				err := newPackFile(c.file)
				if err != nil {
					t.Fatal(err)
				}
			}
			tooManyPf, err := tooManyPackfiles(gitDir, packLimit)
			if err != nil {
				t.Fatal(err)
			}
			if tooManyPf != c.want {
				t.Fatalf("want %t, got %t\n", c.want, tooManyPf)
			}
		})
	}
}

func TestHasCommitGraph(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepareEmptyGitRepo(t, dir)

	t.Run("empty git repo", func(t *testing.T) {
		hasBm, err := hasCommitGraph(gitDir)
		if err != nil {
			t.Fatal(err)
		}
		if hasBm {
			t.Fatalf("expected no commit-graph file for an empty git repository")
		}
	})

	t.Run("commit-graph", func(t *testing.T) {
		script := `echo acont > afile
git add afile
git commit -am amsg
git commit-graph write --reachable --changed-paths
`
		cmd := exec.Command("/bin/sh", "-euxc", script)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("out=%s, err=%s", out, err)
		}
		hasCg, err := hasCommitGraph(gitDir)
		if err != nil {
			t.Fatal(err)
		}
		if !hasCg {
			t.Fatalf("expected commit-graph file after running git commit-graph write --reachable --changed-paths")
		}
	})
}

func TestNeedsMaintenance(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepareEmptyGitRepo(t, dir)

	needed, reason, err := needsMaintenance(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if reason != "bitmap" {
		t.Fatalf("want %s, got %s", "bitmap", reason)
	}
	if !needed {
		t.Fatal("repos without a bitmap should require a repack")
	}

	// create bitmap file and commit-graph
	script := `echo acont > afile
git add afile
git commit -am amsg
git repack -d -l -A --write-bitmap
git commit-graph write --reachable --changed-paths
`
	cmd := exec.Command("/bin/sh", "-euxc", script)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("out=%s, err=%s", out, err)
	}

	needed, reason, err = needsMaintenance(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if reason != "skipped" {
		t.Fatalf("want %s, got %s", "skipped", reason)
	}
	if needed {
		t.Fatal("this repo doesn't need maintenance")
	}
}

func TestPruneIfNeeded(t *testing.T) {
	gitDir := prepareEmptyGitRepo(t, t.TempDir())

	// create sentinel object folder
	if err := os.MkdirAll(gitDir.Path("objects", "17"), fs.ModePerm); err != nil {
		t.Fatal(err)
	}

	limit := -1 // always run prune
	if err := pruneIfNeeded(gitDir, limit); err != nil {
		t.Fatal(err)
	}
}

func TestCleanup_setRepoSizes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	root := t.TempDir()

	for _, name := range []string{
		"ghe.sgdev.org/sourcegraph/gorilla-websocket",
		"ghe.sgdev.org/sourcegraph/gorilla-mux",
		"ghe.sgdev.org/sourcegraph/gorilla-sessions",
	} {
		p := path.Join(root, name, ".git")
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "--bare", "init", p)
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}

	// We run cleanupRepos because we want to test as a side-effect it creates
	// the correct file in the correct place.
	s := &Server{ReposDir: root, Logger: logtest.Scoped(t)}
	s.Handler() // Handler as a side-effect sets up Server
	db := dbtest.NewDB(t)
	s.DB = database.NewDB(db)

	// inserting info about repos to DB. Repo with ID = 1 already has its size
	if _, err := db.Exec(`
insert into repo(id, name)
values (1, 'ghe.sgdev.org/sourcegraph/gorilla-websocket'),
       (2, 'ghe.sgdev.org/sourcegraph/gorilla-mux'),
       (3, 'ghe.sgdev.org/sourcegraph/gorilla-sessions');
update gitserver_repos set shard_id = 1;
update gitserver_repos set repo_size_bytes = 228 where repo_id = 1;
`); err != nil {
		t.Fatalf("unexpected error while inserting test data: %s", err)
	}

	s.cleanupRepos([]string{"gitserver-0"})

	for i := 1; i <= 3; i++ {
		repo, err := s.DB.GitserverRepos().GetByID(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if repo.RepoSizeBytes == 0 {
			t.Fatal("repo_size_bytes is not updated")
		}
		if i == 1 && repo.RepoSizeBytes != 228 {
			t.Fatal("existing repo_size_bytes has been updated")
		}
		if repo.ShardID != "1" {
			t.Fatal("shard_id has been corrupted")
		}
	}
}

func TestSGMLogFile(t *testing.T) {
	logger := logtest.Scoped(t)
	dir := GitDir(t.TempDir())
	cmd := exec.Command("git", "--bare", "init")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	mustHaveLogFile := func(t *testing.T) {
		t.Helper()
		content, err := os.ReadFile(dir.Path(sgmLog))
		if err != nil {
			t.Fatalf("%s should have been set: %s", sgmLog, err)
		}
		if len(content) == 0 {
			t.Fatal("log file should have contained command output")
		}
	}

	// break the repo
	fakeRef := dir.Path("refs", "heads", "apple")
	if _, err := os.Create(fakeRef); err != nil {
		t.Fatal("test setup failed. Could not create fake ref")
	}

	// failed run => log file
	if err := sgMaintenance(logger, dir); err == nil {
		t.Fatal("sgMaintenance should have returned an error")
	}
	mustHaveLogFile(t)

	if got := bestEffortReadFailed(dir); got != 1 {
		t.Fatalf("want 1, got %d", got)
	}

	// fix the repo
	os.Remove(fakeRef)

	// fresh sgmLog file => skip execution
	if err := sgMaintenance(logger, dir); err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	mustHaveLogFile(t)

	// backdate sgmLog file => sgMaintenance ignores log file
	old := time.Now().Add(-2 * sgmLogExpire)
	if err := os.Chtimes(dir.Path(sgmLog), old, old); err != nil {
		t.Fatal(err)
	}
	if err := sgMaintenance(logger, dir); err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if _, err := os.Stat(dir.Path(sgmLog)); err == nil {
		t.Fatalf("%s should have been removed", sgmLog)
	}
}

func TestBestEffortReadFailed(t *testing.T) {
	tc := []struct {
		content     []byte
		wantRetries int
	}{
		{
			content:     nil,
			wantRetries: 0,
		},
		{
			content:     []byte("any content"),
			wantRetries: 0,
		},
		{
			content: []byte(`failed=1

error message`),
			wantRetries: 1,
		},
		{
			content: []byte(`header text
failed=2
error message`),
			wantRetries: 2,
		},
		{
			content: []byte(`failed=

error message`),
			wantRetries: 0,
		},
		{
			content: []byte(`failed=deadbeaf

error message`),
			wantRetries: 0,
		},
		{
			content: []byte(`failed
failed=deadbeaf
failed=1`),
			wantRetries: 0,
		},
		{
			content: []byte(`failed
failed=1
failed=deadbead`),
			wantRetries: 1,
		},
		{
			content: []byte(`failed=
failed=
error message`),
			wantRetries: 0,
		},
		{
			content: []byte(`header failed text

failed=3
failed=4

error message
`),
			wantRetries: 3,
		},
	}

	for _, tt := range tc {
		t.Run(string(tt.content), func(t *testing.T) {
			if got := bestEffortParseFailed(tt.content); got != tt.wantRetries {
				t.Fatalf("want %d, got %d", tt.wantRetries, got)
			}
		})
	}
}
