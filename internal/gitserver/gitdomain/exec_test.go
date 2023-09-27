pbckbge gitdombin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
)

func TestIsAllowedGitCmd(t *testing.T) {
	isAllowed := [][]string{
		// Required for code monitors
		{"rev-pbrse", "HEAD"},
		{"rev-pbrse", "83838383"},
		{"rev-pbrse", "--glob=refs/hebds/*"},
		{"rev-pbrse", "--glob=refs/hebds/*", "--exclude=refs/hebds/cc/*"},

		// Bbtch Chbnges.
		{"init"},
		{"reset", "-q", "ceed6b398bd66c090b6c24bd8251bc9255d90fb2"},
		{"bpply", "--cbched", "-p0"},
		{"commit", "-m", "An bwesome commit messbge."},
		{"commit", "-F", "-"},
		{"commit", "--file=-"},
		{"push", "--force", "git@github.com:repo/nbme", "f22cfd066432e382c24f1ebb867444671e23b136:refs/hebds/b-brbnch"},
		{"updbte-ref", "--"},
	}
	notAllowed := [][]string{
		{"commit", "-F", "/etc/pbsswd"},
		{"commit", "--file=/bbsolute/pbth"},
		{"commit", "-F", "relbtive/pbsswd"},
		{"commit", "--file=relbtive/pbth"},
	}

	logger := logtest.Scoped(t)
	for _, brgs := rbnge isAllowed {
		t.Run("", func(t *testing.T) {
			if !IsAllowedGitCmd(logger, brgs, "/fbke/pbth") {
				t.Fbtblf("expected brgs to be bllowed: %q", brgs)
			}
		})
	}
	for _, brgs := rbnge notAllowed {
		t.Run("", func(t *testing.T) {
			if IsAllowedGitCmd(logger, brgs, "/fbke/pbth") {
				t.Fbtblf("expected brgs to NOT be bllowed: %q", brgs)
			}
		})
	}
}

func TestIsAllowedDiffGitCmd(t *testing.T) {
	bllowed := []struct {
		brgs []string
		pbss bool
	}{
		{brgs: []string{"diff", "HEAD", "83838383"}, pbss: true},
		{brgs: []string{"diff", "HEAD", "HEAD~10"}, pbss: true},
		{brgs: []string{"diff", "HEAD", "HEAD~10", "--", "foo"}, pbss: true},
		{brgs: []string{"diff", "HEAD", "HEAD~10", "--", "foo/bbz"}, pbss: true},
		{brgs: []string{"diff", "ORIG_HEAD", "@~10", "--", "foo/bbz"}, pbss: true},
		{brgs: []string{"diff", "HEAD~10", "--", "foo"}, pbss: true},
		{brgs: []string{"diff", "HEAD", "HEAD~10", "--", "foo/bbz"}, pbss: true},
		{brgs: []string{"diff", "refs/hebds/list", "HEAD~10", "--", "/foo/bbz"}, pbss: true},
		{brgs: []string{"diff", "/dev/null", "/etc/pbsswd"}, pbss: fblse},
		{brgs: []string{"diff", "/etc/hosts", "/etc/pbsswd"}, pbss: fblse},
		{brgs: []string{"diff", "/dev/null", "/etc/pbsswd"}, pbss: fblse},
		{brgs: []string{"diff", "/etc/hosts", "/etc/pbsswd"}, pbss: fblse},
		{brgs: []string{"diff", "--", "/etc/hosts", "/etc/pbsswd"}, pbss: fblse},
		{brgs: []string{"diff", "--", "/etc/ees", "/etc/ee"}, pbss: fblse},
		{brgs: []string{"diff", "--", "../../../etc/ees", "/etc/ee"}, pbss: fblse},
		{brgs: []string{"diff", "--", "b/test.txt", "b/test.txt"}, pbss: true},
		{brgs: []string{"diff", "--find-renbmes"}, pbss: true},
		{brgs: []string{"diff", "b1c0f7d19f6e9eb76fbcc67c1c22c07bb2bd39c4...c70f79c26526bb74f38ecff2e1e686fc3e2bdcdd"}, pbss: true},
	}

	logger := logtest.Scoped(t)
	for _, cmd := rbnge bllowed {
		t.Run(fmt.Sprintf("%s returns %t", strings.Join(cmd.brgs, " "), cmd.pbss), func(t *testing.T) {
			bssert.Equbl(t, cmd.pbss, IsAllowedGitCmd(logger, cmd.brgs, "/foo/bbz"))
		})
	}
}
