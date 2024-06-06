package gitcli

import (
	"context"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) Config() git.GitConfigBackend {
	return g
}

func (g *gitCLIBackend) Get(ctx context.Context, key string) (string, error) {
	r, err := g.NewCommand(ctx, WithArguments("config", "--get", key))
	if err != nil {
		return "", err
	}

	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		// Exit code 1 means the key is not set.
		var e *commandFailedError
		if errors.As(err, &e) && e.ExitStatus == 1 {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (g *gitCLIBackend) Set(ctx context.Context, key, value string) error {
	r, err := g.NewCommand(ctx, WithArguments("config", key, value))
	if err != nil {
		return err
	}

	defer r.Close()

	// Drain reader so process can exit.
	_, err = io.Copy(io.Discard, r)
	if err != nil {
		return err
	}

	return nil
}

func (g *gitCLIBackend) Unset(ctx context.Context, key string) error {
	r, err := g.NewCommand(ctx, WithArguments("config", "--unset-all", key))
	if err != nil {
		return err
	}

	defer r.Close()

	// Drain reader so process can exit.
	_, err = io.Copy(io.Discard, r)
	if err != nil {
		// Exit code 5 means the key is not set.
		var e *commandFailedError
		if errors.As(err, &e) && e.ExitStatus == 5 {
			return nil
		}

		return err
	}

	return nil
}
