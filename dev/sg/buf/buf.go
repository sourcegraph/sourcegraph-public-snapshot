// Package buf defines shared functionality and utilities for interacting with the buf cli.
package buf

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var dependencies = []dependency{
	"github.com/bufbuild/buf/cmd/buf@v1.11.0",
	"github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1",
	"google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1",
}

type dependency string

func (d dependency) String() string { return string(d) }

// InstallDependencies installs the dependencies required to run the buf cli.
func InstallDependencies(ctx context.Context, output output.Writer) error {
	rootDir, err := root.RepositoryRoot()

	if err != nil {
		return errors.Wrap(err, "finding repository root")
	}

	gobin := filepath.Join(rootDir, ".bin")
	for _, d := range dependencies {
		err := run.Cmd(ctx, "go", "install", d.String()).
			Environ(os.Environ()).
			Env(map[string]string{
				"GOBIN": gobin,
			}).
			Run().StreamLines(output.Verbose)

		if err != nil {
			commandString := fmt.Sprintf("go install %s", d)
			return errors.Wrapf(err, "running %q", commandString)
		}
	}

	return nil
}

// ProtoFiles lists the absolute path of all Protobuf files contained in the repository.
func ProtoFiles() ([]string, error) {
	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrap(err, "finding repository root")
	}

	var files []string
	err = filepath.WalkDir(rootDir, root.SkipGitIgnoreWalkFunc(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".proto" {
			files = append(files, path)
		}

		return nil
	}))

	if err != nil {
		return nil, errors.Wrapf(err, "walking %q", rootDir)
	}

	return files, err
}

// ModuleFiles lists the absolute path of all Buf Module files contained in the repository.
func ModuleFiles() ([]string, error) {
	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrap(err, "finding repository root")
	}

	var files []string
	err = filepath.WalkDir(rootDir, root.SkipGitIgnoreWalkFunc(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Base(path) == "buf.yaml" {
			files = append(files, path)
		}

		return nil
	}))

	if err != nil {
		return nil, errors.Wrapf(err, "walking %q", rootDir)
	}

	return files, err
}

// PluginConfigurationFiles lists the absolute path of all Buf plugin template configuration files (buf.gen.yaml) contained in the repository.
func PluginConfigurationFiles() ([]string, error) {
	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrap(err, "finding repository root")
	}

	var files []string
	err = filepath.WalkDir(rootDir, root.SkipGitIgnoreWalkFunc(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Base(path) == "buf.gen.yaml" {
			files = append(files, path)
		}

		return nil
	}))

	if err != nil {
		return nil, errors.Wrapf(err, "walking %q", rootDir)
	}

	return files, err
}

// CodegenFiles lists the absolute path of all the Go-generated GRPC files (*.pb.go) contained in the repository.
func CodegenFiles() ([]string, error) {
	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrap(err, "finding repository root")
	}

	var files []string
	err = filepath.WalkDir(rootDir, root.SkipGitIgnoreWalkFunc(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".pb.go") {
			files = append(files, path)
		}

		return nil
	}))

	if err != nil {
		return nil, errors.Wrapf(err, "walking %q", rootDir)
	}

	return files, err
}

// Cmd returns a run.Command that will execute the buf cli with the given parameters
// from the repository root.
func Cmd(ctx context.Context, parameters ...string) (*run.Command, error) {
	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrap(err, "finding repository root")
	}

	bufPath := filepath.Join(rootDir, ".bin", "buf")
	arguments := []string{bufPath}
	arguments = append(arguments, parameters...)

	c := run.Cmd(ctx, arguments...).
		Dir(rootDir).
		Environ(os.Environ())

	return c, nil
}
