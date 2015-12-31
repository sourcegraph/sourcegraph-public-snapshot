package sgx

import (
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/go-flags"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Test that we can specify CLI flag names as INI property names. This
// is implemented and tested in our fork of go-flags (see
// https://github.com/jessevdk/go-flags/pull/158), but it's so crucial
// to our UX that we also should test it here.
func TestConfigFile_iniFlagArgs(t *testing.T) {
	var testFlags struct {
		MyFlag string `long:"myflag"`
	}
	_, err := cli.Serve.AddGroup("myGroup", "myGroupLongDesc", &testFlags)
	if err != nil {
		t.Fatal(err)
	}

	config := `
[serve]
http-addr=:1234
myflag=foo
`

	if err := flags.NewIniParser(cli.CLI).Parse(strings.NewReader(config)); err != nil {
		t.Fatal(err)
	}
}
