package perforce

import (
	"bytes"
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithEnvironment(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	opts := []P4OptionFunc{
		WithEnvironment("KEY1=value1"),
		WithEnvironment("KEY2=value2"),
	}

	c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

	require.Contains(t, c.Env, "KEY1=value1")
	require.Contains(t, c.Env, "KEY2=value2")
}

func TestWithAuthentication(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	t.Run("overrides base environment", func(t *testing.T) {
		fakeUser, fakePass := "fakeUser", "fakePass"
		realUser, realPass := "realUser", "realPass"

		opts := []P4OptionFunc{
			WithAuthentication(realUser, realPass),
			WithEnvironment("P4USER="+fakeUser, "P4PASSWD="+fakePass),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		fakeUserIndex := slices.Index(c.Env, "P4USER="+fakeUser)
		fakePassIndex := slices.Index(c.Env, "P4PASSWD="+fakePass)
		realUserIndex := slices.Index(c.Env, "P4USER="+realUser)
		realPassIndex := slices.Index(c.Env, "P4PASSWD="+realPass)

		// Ensure that all environment variables are set
		require.GreaterOrEqual(t, fakeUserIndex, 0)
		require.GreaterOrEqual(t, fakePassIndex, 0)
		require.GreaterOrEqual(t, realUserIndex, 0)
		require.GreaterOrEqual(t, realPassIndex, 0)

		// Ensure that the override environment variables take precedence over
		// any duplicates in existing environment variables.
		require.Greater(t, realUserIndex, fakeUserIndex)
		require.Greater(t, realPassIndex, fakePassIndex)
	})

	t.Run("last option overrides previous options", func(t *testing.T) {
		oldUser, oldPass := "oldUser", "oldPass"
		newUser, newPass := "newUser", "newPass"

		opts := []P4OptionFunc{
			WithAuthentication(oldUser, oldPass),
			WithAuthentication(newUser, newPass),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		require.NotContains(t, c.Env, "P4USER="+oldUser)
		require.NotContains(t, c.Env, "P4PASSWD="+oldPass)

		require.Contains(t, c.Env, "P4USER="+newUser)
		require.Contains(t, c.Env, "P4PASSWD="+newPass)
	})
}

func TestWithClient(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	t.Run("overrides base environment", func(t *testing.T) {
		fakeClient := "fakeClient"
		realClient := "realClient"

		opts := []P4OptionFunc{
			WithClient(realClient),
			WithEnvironment("P4CLIENT=" + fakeClient),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		fakeClientIndex := slices.Index(c.Env, "P4CLIENT="+fakeClient)
		realClientIndex := slices.Index(c.Env, "P4CLIENT="+realClient)

		// Ensure that all environment variables are set
		require.GreaterOrEqual(t, fakeClientIndex, 0)
		require.GreaterOrEqual(t, realClientIndex, 0)

		// Ensure that the override environment variables take precedence over
		// any duplicates in existing environment variables.
		require.Greater(t, realClientIndex, fakeClientIndex)
	})

	t.Run("last option overrides previous options", func(t *testing.T) {
		oldClient := "oldClient"
		newClient := "newClient"

		opts := []P4OptionFunc{
			WithClient(oldClient),
			WithClient(newClient),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		require.NotContains(t, c.Env, "P4CLIENT="+oldClient)
		require.Contains(t, c.Env, "P4CLIENT="+newClient)
	})
}

func TestWithAlternateHomeDir(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	t.Run("overrides base environment", func(t *testing.T) {
		fakeHomeDir := t.TempDir()
		realHomeDir := t.TempDir()

		opts := []P4OptionFunc{
			WithAlternateHomeDir(realHomeDir),
			WithEnvironment("HOME=" + fakeHomeDir),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		fakeHomeDirIndex := slices.Index(c.Env, "HOME="+fakeHomeDir)
		realHomeDirIndex := slices.Index(c.Env, "HOME="+realHomeDir)

		// Ensure that all environment variables are set
		require.GreaterOrEqual(t, fakeHomeDirIndex, 0)
		require.GreaterOrEqual(t, realHomeDirIndex, 0)

		// Ensure that the override environment variables take precedence over
		// any duplicates in existing environment variables.
		require.Greater(t, realHomeDirIndex, fakeHomeDirIndex)
	})

	t.Run("last option overrides previous options", func(t *testing.T) {
		oldHomeDir := t.TempDir()
		newHomeDir := t.TempDir()

		opts := []P4OptionFunc{
			WithAlternateHomeDir(oldHomeDir),
			WithAlternateHomeDir(newHomeDir),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		require.NotContains(t, c.Env, "HOME="+oldHomeDir)
		require.Contains(t, c.Env, "HOME="+newHomeDir)

	})
}

func TestWithHost(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	t.Run("overrides base environment", func(t *testing.T) {
		fakeHost := "ssl:fakeHost:1666"
		realHost := "ssl:realHost:1666"

		opts := []P4OptionFunc{
			WithHost(realHost),
			WithEnvironment("P4PORT=" + fakeHost),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		fakeHostIndex := slices.Index(c.Env, "P4PORT="+fakeHost)
		realHostIndex := slices.Index(c.Env, "P4PORT="+realHost)

		// Ensure that all environment variables are set
		require.GreaterOrEqual(t, fakeHostIndex, 0)
		require.GreaterOrEqual(t, realHostIndex, 0)
		// Ensure that the override environment variables take precedence over
		// any duplicates in existing environment variables.
		require.Greater(t, realHostIndex, fakeHostIndex)
	})

	t.Run("last option overrides previous options", func(t *testing.T) {
		oldHost := "ssl:oldHost:1666"
		newHost := "ssl:newHost:1666"

		opts := []P4OptionFunc{
			WithHost(oldHost),
			WithHost(newHost),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()
		require.NotContains(t, c.Env, "P4PORT="+oldHost)
		require.Contains(t, c.Env, "P4PORT="+newHost)
	})
}

func TestWithArguments(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	t.Run("basic usage", func(t *testing.T) {
		opts := []P4OptionFunc{
			WithArguments("arg1", "arg2"),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

		require.Equal(t, []string{"arg1", "arg2"}, c.Args[1:])
	})

	t.Run("last option overrides previous options", func(t *testing.T) {
		oldArgs := []string{"oldArg1", "oldArg2"}
		newArgs := []string{"newArg1", "newArg2"}

		opts := []P4OptionFunc{
			WithArguments(oldArgs...),
			WithArguments(newArgs...),
		}

		c := NewBaseCommand(ctx, cwd, opts...).Unwrap()
		require.Equal(t, newArgs, c.Args[1:])
	})
}

func TestWithStdin(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	fakeReader := bytes.NewReader(nil)

	stdin := fakeReader
	opts := []P4OptionFunc{
		WithStdin(fakeReader),
	}

	c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

	require.Equal(t, stdin, c.Stdin)
}

func TestWithStdout(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	fakeWriter := bytes.NewBuffer(nil)

	stdout := fakeWriter
	opts := []P4OptionFunc{
		WithStdout(fakeWriter),
	}

	c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

	require.Equal(t, stdout, c.Stdout)
}

func TestWithStderr(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	fakeWriter := bytes.NewBuffer(nil)

	stderr := fakeWriter
	opts := []P4OptionFunc{
		WithStderr(fakeWriter),
	}

	c := NewBaseCommand(ctx, cwd, opts...).Unwrap()

	require.Equal(t, stderr, c.Stderr)
}

func TestP4CLIENTPath(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()

	opts := []P4OptionFunc{}

	c := NewBaseCommand(ctx, cwd, opts...).Unwrap()
	require.Contains(t, c.Env, "P4CLIENTPATH="+cwd)
}
