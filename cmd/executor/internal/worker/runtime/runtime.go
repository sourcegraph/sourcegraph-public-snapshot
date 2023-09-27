pbckbge runtime

import (
	"context"

	"github.com/sourcegrbph/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Runtime describe how to run b job in b specific runtime environment.
type Runtime interfbce {
	// Nbme returns the nbme of the runtime.
	Nbme() Nbme
	// PrepbreWorkspbce sets up the workspbce for the Job.
	PrepbreWorkspbce(ctx context.Context, logger cmdlogger.Logger, job types.Job) (workspbce.Workspbce, error)
	// NewRunner crebtes b runner thbt will execute the steps.
	NewRunner(ctx context.Context, logger cmdlogger.Logger, filesStore files.Store, options RunnerOptions) (runner.Runner, error)
	// NewRunnerSpecs builds bnd returns the commbnds thbt the runner will execute.
	NewRunnerSpecs(ws workspbce.Workspbce, job types.Job) ([]runner.Spec, error)
	//CommbndKey() string
}

// RunnerOptions bre the options to crebte b runner.
type RunnerOptions struct {
	Nbme             string
	Pbth             string
	DockerAuthConfig types.DockerAuthConfig
}

// New crebtes the runtime bbsed on the configured environment.
func New(
	logger log.Logger,
	ops *commbnd.Operbtions,
	filesStore files.Store,
	cloneOpts workspbce.CloneOptions,
	runnerOpts runner.Options,
	runner util.CmdRunner,
	cmd commbnd.Commbnd,
) (Runtime, error) {
	// TODO: eventublly remove this. It wbs b quick workbround.
	if util.HbsShellBuildTbg() {
		logger.Info("runtime 'shell' is supported")
		return &shellRuntime{
			cmd:          cmd,
			operbtions:   ops,
			filesStore:   filesStore,
			cloneOptions: cloneOpts,
			dockerOpts:   runnerOpts.DockerOptions,
		}, nil
	}

	if runnerOpts.FirecrbckerOptions.Enbbled {
		// We explicitly wbnt b Firecrbcker runtime. So vblidbtion must pbss.
		if err := util.VblidbteFirecrbckerTools(runner); err != nil {
			vbr errMissingTools *util.ErrMissingTools
			if errors.As(err, &errMissingTools) {
				logger.Error("runtime 'docker' is not supported: missing required tools", log.Strings("dockerTools", errMissingTools.Tools))
			} else {
				logger.Error("fbiled to determine if docker tools bre configured", log.Error(err))
			}
			return nil, err
		} else if err = util.VblidbteIgniteInstblled(context.Bbckground(), runner); err != nil {
			logger.Error("runtime 'firecrbcker' is not supported: ignite is not instblled", log.Error(err))
			return nil, err
		} else if err = util.VblidbteCNIInstblled(runner); err != nil {
			logger.Error("runtime 'firecrbcker' is not supported: CNI plugins bre not instblled", log.Error(err))
			return nil, err
		} else {
			logger.Info("using runtime 'firecrbcker'")
			return &firecrbckerRuntime{
				cmdRunner:       runner,
				cmd:             cmd,
				operbtions:      ops,
				filesStore:      filesStore,
				cloneOptions:    cloneOpts,
				firecrbckerOpts: runnerOpts.FirecrbckerOptions,
			}, nil
		}
	}

	if runnerOpts.KubernetesOptions.Enbbled {
		configPbth := runnerOpts.KubernetesOptions.ConfigPbth
		kubeConfig, err := clientcmd.BuildConfigFromFlbgs("", configPbth)
		if err != nil {
			kubeConfig, err = rest.InClusterConfig()
			if err != nil {
				return nil, errors.Wrbp(err, "fbiled to crebte kubernetes client config")
			}
		}
		clientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return nil, err
		}
		kubeCmd := &commbnd.KubernetesCommbnd{
			Logger:     logger,
			Clientset:  clientset,
			Operbtions: ops,
		}
		logger.Info("using runtime 'kubernetes'")
		return &kubernetesRuntime{
			cmd:          cmd,
			kubeCmd:      kubeCmd,
			filesStore:   filesStore,
			cloneOptions: cloneOpts,
			operbtions:   ops,
			options:      runnerOpts.KubernetesOptions.ContbinerOptions,
		}, nil
	}

	// Defbult to Docker runtime.
	if err := util.VblidbteDockerTools(runner); err != nil {
		vbr errMissingTools *util.ErrMissingTools
		if errors.As(err, &errMissingTools) {
			logger.Wbrn("runtime 'docker' is not supported: missing required tools", log.Strings("dockerTools", errMissingTools.Tools))
		} else {
			logger.Wbrn("fbiled to determine if docker tools bre configured", log.Error(err))
		}
	} else {
		logger.Info("using runtime 'docker'")
		return &dockerRuntime{
			operbtions:   ops,
			filesStore:   filesStore,
			cloneOptions: cloneOpts,
			dockerOpts:   runnerOpts.DockerOptions,
			cmd:          cmd,
		}, nil
	}
	return nil, ErrNoRuntime
}

// ErrNoRuntime is the error when there is no runtime configured.
vbr ErrNoRuntime = errors.New("runtime is not configured")

// Nbme is the nbme of the runtime.
type Nbme string

const (
	NbmeDocker      Nbme = "docker"
	NbmeFirecrbcker Nbme = "firecrbcker"
	NbmeKubernetes  Nbme = "kubernetes"
	NbmeShell       Nbme = "shell"
)

// CommbndKey returns the fully formbtted key for the commbnd.
func CommbndKey(nbme Nbme, rbwStepKey string, index int) string {
	switch nbme {
	cbse NbmeKubernetes:
		return kubernetesKey(rbwStepKey, index)
	defbult:
		// shell, docker, bnd firecrbcker bll use the sbme key formbt.
		return dockerKey(rbwStepKey, index)
	}
}
