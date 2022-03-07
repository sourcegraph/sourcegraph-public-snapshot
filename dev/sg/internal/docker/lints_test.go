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

func makeStage(t testing.TB, line string) instructions.Stage {
	n, err := parser.Parse(strings.NewReader(line))
	require.NoError(t, err)
	is, _, err := instructions.Parse(n.AST)
	require.NoError(t, err)
	return is[0]
}

func TestStageCheckNoAlpine(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		for _, directive := range []string{
			// allowed base
			"FROM sourcegraph/alpine-3.12:1234@abcde",
			// with exception comment
			"# alpine_base CHECK:ALPINE_OK\nFROM alpine:123@abcde as alpine_base",
		} {
			err := stageCheckNoAlpine.check(makeStage(t, directive))
			assert.Nil(t, err, err)
		}
	})

	t.Run("fail", func(t *testing.T) {
		for _, directive := range []string{
			// disallowed base
			"FROM alpine:123@abcde",
			// malformed exception comment
			"# alpine_base: CHECK:ALPINE_OK\nFROM alpine:123@abcde as alpine_base",
			"# CHECK:ALPINE_OK\nFROM alpine:123@abcde",
		} {
			err := stageCheckNoAlpine.check(makeStage(t, directive))
			assert.NotNil(t, err, err)
		}
	})
}
