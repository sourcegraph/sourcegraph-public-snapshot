package docker

import (
	"strings"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCommand(t testing.TB, line string) instructions.Command {
	n, err := parser.Parse(strings.NewReader(line))
	require.NoError(t, err)
	is, err := instructions.ParseCommand(n.AST.Children[0])
	require.NoError(t, err)
	return is
}

func TestCommandCheckApkAdd(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		for _, pkg := range []string{
			// >= ok
			"package>=1.12",
			// no version ok
			"package",
		} {
			err := commandCheckApkAdd.check(makeCommand(t, "RUN apk add "+pkg))
			assert.Nil(t, err, err)
		}
	})

	t.Run("fail", func(t *testing.T) {
		for _, pkg := range []string{
			// no =
			"package=1.12",
			// no ==
			"package==1.12",
			// no <
			"package<1.12",
			// no <=
			"package<=1.12",
		} {
			err := commandCheckApkAdd.check(makeCommand(t, "RUN apk add "+pkg))
			assert.NotNil(t, err, err)
		}
	})
}
