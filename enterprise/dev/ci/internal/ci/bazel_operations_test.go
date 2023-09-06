package ci

import (
	"testing"
)

func TestVerifyBazelCommand(t *testing.T) {
	tests := []struct {
		cmd       string
		wantError bool
	}{
		// normal commands
		{"test //foobar/...", false},
		{"test //foobar/... -//foobar/baz", false},
		{"test //foobar/... --runs_per_test=10", false},
		{"test //foobar/... --runs_per_test=10 --test_timeout=30", false},
		{"build //foobar/... -//foobar/baz --runs_per_test=10 --test_timeout=30", false},

		// invalid commands
		{"test --runs_per_test=10", true},
		{"build --nobuild", true},

		// forbidden commands
		{"run //foobar/...", true},
		{"query //foobar/...", true},
		{"query //foobar/... --output=build", true},
		{"cquery //foobar/...", true},
		{"cquery //foobar/... --output=files", true},

		// shell escapes
		{"test //foobar/...; curl", true},
		{"test //foobar/...& curl", true},
		{"test //foobar/...&& curl", true},
		{"test //foobar/...|| curl", true},
		{"test //foobar/$(foo)", true},
		{"test //foobar/... $(foo)", true},
		{"test //foobar/`foo`)", true},
		{"test //foobar/... `foo`", true},

		// forbidden flags
		{"test //foobar/... --shell_executable", true},
		{"test //foobar/... -//foobar/baz --shell_executable", true},
		{"test //foobar/... --runs_per_test=20 --shell_executable", true},
		{"test //foobar/... --shell_executable --runs_per_test=20 ", true},
	}

	for _, test := range tests {
		t.Run(test.cmd, func(t *testing.T) {
			err := verifyBazelCommand(test.cmd)
			if (test.wantError && err == nil) || (!test.wantError && err != nil) {
				t.Log(err)
				t.Fail()
			}
		})
	}
}
