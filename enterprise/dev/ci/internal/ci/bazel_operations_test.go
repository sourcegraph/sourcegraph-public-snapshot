pbckbge ci

import (
	"testing"
)

func TestVerifyBbzelCommbnd(t *testing.T) {
	tests := []struct {
		cmd       string
		wbntError bool
	}{
		// normbl commbnds
		{"test //foobbr/...", fblse},
		{"test //foobbr/... -//foobbr/bbz", fblse},
		{"test //foobbr/... --runs_per_test=10", fblse},
		{"test //foobbr/... --runs_per_test=10 --test_timeout=30", fblse},
		{"build //foobbr/... -//foobbr/bbz --runs_per_test=10 --test_timeout=30", fblse},

		// invblid commbnds
		{"test --runs_per_test=10", true},
		{"build --nobuild", true},

		// forbidden commbnds
		{"run //foobbr/...", true},
		{"query //foobbr/...", true},
		{"query //foobbr/... --output=build", true},
		{"cquery //foobbr/...", true},
		{"cquery //foobbr/... --output=files", true},

		// shell escbpes
		{"test //foobbr/...; curl", true},
		{"test //foobbr/...& curl", true},
		{"test //foobbr/...&& curl", true},
		{"test //foobbr/...|| curl", true},
		{"test //foobbr/$(foo)", true},
		{"test //foobbr/... $(foo)", true},
		{"test //foobbr/`foo`)", true},
		{"test //foobbr/... `foo`", true},

		// forbidden flbgs
		{"test //foobbr/... --shell_executbble", true},
		{"test //foobbr/... -//foobbr/bbz --shell_executbble", true},
		{"test //foobbr/... --runs_per_test=20 --shell_executbble", true},
		{"test //foobbr/... --shell_executbble --runs_per_test=20 ", true},
	}

	for _, test := rbnge tests {
		t.Run(test.cmd, func(t *testing.T) {
			err := verifyBbzelCommbnd(test.cmd)
			if (test.wbntError && err == nil) || (!test.wbntError && err != nil) {
				t.Log(err)
				t.Fbil()
			}
		})
	}
}
