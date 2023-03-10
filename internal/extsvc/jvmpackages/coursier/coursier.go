package coursier

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var CoursierBinary = "coursier"

var (
	coursierCacheDir string
	invocTimeout, _  = time.ParseDuration(env.Get("SRC_COURSIER_TIMEOUT", "2m", "Time limit per Coursier invocation, which is used to resolve JVM/Java dependencies."))
)

func init() {
	// if COURSIER_CACHE_DIR is set, try create that dir and use it. If not set, use the SRC_REPOS_DIR value (or default).
	// This is expected to only be used in gitserver, if this assumption changes, please revisit this due to the failability
	// of this on read-only filesystems.
	coursierCacheDir = env.Get("COURSIER_CACHE_DIR", "", "Directory in which coursier data is cached for JVM package repos.")
	srcReposDir := env.Get("SRC_REPOS_DIR", "/data/repos", "Root dir containing repos.")
	if coursierCacheDir == "" && srcReposDir != "" {
		coursierCacheDir = filepath.Join(srcReposDir, "coursier")
	}
}

func FetchSources(ctx context.Context, config *schema.JVMPackagesConnection, dependency *reposource.MavenVersionedPackage) (sourceCodeJarPath string, err error) {
	operations := getOperations()

	ctx, _, endObservation := operations.fetchSources.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.VersionedPackageSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if dependency.IsJDK() {
		output, err := runCoursierCommand(
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
	paths, err := runCoursierCommand(
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

func FetchByteCode(ctx context.Context, config *schema.JVMPackagesConnection, dependency *reposource.MavenVersionedPackage) (byteCodeJarPath string, err error) {
	operations := getOperations()

	ctx, _, endObservation := operations.fetchByteCode.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	paths, err := runCoursierCommand(
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

func Exists(ctx context.Context, config *schema.JVMPackagesConnection, dependency *reposource.MavenVersionedPackage) (err error) {
	operations := getOperations()

	ctx, _, endObservation := operations.exists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.VersionedPackageSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if dependency.IsJDK() {
		_, err = FetchSources(ctx, config, dependency)
	} else {
		_, err = runCoursierCommand(
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

func runCoursierCommand(ctx context.Context, config *schema.JVMPackagesConnection, args ...string) (stdoutLines []string, err error) {
	operations := getOperations()

	ctx, cancel := context.WithTimeout(ctx, invocTimeout)
	defer cancel()

	ctx, trace, endObservation := operations.runCommand.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("repositories", strings.Join(config.Maven.Repositories, "|")),
		otlog.String("args", strings.Join(args, ", ")),
	}})
	defer endObservation(1, observation.Args{})

	cmd := exec.CommandContext(ctx, CoursierBinary, args...)
	if config.Maven.Credentials != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("COURSIER_CREDENTIALS=%v", config.Maven.Credentials))
	}
	if len(config.Maven.Repositories) > 0 {
		cmd.Env = append(
			cmd.Env,
			fmt.Sprintf("COURSIER_REPOSITORIES=%v", strings.Join(config.Maven.Repositories, "|")),
		)
	}
	if coursierCacheDir != "" {
		// TODO: Don't run this every time we run a coursier command. Hotfix to fix
		// production.
		if err := os.MkdirAll(coursierCacheDir, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "failed to create coursier cache dir in %q", coursierCacheDir)
		}
		cmd.Env = append(cmd.Env, "COURSIER_CACHE="+coursierCacheDir)
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
