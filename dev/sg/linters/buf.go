package linters

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/buf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var bufFormat = &linter{
	Name: "Buf Format",
	Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "getting current working directory")
		}
		defer func() {
			os.Chdir(cwd)
		}()

		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrap(err, "getting repository root")
		}

		err = os.Chdir(rootDir)
		if err != nil {
			return errors.Wrap(err, "changing directory to repository root")
		}

		err = buf.InstallDependencies(ctx, out)
		if err != nil {
			return errors.Wrap(err, "installing buf dependencies")
		}

		protoFiles, err := findProtoFiles(rootDir)
		if err != nil {
			return errors.Wrapf(err, "finding .proto files")
		}

		bufArgs := []string{
			"format",
			"--diff",
			"--exit-code",
		}

		for _, file := range protoFiles {
			bufArgs = append(bufArgs, "--path", file)
		}

		gobin := filepath.Join(rootDir, ".bin")
		return runBuf(ctx, gobin, out, bufArgs...)
	},

	Fix: func(ctx context.Context, cio check.IO, args *repo.State) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "getting current working directory")
		}
		defer func() {
			os.Chdir(cwd)
		}()

		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrap(err, "getting repository root")
		}

		err = os.Chdir(rootDir)
		if err != nil {
			return errors.Wrap(err, "changing directory to repository root")
		}

		err = buf.InstallDependencies(ctx, cio.Output)
		if err != nil {
			return errors.Wrap(err, "installing buf dependencies")
		}

		protoFiles, err := findProtoFiles(rootDir)
		if err != nil {
			return errors.Wrapf(err, "finding .proto files")
		}

		bufArgs := []string{
			"format",
			"--write",
		}

		for _, file := range protoFiles {
			bufArgs = append(bufArgs, "--path", file)
		}

		gobin := filepath.Join(rootDir, ".bin")
		return runBuf(ctx, gobin, cio.Output, bufArgs...)
	},
}

func findProtoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, root.SkipGitIgnoreWalkFunc(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".proto" {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return errors.Wrapf(err, "getting relative path for .proto file %q (base %q)", path, dir)
			}

			files = append(files, relPath)
		}

		return nil
	}))

	if err != nil {
		return nil, errors.Wrapf(err, "walking %q to find proto files", dir)
	}

	return files, nil
}

func runBuf(ctx context.Context, gobin string, out *std.Output, parameters ...string) error {
	bufPath := filepath.Join(gobin, "buf")

	arguments := []string{bufPath}
	arguments = append(arguments, parameters...)

	err := root.Run(run.Cmd(ctx, arguments...).
		Environ(os.Environ()).
		Env(map[string]string{
			"GOBIN": gobin,
		})).
		StreamLines(out.Write)
	if err != nil {
		commandStr := fmt.Sprintf("buf %s", strings.Join(parameters, " "))
		return errors.Wrapf(err, "%q returned an error", commandStr)
	}
	return nil
}
