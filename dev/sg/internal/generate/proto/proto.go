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

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var dependencies = []dependency{
	"github.com/bufbuild/buf/cmd/buf@v1.11.0",
	"github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1",
	"google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1",
}

func Generate(ctx context.Context) *generate.Report {
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
	gobin := filepath.Join(rootDir, ".bin")

	// Install dependencies into $ROOT/.bin
	for _, d := range dependencies {
		if err := d.install(ctx, gobin, &sb); err != nil {
			return &generate.Report{Err: err}
		}
	}

	// Run buf generate in every directory with buf.gen.yaml
	var bufGenFilePaths []string
	err = filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if filepath.Base(path) == "buf.gen.yaml" {
			bufGenFilePaths = append(bufGenFilePaths, path)
		}
		return nil
	})
	if err != nil {
		return &generate.Report{Err: err}
	}
	for _, p := range bufGenFilePaths {
		dir := filepath.Dir(p)
		os.Chdir(dir)
		if err := runBufGenerate(ctx, gobin, &sb); err != nil {
			return &generate.Report{Err: err}
		}
	}

	return &generate.Report{
		Output:   sb.String(),
		Duration: time.Since(start),
	}
}

func FindGeneratedFiles(dir string) ([]string, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".pb.go") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

type dependency string

func (d dependency) String() string { return string(d) }

func (d dependency) install(ctx context.Context, gobin string, w io.Writer) error {
	err := run.Cmd(ctx, "go", "install", d.String()).
		Environ(os.Environ()).
		Env(map[string]string{
			"GOBIN": gobin,
		}).
		Run().
		Stream(w)
	if err != nil {
		return errors.Wrapf(err, "go install %s returned an error", d)
	}
	return nil
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
