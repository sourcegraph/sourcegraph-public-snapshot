package git_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestRepository_GitCmdRaw(t *testing.T) {
	t.Parallel()

	whiteListedCommands := [][]string{
		[]string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?"},
		[]string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?", "-m", "-i", "-n200", "--author=a@a.com"},
		[]string{"show"},
	}

	repo := makeGitRepository(t, "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z")

	for _, cmd := range whiteListedCommands {
		_, err := git.GitCmdRaw(ctx, repo, cmd)
		if err != nil {
			t.Errorf("GitCmdRaw failed. Got error: %s. For whitelisted cmd: %s\n", err, cmd)
		}
	}
}

func TestRepository_GitCmdRaw_error(t *testing.T) {
	t.Parallel()

	rejectedCommands := [][]string{
		[]string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?", ";show"},
		[]string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?;", "show"},
		[]string{"rm"},
		[]string{"checkout"},
		[]string{"show;", "echo", "hello"},
	}

	repo := makeGitRepository(t, "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z")

	for _, cmd := range rejectedCommands {
		_, err := git.GitCmdRaw(ctx, repo, cmd)
		if err == nil {
			t.Errorf("GitCmdRaw failed. Got output expected error for cmd:\n%s\n", cmd)
		}
	}
}
