package linters

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/buf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/proto"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var bufFormat = &linter{
	Name: "Buf Format",
	Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrap(err, "getting repository root")
		}

		err = buf.InstallDependencies(ctx, out)
		if err != nil {
			return errors.Wrap(err, "installing buf dependencies")
		}

		protoFiles, err := buf.ProtoFiles()
		if err != nil {
			return errors.Wrapf(err, "finding .proto files")
		}

		if len(protoFiles) == 0 {
			return errors.New("no .proto files found")
		}

		bufArgs := []string{
			"format",
			"--diff",
			"--exit-code",
		}

		for _, file := range protoFiles {
			f, err := filepath.Rel(rootDir, file)
			if err != nil {
				return errors.Wrapf(err, "getting relative path for file %q (base %q)", file, rootDir)
			}

			bufArgs = append(bufArgs, "--path", f)
		}

		c, err := buf.Cmd(ctx, bufArgs...)
		if err != nil {
			return errors.Wrap(err, "creating buf command")
		}

		err = c.Run().StreamLines(out.Write)
		if err != nil {
			commandString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
			return errors.Wrapf(err, "running %q", commandString)
		}

		return nil

	},

	Fix: func(ctx context.Context, cio check.IO, args *repo.State) error {
		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrap(err, "getting repository root")
		}

		err = buf.InstallDependencies(ctx, cio.Output)
		if err != nil {
			return errors.Wrap(err, "installing buf dependencies")
		}

		protoFiles, err := buf.ProtoFiles()
		if err != nil {
			return errors.Wrapf(err, "finding .proto files")
		}

		bufArgs := []string{
			"format",
			"--write",
		}

		for _, file := range protoFiles {
			f, err := filepath.Rel(rootDir, file)
			if err != nil {
				return errors.Wrapf(err, "getting relative path for file %q (base %q)", file, rootDir)
			}

			bufArgs = append(bufArgs, "--path", f)
		}

		c, err := buf.Cmd(ctx, bufArgs...)
		if err != nil {
			return errors.Wrap(err, "creating buf command")
		}

		err = c.Run().StreamLines(cio.Output.Write)
		if err != nil {
			commandString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
			return errors.Wrapf(err, "running %q", commandString)
		}

		return nil

	},
}

var bufLint = &linter{
	Name: "Buf Lint",
	Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrap(err, "getting repository root")
		}

		err = buf.InstallDependencies(ctx, out)
		if err != nil {
			return errors.Wrap(err, "installing buf dependencies")
		}

		bufModules, err := buf.ModuleFiles()
		if err != nil {
			return errors.Wrapf(err, "finding buf module files")
		}

		if len(bufModules) == 0 {
			return errors.New("no buf modules found")
		}

		for _, file := range bufModules {
			file, err := filepath.Rel(rootDir, file)
			if err != nil {
				return errors.Wrapf(err, "getting relative path for module %q (base %q)", file, rootDir)
			}

			moduleDir := filepath.Dir(file)

			bufArgs := []string{"lint"}

			c, err := buf.Cmd(ctx, bufArgs...)
			if err != nil {
				return errors.Wrap(err, "creating buf command")
			}

			c.Dir(moduleDir)

			err = c.Run().StreamLines(out.Write)
			if err != nil {
				commandString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
				return errors.Wrapf(err, "running %q in %q", commandString, moduleDir)
			}

		}

		return nil
	},
}

var bufGenerate = &linter{
	Name: "Buf Generate",
	Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
		if args.Dirty {
			return errors.New("cannot run 'buf generate' with uncommitted changes")
		}

		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrap(err, "getting repository root")
		}

		err = buf.InstallDependencies(ctx, out)
		if err != nil {
			return errors.Wrap(err, "installing buf dependencies")
		}

		report := proto.Generate(ctx, nil, false)
		if report.Err != nil {
			return report.Err
		}

		generatedFiles, err := buf.CodegenFiles()
		if err != nil {
			return errors.Wrap(err, "finding generated Protobuf files")
		}

		if len(generatedFiles) == 0 {
			return errors.New("no generated files found")
		}

		gitArgs := []string{
			"diff",
			"--exit-code",
			"--color=always",
			"--",
		}

		for _, file := range generatedFiles {
			f, err := filepath.Rel(rootDir, file)
			if err != nil {
				return errors.Wrapf(err, "getting relative path for file %q (base %q)", file, rootDir)
			}

			gitArgs = append(gitArgs, f)
		}

		// Check if there are any changes to the generated files.

		output, err := run.GitCmd(gitArgs...)
		if err != nil && output != "" {
			out.WriteWarningf("Uncommitted changes found after running buf generate:")
			out.Write(strings.TrimSpace(output))
			// Reset repo state
			if _, resetErr := run.GitCmd("reset", "HEAD", "--hard"); resetErr != nil {
				return errors.Wrap(resetErr, "resetting repository state")
			}

			return err
		}

		return nil
	},

	Fix: func(ctx context.Context, cio check.IO, args *repo.State) error {
		report := proto.Generate(ctx, nil, false)
		return report.Err
	},
}
