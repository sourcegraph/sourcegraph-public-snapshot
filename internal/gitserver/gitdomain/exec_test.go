package gitdomain

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
)

func TestIsAllowedGitCmd(t *testing.T) {
	allowed := [][]string{
		// Required for code monitors
		{"rev-parse", "HEAD"},
		{"rev-parse", "83838383"},
		{"rev-parse", "--glob=refs/heads/*"},
		{"rev-parse", "--glob=refs/heads/*", "--exclude=refs/heads/cc/*"},

		// Batch Changes.
		{"init"},
		{"reset", "-q", "ceed6a398bd66c090b6c24bd8251ac9255d90fb2"},
		{"apply", "--cached", "-p0"},
		{"commit", "-m", "An awesome commit message."},
		{"push", "--force", "git@github.com:repo/name", "f22cfd066432e382c24f1eaa867444671e23a136:refs/heads/a-branch"},
		{"update-ref", "--"},
	}

	logger := logtest.Scoped(t)
	for _, args := range allowed {
		t.Run("", func(t *testing.T) {
			if !IsAllowedGitCmd(logger, args) {
				t.Fatalf("expected args to be allowed: %q", args)
			}
		})
	}
}

func TestIsAllowedDiffGitCmd(t *testing.T) {
	allowed := []struct {
		args []string
		pass bool
	}{
		{args: []string{"diff", "HEAD", "83838383"}, pass: true},
		{args: []string{"diff", "HEAD", "HEAD~10"}, pass: true},
		{args: []string{"diff", "HEAD", "HEAD~10", "--", "foo"}, pass: true},
		{args: []string{"diff", "HEAD", "HEAD~10", "--", "foo/baz"}, pass: true},
		{args: []string{"diff", "ORIG_HEAD", "@~10", "--", "foo/baz"}, pass: true},
		{args: []string{"diff", "HEAD~10", "--", "foo"}, pass: true},
		{args: []string{"diff", "HEAD", "HEAD~10", "--", "foo/baz"}, pass: true},
		{args: []string{"diff", "refs/heads/list", "HEAD~10", "--", "/foo/baz"}, pass: true},
		{args: []string{"diff", "/dev/null", "/etc/passwd"}, pass: false},
		{args: []string{"diff", "/etc/hosts", "/etc/passwd"}, pass: false},
		{args: []string{"diff", "/dev/null", "/etc/passwd"}, pass: false},
		{args: []string{"diff", "/etc/hosts", "/etc/passwd"}, pass: false},
	}

	logger := logtest.Scoped(t)
	for _, cmd := range allowed {
		t.Run(fmt.Sprintf("%s returns %t", strings.Join(cmd.args, " "), cmd.pass), func(t *testing.T) {
			assert.Equal(t, cmd.pass, IsAllowedGitCmd(logger, cmd.args))
		})
	}
}
