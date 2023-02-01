// Package buf defines shared functionality and utilities for the buf cli.
package buf

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var dependencies = []dependency{
	"github.com/bufbuild/buf/cmd/buf@v1.11.0",
	"github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1",
	"google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1",
}

// InstallDependencies installs the dependencies required to run the buf cli.
func InstallDependencies(ctx context.Context, out *std.Output) error {
	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return errors.Wrap(err, "finding repository root")
	}

	gobin := filepath.Join(rootDir, ".bin")

	for _, d := range dependencies {
		if err := d.install(ctx, gobin, out); err != nil {
			return errors.Wrapf(err, "installing dependency %q", d)
		}
	}

	return nil
}

type dependency string

func (d dependency) String() string { return string(d) }

func (d dependency) install(ctx context.Context, gobin string, out *std.Output) error {
	err := run.Cmd(ctx, "go", "install", d.String()).
		Environ(os.Environ()).
		Env(map[string]string{
			"GOBIN": gobin,
		}).
		Run().
		StreamLines(out.Verbose)
	if err != nil {
		return errors.Wrapf(err, "go install %s returned an error", d)
	}
	return nil
}
