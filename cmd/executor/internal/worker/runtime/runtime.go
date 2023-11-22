package runtime

import (
	"context"

	"github.com/sourcegraph/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/files"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Runtime describe how to run a job in a specific runtime environment.
type Runtime interface {
	// Name returns the name of the runtime.
	Name() Name
	// PrepareWorkspace sets up the workspace for the Job.
	PrepareWorkspace(ctx context.Context, logger cmdlogger.Logger, job types.Job) (workspace.Workspace, error)
	// NewRunner creates a runner that will execute the steps.
	NewRunner(ctx context.Context, logger cmdlogger.Logger, filesStore files.Store, options RunnerOptions) (runner.Runner, error)
	// NewRunnerSpecs builds and returns the commands that the runner will execute.
	NewRunnerSpecs(ws workspace.Workspace, job types.Job) ([]runner.Spec, error)
	//CommandKey() string
}

// RunnerOptions are the options to create a runner.
type RunnerOptions struct {
	Name             string
	Path             string
	DockerAuthConfig types.DockerAuthConfig
}

// New creates the runtime based on the configured environment.
func New(
	logger log.Logger,
	ops *command.Operations,
	filesStore files.Store,
	cloneOpts workspace.CloneOptions,
	runnerOpts runner.Options,
	runner util.CmdRunner,
	cmd command.Command,
) (Runtime, error) {
	// TODO: eventually remove this. It was a quick workaround.
	if util.HasShellBuildTag() {
		logger.Info("runtime 'shell' is supported")
		return &shellRuntime{
			cmd:          cmd,
			operations:   ops,
			filesStore:   filesStore,
			cloneOptions: cloneOpts,
			dockerOpts:   runnerOpts.DockerOptions,
		}, nil
	}

	if runnerOpts.FirecrackerOptions.Enabled {
		// We explicitly want a Firecracker runtime. So validation must pass.
		if err := util.ValidateFirecrackerTools(runner); err != nil {
			var errMissingTools *util.ErrMissingTools
			if errors.As(err, &errMissingTools) {
				logger.Error("runtime 'docker' is not supported: missing required tools", log.Strings("dockerTools", errMissingTools.Tools))
			} else {
				logger.Error("failed to determine if docker tools are configured", log.Error(err))
			}
			return nil, err
		} else if err = util.ValidateIgniteInstalled(context.Background(), runner); err != nil {
			logger.Error("runtime 'firecracker' is not supported: ignite is not installed", log.Error(err))
			return nil, err
		} else if err = util.ValidateCNIInstalled(runner); err != nil {
			logger.Error("runtime 'firecracker' is not supported: CNI plugins are not installed", log.Error(err))
			return nil, err
		} else {
			logger.Info("using runtime 'firecracker'")
			return &firecrackerRuntime{
				cmdRunner:       runner,
				cmd:             cmd,
				operations:      ops,
				filesStore:      filesStore,
				cloneOptions:    cloneOpts,
				firecrackerOpts: runnerOpts.FirecrackerOptions,
			}, nil
		}
	}

	if runnerOpts.KubernetesOptions.Enabled {
		configPath := runnerOpts.KubernetesOptions.ConfigPath
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			kubeConfig, err = rest.InClusterConfig()
			if err != nil {
				return nil, errors.Wrap(err, "failed to create kubernetes client config")
			}
		}
		clientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return nil, err
		}
		kubeCmd := &command.KubernetesCommand{
			Logger:     logger,
			Clientset:  clientset,
			Operations: ops,
		}
		logger.Info("using runtime 'kubernetes'")
		return &kubernetesRuntime{
			cmd:          cmd,
			kubeCmd:      kubeCmd,
			filesStore:   filesStore,
			cloneOptions: cloneOpts,
			operations:   ops,
			options:      runnerOpts.KubernetesOptions.ContainerOptions,
		}, nil
	}

	// Default to Docker runtime.
	if err := util.ValidateDockerTools(runner); err != nil {
		var errMissingTools *util.ErrMissingTools
		if errors.As(err, &errMissingTools) {
			logger.Warn("runtime 'docker' is not supported: missing required tools", log.Strings("dockerTools", errMissingTools.Tools))
		} else {
			logger.Warn("failed to determine if docker tools are configured", log.Error(err))
		}
	} else {
		logger.Info("using runtime 'docker'")
		return &dockerRuntime{
			operations:   ops,
			filesStore:   filesStore,
			cloneOptions: cloneOpts,
			dockerOpts:   runnerOpts.DockerOptions,
			cmd:          cmd,
		}, nil
	}
	return nil, ErrNoRuntime
}

// ErrNoRuntime is the error when there is no runtime configured.
var ErrNoRuntime = errors.New("runtime is not configured")

// Name is the name of the runtime.
type Name string

const (
	NameDocker      Name = "docker"
	NameFirecracker Name = "firecracker"
	NameKubernetes  Name = "kubernetes"
	NameShell       Name = "shell"
)

// CommandKey returns the fully formatted key for the command.
func CommandKey(name Name, rawStepKey string, index int) string {
	switch name {
	case NameKubernetes:
		return kubernetesKey(rawStepKey, index)
	default:
		// shell, docker, and firecracker all use the same key format.
		return dockerKey(rawStepKey, index)
	}
}
