package gitdomain

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
)

func TestIsAllowedGitCmd(t *testing.T) {
	isAllowed := [][]string{
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
		{"commit", "-F", "-"},
		{"commit", "--file=-"},
		{"push", "--force", "git@github.com:repo/name", "f22cfd066432e382c24f1eaa867444671e23a136:refs/heads/a-branch"},
		{"update-ref", "--"},
	}
	notAllowed := [][]string{
		{"commit", "-F", "/etc/passwd"},
		{"commit", "--file=/absolute/path"},
		{"commit", "-F", "relative/passwd"},
		{"commit", "--file=relative/path"},
	}

	logger := logtest.Scoped(t)
	for _, args := range isAllowed {
		t.Run("", func(t *testing.T) {
			if !IsAllowedGitCmd(logger, args, "/fake/path") {
				t.Fatalf("expected args to be allowed: %q", args)
			}
		})
	}
	for _, args := range notAllowed {
		t.Run("", func(t *testing.T) {
			if IsAllowedGitCmd(logger, args, "/fake/path") {
				t.Fatalf("expected args to NOT be allowed: %q", args)
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
		{args: []string{"diff", "--", "/etc/hosts", "/etc/passwd"}, pass: false},
		{args: []string{"diff", "--", "/etc/ees", "/etc/ee"}, pass: false},
		{args: []string{"diff", "--", "../../../etc/ees", "/etc/ee"}, pass: false},
		{args: []string{"diff", "--", "a/test.txt", "b/test.txt"}, pass: true},
		{args: []string{"diff", "--find-renames"}, pass: true},
		{args: []string{"diff", "a1c0f7d19f6e9eb76facc67c1c22c07bb2ad39c4...c70f79c26526ba74f38ecff2e1e686fc3e2bdcdd"}, pass: true},
	}

	logger := logtest.Scoped(t)
	for _, cmd := range allowed {
		t.Run(fmt.Sprintf("%s returns %t", strings.Join(cmd.args, " "), cmd.pass), func(t *testing.T) {
			assert.Equal(t, cmd.pass, IsAllowedGitCmd(logger, cmd.args, "/foo/baz"))
		})
	}
}
