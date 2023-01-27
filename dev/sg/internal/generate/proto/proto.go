package proto

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/buf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func Generate(ctx context.Context, verboseOutput bool) *generate.Report {
	// Save working directory
	cwd, err := os.Getwd()
	if err != nil {
		return &generate.Report{Err: err}
	}
	defer func() {
		os.Chdir(cwd)
	}()

	var (
		start = time.Now()
		sb    strings.Builder
	)

	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return &generate.Report{Err: err}
	}

	output := std.NewOutput(&sb, verboseOutput)
	err = buf.InstallDependencies(ctx, output)
	if err != nil {
		return &generate.Report{Output: sb.String(), Err: err}
	}

	// Run buf generate in every directory with buf.gen.yaml
	var bufGenFilePaths []string

	err = filepath.WalkDir(rootDir, root.SkipGitIgnoreWalkFunc(func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == "buf.gen.yaml" {
			bufGenFilePaths = append(bufGenFilePaths, path)
		}
		return nil
	}))

	if err != nil {
		return &generate.Report{Err: err}
	}

	gobin := filepath.Join(rootDir, ".bin")
	for _, p := range bufGenFilePaths {
		dir := filepath.Dir(p)
		err := os.Chdir(dir)
		if err != nil {
			err = errors.Wrapf(err, "changing directory to %q", dir)
			return &generate.Report{Err: err}
		}

		err = runBufGenerate(ctx, gobin, &sb)
		if err != nil {
			return &generate.Report{Output: sb.String(), Err: err}
		}
	}

	return &generate.Report{
		Output:   sb.String(),
		Duration: time.Since(start),
	}
}

func FindGeneratedFiles(dir string) ([]string, error) {
	var paths []string

	err := filepath.WalkDir(dir, root.SkipGitIgnoreWalkFunc(func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && strings.HasSuffix(path, ".pb.go") {
			paths = append(paths, path)
		}

		return nil
	}))

	return paths, err
}

func runBufGenerate(ctx context.Context, gobin string, w io.Writer) error {
	bufPath := filepath.Join(gobin, "buf")
	err := run.Cmd(ctx, bufPath, "generate").
		Environ(os.Environ()).
		Env(map[string]string{
			"GOBIN": gobin,
		}).
		Run().
		Stream(w)
	if err != nil {
		return errors.Wrap(err, "buf generate returned an error")
	}
	return nil
}
