package coursier

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

var CoursierBinary = "coursier"

var (
	observationContext *observation.Context
	operations         *Operations
)

func init() {
	observationContext = &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	operations = NewOperationsFromMetrics(observationContext)
}

func FetchSources(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) (_ []string, err error) {
	ctx, endObservation := operations.fetchSources.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("dependency", dependency.CoursierSyntax()),
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
			return nil, err
		}
		for _, outputPath := range output {
			for _, srcPath := range []string{
				path.Join(outputPath, "src.zip"),
				path.Join(outputPath, "lib", "src.zip"),
			} {
				stat, err := os.Stat(srcPath)
				if !os.IsNotExist(err) && stat.Mode().IsRegular() {
					return []string{srcPath}, nil
				}
			}
		}
		return nil, errors.Errorf("failed to find src.zip for JVM dependency %s", dependency)
	}
	return runCoursierCommand(
		ctx,
		config,
		// NOTE: make sure to update the method `coursierScript` in
		// vcs_syncer_jvm_packages_test.go if you change the arguments
		// here. The test case assumes that the "--classifier sources"
		// arguments appears at a specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intransitive", dependency.CoursierSyntax(),
		"--classifier", "sources",
	)
}

func FetchByteCode(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) (_ []string, err error) {
	ctx, endObservation := operations.fetchByteCode.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return runCoursierCommand(
		ctx,
		config,
		// NOTE: make sure to update the method `coursierScript` in
		// vcs_syncer_jvm_packages_test.go if you change the arguments
		// here. The test case assumes that the "--classifier sources"
		// arguments appears at a specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intransitive", dependency.CoursierSyntax(),
	)
}

func Exists(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) bool {
	var err error
	ctx, endObservation := operations.exists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("dependency", dependency.CoursierSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if dependency.IsJDK() {
		var sources []string
		sources, err = FetchSources(ctx, config, dependency)
		return err == nil && len(sources) == 1
	}
	_, err = runCoursierCommand(
		ctx,
		config,
		"resolve",
		"--quiet", "--quiet",
		"--intransitive", dependency.CoursierSyntax(),
	)
	return err == nil
}

func runCoursierCommand(ctx context.Context, config *schema.JVMPackagesConnection, args ...string) (_ []string, err error) {
	ctx, traceLog, endObservation := operations.runCommand.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repositories", strings.Join(config.Maven.Repositories, "|")),
		log.String("args", strings.Join(args, ", ")),
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
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "coursier command %q failed with stderr %q and stdout %q", cmd, stderr, &stdout)
	}
	traceLog(log.String("stdout", stdout.String()), log.String("stderr", stderr.String()))

	if stdout.String() == "" {
		return []string{}, nil
	}

	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}
