package coursier

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var CoursierBinary = "coursier"

var (
	invocTimeout, _ = time.ParseDuration(env.Get("SRC_COURSIER_TIMEOUT", "2m", "Time limit per Coursier invocation, which is used to resolve JVM/Java dependencies."))
	mkdirOnce       sync.Once
)

type CoursierHandle struct {
	operations *operations
	cacheDir   string
}

func NewCoursierHandle(obsctx *observation.Context, cacheDir string) *CoursierHandle {
	mkdirOnce.Do(func() {
		if cacheDir == "" {
			return
		}
		if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("failed to create coursier cache dir in %q: %s\n", cacheDir, err))
		}
	})
	return &CoursierHandle{
		operations: newOperations(obsctx),
		cacheDir:   cacheDir,
	}
}

func (c *CoursierHandle) FetchSources(ctx context.Context, config *schema.JVMPackagesConnection, dependency *reposource.MavenVersionedPackage) (sourceCodeJarPath string, err error) {
	ctx, _, endObservation := c.operations.fetchSources.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("dependency", dependency.VersionedPackageSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if dependency.IsJDK() {
		output, err := c.runCoursierCommand(
			ctx,
			config,
			"java-home", "--jvm",
			dependency.Version,
		)
		if err != nil {
			return "", err
		}
		for _, outputPath := range output {
			for _, srcPath := range []string{
				path.Join(outputPath, "src.zip"),
				path.Join(outputPath, "lib", "src.zip"),
			} {
				stat, err := os.Stat(srcPath)
				if !os.IsNotExist(err) && stat.Mode().IsRegular() {
					return srcPath, nil
				}
			}
		}
		return "", errors.Errorf("failed to find src.zip for JVM dependency %s", dependency)
	}
	paths, err := c.runCoursierCommand(
		ctx,
		config,
		// NOTE: make sure to update the method `coursierScript` in
		// vcs_syncer_jvm_packages_test.go if you change the arguments
		// here. The test case assumes that the "--classifier sources"
		// arguments appears at a specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intransitive", dependency.VersionedPackageSyntax(),
		"--classifier", "sources",
	)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		return "", errors.Errorf("no sources for %s", dependency)
	}
	if len(paths) > 1 {
		return "", errors.Errorf("expected single JAR path but found multiple: %v", paths)
	}
	return paths[0], nil
}

func (c *CoursierHandle) FetchByteCode(ctx context.Context, config *schema.JVMPackagesConnection, dependency *reposource.MavenVersionedPackage) (byteCodeJarPath string, err error) {
	ctx, _, endObservation := c.operations.fetchByteCode.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	paths, err := c.runCoursierCommand(
		ctx,
		config,
		// NOTE: make sure to update the method `coursierScript` in
		// vcs_syncer_jvm_packages_test.go if you change the arguments
		// here. The test case assumes that the "--classifier sources"
		// arguments appears at a specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intransitive", dependency.VersionedPackageSyntax(),
	)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 || (paths[0] == "") {
		return "", errors.Errorf("no bytecode jar for dependency %s", dependency)
	}
	if len(paths) > 1 {
		return "", errors.Errorf("expected single JAR path but found multiple: %v", paths)
	}
	return paths[0], nil
}

func (c *CoursierHandle) Exists(ctx context.Context, config *schema.JVMPackagesConnection, dependency *reposource.MavenVersionedPackage) (err error) {
	ctx, _, endObservation := c.operations.exists.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("dependency", dependency.VersionedPackageSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if dependency.IsJDK() {
		_, err = c.FetchSources(ctx, config, dependency)
	} else {
		_, err = c.runCoursierCommand(
			ctx,
			config,
			"resolve",
			"--quiet", "--quiet",
			"--intransitive", dependency.VersionedPackageSyntax(),
		)
	}
	if err != nil {
		return &coursierError{err}
	}
	return nil
}

type coursierError struct{ error }

func (e coursierError) NotFound() bool {
	return true
}

func (c *CoursierHandle) runCoursierCommand(ctx context.Context, config *schema.JVMPackagesConnection, args ...string) (stdoutLines []string, err error) {
	ctx, cancel := context.WithTimeout(ctx, invocTimeout)
	defer cancel()

	ctx, trace, endObservation := c.operations.runCommand.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.StringSlice("repositories", config.Maven.Repositories),
		attribute.StringSlice("args", args),
	}})
	defer endObservation(1, observation.Args{})

	arguments := args

	if config.Maven.Credentials != "" {
		lines := strings.Split(config.Maven.Credentials, "\n")
		for _, line := range lines {
			arguments = append(arguments, "--credentials", strings.TrimSpace(line))
		}
	}
	cmd := exec.CommandContext(ctx, CoursierBinary, arguments...)

	if len(config.Maven.Repositories) > 0 {
		cmd.Env = append(
			cmd.Env,
			fmt.Sprintf("COURSIER_REPOSITORIES=%v", strings.Join(config.Maven.Repositories, "|")),
		)
	}
	if c.cacheDir != "" {
		cmd.Env = append(cmd.Env, "COURSIER_CACHE="+c.cacheDir)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "coursier command %q failed with stderr %q and stdout %q", cmd, stderr, &stdout)
	}
	trace.AddEvent("TODO Domain Owner", attribute.String("stdout", stdout.String()), attribute.String("stderr", stderr.String()))

	if stdout.String() == "" {
		return []string{}, nil
	}

	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}
