package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/src-cli/internal/exec/expect"
)

func Test_CurrentContext(t *testing.T) {
	ctx := context.Background()

	t.Run("docker fails", func(t *testing.T) {
		expect.Commands(t, contextInspectFailure())

		name, err := CurrentContext(ctx)
		assert.Empty(t, name)
		assert.Error(t, err)
	})

	t.Run("docker times out", func(t *testing.T) {
		tctx, cancel := context.WithTimeout(ctx, -1*time.Second)
		t.Cleanup(cancel)

		expect.Commands(t, contextInspectSuccess("desktop-linux"))

		name, err := CurrentContext(tctx)
		assert.Zero(t, name)
		var terr *fastCommandTimeoutError
		assert.ErrorAs(t, err, &terr)
		assert.Equal(t, []string{"context", "inspect", "--format", "{{ .Name }}"}, terr.args)
		assert.Equal(t, fastCommandTimeoutDefault, terr.timeout)
	})

	t.Run("docker succeeds, but returns nothing", func(t *testing.T) {
		expect.Commands(t, contextInspectSuccess(""))

		name, err := CurrentContext(ctx)
		assert.Empty(t, name)
		assert.Error(t, err)
	})

	t.Run("docker succeeds", func(t *testing.T) {
		expect.Commands(t, contextInspectSuccess("desktop-linux"))

		name, err := CurrentContext(ctx)
		assert.Equal(t, "desktop-linux", name)
		assert.NoError(t, err)
	})
}

func contextInspectSuccess(ncpu string) *expect.Expectation {
	return expect.NewLiteral(
		expect.Behaviour{Stdout: []byte(fmt.Sprintf("%s\n", ncpu))},
		"docker", "context", "inspect", "--format", "{{ .Name }}",
	)
}

func contextInspectFailure() *expect.Expectation {
	return expect.NewLiteral(
		expect.Behaviour{ExitCode: 1},
		"docker", "context", "inspect", "--format", "{{ .Name }}",
	)
}

func Test_NCPU(t *testing.T) {
	ctx := context.Background()

	t.Run("docker fails", func(t *testing.T) {
		expect.Commands(t, infoFailure())

		ncpu, err := NCPU(ctx)
		assert.Zero(t, ncpu)
		assert.Error(t, err)
	})

	t.Run("docker times out", func(t *testing.T) {
		tctx, cancel := context.WithTimeout(ctx, -1*time.Second)
		t.Cleanup(cancel)

		expect.Commands(t, infoSuccess("4"))

		ncpu, err := NCPU(tctx)
		assert.Zero(t, ncpu)
		var terr *fastCommandTimeoutError
		assert.ErrorAs(t, err, &terr)
		assert.Equal(t, []string{"info", "--format", "{{ json .}}"}, terr.args)
		assert.Equal(t, fastCommandTimeoutDefault, terr.timeout)
	})

	t.Run("docker succeeds, but returns nothing", func(t *testing.T) {
		expect.Commands(t, infoSuccess(""))

		ncpu, err := NCPU(ctx)
		assert.Zero(t, ncpu)
		assert.Error(t, err)
	})

	t.Run("docker succeeds, but returns something invalid", func(t *testing.T) {
		expect.Commands(t, infoSuccess("foo"))

		ncpu, err := NCPU(ctx)
		assert.Zero(t, ncpu)
		assert.Error(t, err)
	})

	t.Run("docker succeeds", func(t *testing.T) {
		expect.Commands(t, infoSuccess("4"))

		ncpu, err := NCPU(ctx)
		assert.Equal(t, 4, ncpu)
		assert.NoError(t, err)
	})
}

func infoSuccess(ncpu string) *expect.Expectation {
	var b []byte
	n, err := strconv.Atoi(ncpu)
	if err == nil {
		b, _ = json.Marshal(Info{NCPU: n})
	} else {
		b = []byte(ncpu)
	}
	return expect.NewLiteral(
		expect.Behaviour{Stdout: b},
		"docker", "info", "--format", "{{ json .}}",
	)
}

func infoFailure() *expect.Expectation {
	return expect.NewLiteral(
		expect.Behaviour{ExitCode: 1},
		"docker", "info", "--format", "{{ json .}}",
	)
}
