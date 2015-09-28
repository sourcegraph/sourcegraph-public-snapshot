package vcs_test

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/git"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/ssh"
)

func init() {
	git.InsecureSkipCheckVerifySSH = true
	gitcmd.InsecureSkipCheckVerifySSH = true
}

func startGitShellSSHServer(t *testing.T, label string, dir string) (*ssh.Server, vcs.RemoteOpts) {
	s, err := ssh.NewServer("git-shell", dir, ssh.PrivateKey(ssh.SamplePrivKey))
	if err != nil {
		t.Fatalf("%s: ssh.NewServer: %s", label, err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("%s: server Start: %s", label, err)
	}
	return s, vcs.RemoteOpts{
		SSH: &vcs.SSHConfig{
			PrivateKey: ssh.SamplePrivKey,
		},
	}
}

func TestRepository_Clone_ssh(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag t0",
		"git checkout -b b0",
	}
	// TODO(sqs): test hg ssh support when it's implemented
	tests := map[string]struct {
		repoDir      string
		cloner       func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error)
		wantCommitID vcs.CommitID // commit ID that tag t0 refers to
	}{
		"git libgit2": {
			repoDir:      initGitRepository(t, gitCommands...),
			cloner:       func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error) { return git.Clone(url, dir, opt) },
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
		"git cmd": {
			repoDir:      initGitRepository(t, gitCommands...),
			cloner:       func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error) { return gitcmd.Clone(url, dir, opt) },
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
	}

	for label, test := range tests {
		func() {
			s, remoteOpts := startGitShellSSHServer(t, label, filepath.Dir(test.repoDir))
			defer s.Close()

			opt := vcs.CloneOpt{
				Bare:       true,
				RemoteOpts: remoteOpts,
			}

			gitURL := s.GitURL + "/" + filepath.Base(test.repoDir)
			cloneDir := makeTmpDir(t, "ssh-clone")
			t.Logf("Cloning from %s to %s", gitURL, cloneDir)
			r, err := test.cloner(gitURL, cloneDir, opt)
			if err != nil {
				t.Fatalf("%s: test.cloner: %s", label, err)
			}

			tags, err := r.Tags()
			if err != nil {
				t.Errorf("%s: Tags: %s", label, err)
			}

			wantTags := []*vcs.Tag{{Name: "t0", CommitID: test.wantCommitID}}
			if !reflect.DeepEqual(tags, wantTags) {
				t.Errorf("%s: got tags %s, want %s", label, asJSON(tags), asJSON(wantTags))
			}

			branches, err := r.Branches(vcs.BranchesOptions{})
			if err != nil {
				t.Errorf("%s: Branches: %s", label, err)
			}
			wantBranches := []*vcs.Branch{
				{Name: "b0", Head: test.wantCommitID},
				{Name: "master", Head: test.wantCommitID},
			}
			if !reflect.DeepEqual(branches, wantBranches) {
				t.Errorf("%s: got branches %s, want %s", label, asJSON(branches), asJSON(wantBranches))
			}
		}()
	}
}

func TestRepository_UpdateEverything_ssh(t *testing.T) {
	t.Parallel()

	// TODO(sqs): this test has a lot of overlap with
	// TestRepository_UpdateEverything.

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	// TODO(sqs): test hg ssh support when it's implemented
	tests := map[string]struct {
		vcs, baseDir, headDir string

		opener func(dir string) (vcs.Repository, error)
		cloner func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error)

		// newCmds should commit a file "newfile" in the repository
		// root and tag the commit with "second". This is used to test
		// that UpdateEverything picks up the new file from the
		// mirror's origin.
		newCmds []string
	}{
		"git": { // git
			"git", initGitRepository(t, gitCommands...), makeTmpDir(t, "git-update-ssh"),
			func(dir string) (vcs.Repository, error) { return git.Open(dir) },
			func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error) { return git.Clone(url, dir, opt) },
			[]string{"git tag t0", "git checkout -b b0"},
		},
		"git cmd": { // gitcmd
			"git", initGitRepository(t, gitCommands...), makeTmpDir(t, "git-update-ssh"),
			func(dir string) (vcs.Repository, error) { return gitcmd.Open(dir) },
			func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error) { return gitcmd.Clone(url, dir, opt) },
			[]string{"git tag t0", "git checkout -b b0"},
		},
	}

	for label, test := range tests {
		func() {
			s, remoteOpts := startGitShellSSHServer(t, label, filepath.Dir(test.baseDir))
			defer s.Close()

			baseURL := s.GitURL + "/" + filepath.Base(test.baseDir)
			t.Logf("Cloning from %s to %s", baseURL, test.headDir)
			_, err := test.cloner(baseURL, test.headDir, vcs.CloneOpt{Bare: true, Mirror: true, RemoteOpts: remoteOpts})
			if err != nil {
				t.Errorf("Clone(%q, %q, %q): %s", label, baseURL, test.headDir, err)
				return
			}

			r, err := test.opener(test.headDir)
			if err != nil {
				t.Errorf("opener[->%s](%q): %s", reflect.TypeOf(test.opener).Out(0), test.headDir, err)
				return
			}

			// r should not have any tags yet.
			tags, err := r.Tags()
			if err != nil {
				t.Errorf("%s: Tags: %s", label, err)
				return
			}
			if len(tags) != 0 {
				t.Errorf("%s: got tags %v, want none", label, tags)
			}

			// run the newCmds to create the new file in the origin repository (NOT
			// the mirror repository; we want to test that UpdateEverything updates the
			// mirror repository).
			for _, cmd := range test.newCmds {
				c := exec.Command("bash", "-c", cmd)
				c.Dir = test.baseDir
				out, err := c.CombinedOutput()
				if err != nil {
					t.Fatalf("%s: exec `%s` failed: %s. Output was:\n\n%s", label, cmd, err, out)
				}
			}

			// update the mirror.
			err = r.(vcs.RemoteUpdater).UpdateEverything(remoteOpts)
			if err != nil {
				t.Errorf("%s: UpdateEverything: %s", label, err)
				return
			}

			// r should now have the tag t0 we added to the base repo,
			// since we just updated r.
			tags, err = r.Tags()
			if err != nil {
				t.Errorf("%s: Tags: %s", label, err)
				return
			}
			if got, want := tagNames(tags), []string{"t0"}; !reflect.DeepEqual(got, want) {
				t.Errorf("%s: got tags %v, want %v", label, got, want)
			}

			// r should now have the branch b0 we added to the base
			// repo, since we just updated r.
			branches, err := r.Branches(vcs.BranchesOptions{})
			if err != nil {
				t.Errorf("%s: Branches: %s", label, err)
				return
			}
			if got, want := branchNames(branches), []string{"b0", "master"}; !reflect.DeepEqual(got, want) {
				t.Errorf("%s: got branches %v, want %v", label, got, want)
			}
		}()
	}
}

func tagNames(tags []*vcs.Tag) []string {
	names := make([]string, len(tags))
	for i, b := range tags {
		names[i] = b.Name
	}
	sort.Strings(names)
	return names
}

func branchNames(branches []*vcs.Branch) []string {
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name
	}
	sort.Strings(names)
	return names
}
