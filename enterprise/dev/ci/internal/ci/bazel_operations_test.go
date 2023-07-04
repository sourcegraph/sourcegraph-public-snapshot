package ci

import "testing"

func TestVerifyBazelCommand(t *testing.T) {
	tests := []struct {
		cmd       string
		wantError bool
	}{
		{"test //foobar/...", false},
		{"test //foobar/... --runs_per_test=20", false},
		{"test //foobar/...; curl", true},
		{"test //foobar/...& curl", true},
		{"test //foobar/...&& curl", true},
		{"test //foobar/...|| curl", true},
		{"run //foobar/...", true},
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
