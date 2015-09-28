package makex

import (
	"bytes"
	"os"
	"testing"
)

func TestExternal(t *testing.T) {
	tests := []struct {
		makefile   string
		extraArgs  []string
		wantOutput string
	}{
		{
			`
foo123:
	echo hello
`,
			nil, "echo hello",
		},
	}
	for i, test := range tests {
		args := append(test.extraArgs, "-n")
		out, err := External(os.TempDir(), []byte(test.makefile), args)
		if err != nil {
			t.Errorf("#%d: ExternalMake: %s", i, err)
		}
		if out := string(bytes.TrimSpace(out)); out != test.wantOutput {
			t.Errorf("#%d: bad output\n\n====== got output\n%s\n\n====== want output\n%s", i, out, test.wantOutput)
		}
	}
}
