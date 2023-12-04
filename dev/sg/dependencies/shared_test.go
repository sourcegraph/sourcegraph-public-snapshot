package dependencies

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGcloudSourceRegexp(t *testing.T) {
	example1 := "==> Source [/foobar/completion.zsh.inc] in your profile to enable shell command completion for gcloud."
	matches := gcloudSourceRegexp.FindStringSubmatch(example1)
	require.Greater(t, len(matches), 0)
	assert.Equal(t, matches[gcloudSourceRegexp.SubexpIndex("path")], "/foobar/completion.zsh.inc")

	example2 := "==> Source [/foobar/path.zsh.inc] in your profile to add the Google Cloud SDK command line tools to your $PATH."
	matches = gcloudSourceRegexp.FindStringSubmatch(example2)
	require.Greater(t, len(matches), 0)
	assert.Equal(t, matches[gcloudSourceRegexp.SubexpIndex("path")], "/foobar/path.zsh.inc")
}
