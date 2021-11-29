package npm

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that all the NPM configuration variables are supported by NPM.
// In case NPM removes/renames certain configuration variables on
// changing versions, we should be able to flag that as a test failure.
func TestNPMConfigurationVariablesAreSupported(t *testing.T) {
	npmAvailableConfigVars, err := exec.Command("npm", "config", "ls", "-l").Output()
	assert.Nil(t, err)
	for _, configVar := range npmUsedConfigVars {
		configVarKebabCase := convertScreamingSnakeCaseToKebabCase(strings.TrimPrefix(configVar, "NPM_CONFIG_"))
		assert.True(t, strings.Contains(string(npmAvailableConfigVars), configVarKebabCase))
	}
}

func convertScreamingSnakeCaseToKebabCase(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", "-"))
}

func TestNPMPackOutput(t *testing.T) {
	ref := func(s string) *string { return &s }
	var table = []struct {
		input  string
		expect *string
	}{
		{`[{"filename": "a"}]`, ref("a")},
		{`[{"id": "mypkg", "filename": "a"}]`, ref("a")},
		{`[{"filename": "a"}, {"filename": "b"}]`, nil},
		{`{"filename": "a"}`, nil},
		{`[{"id": "mypkg"}]`, nil},
		{`[{"filename": ["a"]}]`, nil},
		// See [NOTE: npm-tarball-filename-workaround]
		{`[{"filename": "@scope-abc/pkg"}]`, ref("scope-abc-pkg")},
		{`[{"filename": "scope-pkg"}]`, ref("scope-pkg")},
	}
	for _, entry := range table {
		output, err := parseNPMPackOutput(entry.input)
		if entry.expect != nil {
			assert.Nil(t, err)
			assert.Equal(t, output, *entry.expect)
		} else {
			assert.NotNil(t, err)
		}
	}
}
