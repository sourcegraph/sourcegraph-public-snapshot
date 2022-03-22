package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestExecSafe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args                   []string
		wantStdout, wantStderr string
		wantExitCode           int
		wantError              bool
	}{
		{
			args:       []string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?"},
			wantStdout: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8 -\nauthor a\nauthor-date 2006-01-02 15:04:05 +0000\nparents \nsummary foo\n\nfilename ?\n",
		},
		{
			args:       []string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?", "-m", "-i", "-n200", "--author=a@a.com"},
			wantStdout: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8 -\nauthor a\nauthor-date 2006-01-02 15:04:05 +0000\nparents \nsummary foo\n\nfilename ?\n",
		},
		{
			args:       []string{"show"},
			wantStdout: "commit ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8\nAuthor: a <a@a.com>\nDate:   Mon Jan 2 15:04:05 2006 +0000\n\n    foo\n",
		},
		{
			args:         []string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?", ";show"},
			wantStderr:   "fatal: ambiguous argument ';show': unknown revision or path not in the working tree. Use '--' to separate paths from revisions, like this: 'git <command> [<revision>...] -- [<file>...]'",
			wantExitCode: 128,
		},
		{
			args:         []string{"log", "--name-status", "--full-history", "-M", "--date=iso8601", "--format=%H -%nauthor %an%nauthor-date %ai%nparents %P%nsummary %B%nfilename ?;", "show"},
			wantStderr:   "fatal: ambiguous argument 'show': unknown revision or path not in the working tree. Use '--' to separate paths from revisions, like this: 'git <command> [<revision>...] -- [<file>...]'",
			wantExitCode: 128,
		},
		{
			args:      []string{"rm"},
			wantError: true,
		},
		{
			args:      []string{"checkout"},
			wantError: true,
		},
		{
			args:      []string{"show;", "echo", "hello"},
			wantError: true,
		},
	}

	repo := MakeGitRepository(t, "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z")

	for _, test := range tests {
		t.Run(fmt.Sprint(test.args), func(t *testing.T) {
			stdout, stderr, exitCode, err := execSafe(context.Background(), database.NewMockDB(), repo, test.args)
			if err == nil && test.wantError {
				t.Errorf("got error %v, want error %v", err, test.wantError)
			}
			if test.wantError {
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(stdout) != test.wantStdout {
				t.Errorf("got stdout %q, want %q", stdout, test.wantStdout)
			}
			if string(stderr) != test.wantStderr {
				t.Errorf("got stderr %q, want %q", stderr, test.wantStderr)
			}
			if exitCode != test.wantExitCode {
				t.Errorf("got exitCode %d, want %d", exitCode, test.wantExitCode)
			}
		})
	}
}
