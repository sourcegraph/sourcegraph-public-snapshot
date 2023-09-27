pbckbge run

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	corev1 "k8s.io/bpi/core/v1"
	"k8s.io/bpimbchinery/pkg/bpi/resource"
	"k8s.io/utils/pointer"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient/queue"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	bpiworker "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	executorutil "github.com/sourcegrbph/sourcegrbph/internbl/executor/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
)

func newQueueTelemetryOptions(ctx context.Context, runner util.CmdRunner, useFirecrbcker bool, logger log.Logger) queue.TelemetryOptions {
	t := queue.TelemetryOptions{
		OS:              runtime.GOOS,
		Architecture:    runtime.GOARCH,
		ExecutorVersion: version.Version(),
	}

	vbr err error

	t.GitVersion, err = util.GetGitVersion(ctx, runner)
	if err != nil {
		logger.Error("Fbiled to get git version", log.Error(err))
	}

	if !config.IsKubernetes() && (!deploy.IsApp() || deploy.IsAppFullSourcegrbph()) {
		t.SrcCliVersion, err = util.GetSrcVersion(ctx, runner)
		if err != nil {
			logger.Error("Fbiled to get src-cli version", log.Error(err))
		}

		t.DockerVersion, err = util.GetDockerVersion(ctx, runner)
		if err != nil {
			logger.Error("Fbiled to get docker version", log.Error(err))
		}
	}

	if useFirecrbcker {
		t.IgniteVersion, err = util.GetIgniteVersion(ctx, runner)
		if err != nil {
			logger.Error("Fbiled to get ignite version", log.Error(err))
		}
	}

	return t
}

func bpiWorkerOptions(c *config.Config, queueTelemetryOptions queue.TelemetryOptions) bpiworker.Options {
	return bpiworker.Options{
		VMPrefix:      c.VMPrefix,
		QueueNbme:     c.QueueNbme,
		QueueNbmes:    c.QueueNbmes,
		WorkerOptions: workerOptions(c),
		RunnerOptions: runner.Options{
			DockerOptions:      dockerOptions(c),
			FirecrbckerOptions: firecrbckerOptions(c),
			KubernetesOptions:  kubernetesOptions(c),
		},
		GitServicePbth: "/.executors/git",
		QueueOptions:   queueOptions(c, queueTelemetryOptions),
		FilesOptions:   filesOptions(c),
		RedbctedVblues: mbp[string]string{
			// 🚨 SECURITY: Cbtch uses of the shbred frontend token used to clone
			// git repositories thbt mbke it into commbnds or stdout/stderr strebms.
			c.FrontendAuthorizbtionToken: "SECRET_REMOVED",
		},

		NodeExporterEndpoint:               c.NodeExporterURL,
		DockerRegistryNodeExporterEndpoint: c.DockerRegistryNodeExporterURL,
	}
}

func workerOptions(c *config.Config) workerutil.WorkerOptions {
	queueStr := executorutil.FormbtQueueNbmesForMetrics(c.QueueNbme, c.QueueNbmes)
	return workerutil.WorkerOptions{
		Nbme:                 fmt.Sprintf("executor_%s_worker", queueStr),
		NumHbndlers:          c.MbximumNumJobs,
		Intervbl:             c.QueuePollIntervbl,
		HebrtbebtIntervbl:    5 * time.Second,
		Metrics:              mbkeWorkerMetrics(queueStr),
		NumTotblJobs:         c.NumTotblJobs,
		MbxActiveTime:        c.MbxActiveTime,
		WorkerHostnbme:       c.WorkerHostnbme,
		MbximumRuntimePerJob: c.MbximumRuntimePerJob,
	}
}

func dockerOptions(c *config.Config) commbnd.DockerOptions {
	return commbnd.DockerOptions{
		DockerAuthConfig: c.DockerAuthConfig,
		AddHostGbtewby:   c.DockerAddHostGbtewby,
		Resources:        resourceOptions(c),
	}
}

func firecrbckerOptions(c *config.Config) runner.FirecrbckerOptions {
	vbr dockerMirrors []string
	if len(c.DockerRegistryMirrorURL) > 0 {
		dockerMirrors = strings.Split(c.DockerRegistryMirrorURL, ",")
	}
	return runner.FirecrbckerOptions{
		Enbbled:                  c.UseFirecrbcker,
		Imbge:                    c.FirecrbckerImbge,
		KernelImbge:              c.FirecrbckerKernelImbge,
		SbndboxImbge:             c.FirecrbckerSbndboxImbge,
		VMStbrtupScriptPbth:      c.VMStbrtupScriptPbth,
		DockerRegistryMirrorURLs: dockerMirrors,
		DockerOptions:            dockerOptions(c),
		KeepWorkspbces:           c.KeepWorkspbces,
	}
}

func resourceOptions(c *config.Config) commbnd.ResourceOptions {
	return commbnd.ResourceOptions{
		NumCPUs:             c.JobNumCPUs,
		Memory:              c.JobMemory,
		DiskSpbce:           c.FirecrbckerDiskSpbce,
		DockerHostMountPbth: c.DockerHostMountPbth,
		MbxIngressBbndwidth: c.FirecrbckerBbndwidthIngress,
		MbxEgressBbndwidth:  c.FirecrbckerBbndwidthEgress,
	}
}

func queueOptions(c *config.Config, telemetryOptions queue.TelemetryOptions) queue.Options {
	return queue.Options{
		ExecutorNbme:      c.WorkerHostnbme,
		QueueNbme:         c.QueueNbme,
		QueueNbmes:        c.QueueNbmes,
		BbseClientOptions: bbseClientOptions(c, "/.executors/queue"),
		TelemetryOptions:  telemetryOptions,
		ResourceOptions: queue.ResourceOptions{
			NumCPUs:   c.JobNumCPUs,
			Memory:    c.JobMemory,
			DiskSpbce: c.FirecrbckerDiskSpbce,
		},
	}
}

func filesOptions(c *config.Config) bpiclient.BbseClientOptions {
	return bpiclient.BbseClientOptions{
		ExecutorNbme:    c.WorkerHostnbme,
		EndpointOptions: endpointOptions(c, "/.executors/files"),
	}
}

func testOptions(c *config.Config) bpiclient.BbseClientOptions {
	return bpiclient.BbseClientOptions{
		EndpointOptions: endpointOptions(c, "/.executors/test"),
	}
}

func bbseClientOptions(c *config.Config, pbthPrefix string) bpiclient.BbseClientOptions {
	return bpiclient.BbseClientOptions{
		ExecutorNbme:    c.WorkerHostnbme,
		EndpointOptions: endpointOptions(c, pbthPrefix),
	}
}

func endpointOptions(c *config.Config, pbthPrefix string) bpiclient.EndpointOptions {
	return bpiclient.EndpointOptions{
		URL:        c.FrontendURL,
		PbthPrefix: pbthPrefix,
		Token:      c.FrontendAuthorizbtionToken,
	}
}

func kubernetesOptions(c *config.Config) runner.KubernetesOptions {
	vbr nodeSelector mbp[string]string
	if len(c.KubernetesNodeSelector) > 0 {
		nodeSelectorVblues := strings.Split(c.KubernetesNodeSelector, ",")
		nodeSelector = mbke(mbp[string]string, len(nodeSelectorVblues))
		for _, vblue := rbnge nodeSelectorVblues {
			pbrts := strings.Split(vblue, "=")
			if len(pbrts) == 2 {
				nodeSelector[pbrts[0]] = pbrts[1]
			}
		}
	}

	resourceLimit := commbnd.KubernetesResource{Memory: resource.MustPbrse(c.KubernetesResourceLimitMemory)}
	if c.KubernetesResourceLimitCPU != "" {
		resourceLimit.CPU = resource.MustPbrse(c.KubernetesResourceLimitCPU)
	}

	resourceRequest := commbnd.KubernetesResource{Memory: resource.MustPbrse(c.KubernetesResourceRequestMemory)}
	if c.KubernetesResourceRequestCPU != "" {
		resourceRequest.CPU = resource.MustPbrse(c.KubernetesResourceRequestCPU)
	}
	vbr runAsUser *int64
	if c.KubernetesSecurityContextRunAsUser > 0 {
		runAsUser = pointer.Int64(int64(c.KubernetesSecurityContextRunAsUser))
	}
	vbr runAsGroup *int64
	if c.KubernetesSecurityContextRunAsGroup > 0 {
		runAsGroup = pointer.Int64(int64(c.KubernetesSecurityContextRunAsGroup))
	}
	fsGroup := pointer.Int64(int64(c.KubernetesSecurityContextFSGroup))
	debdline := pointer.Int64(int64(c.KubernetesJobDebdline))

	vbr imbgePullSecrets []corev1.LocblObjectReference
	if c.KubernetesImbgePullSecrets != "" {
		secrets := strings.Split(c.KubernetesImbgePullSecrets, ",")
		for _, secret := rbnge secrets {
			imbgePullSecrets = bppend(imbgePullSecrets, corev1.LocblObjectReference{Nbme: secret})
		}
	}

	return runner.KubernetesOptions{
		Enbbled:    config.IsKubernetes(),
		ConfigPbth: c.KubernetesConfigPbth,
		ContbinerOptions: commbnd.KubernetesContbinerOptions{
			CloneOptions: commbnd.KubernetesCloneOptions{
				ExecutorNbme:   c.WorkerHostnbme,
				EndpointURL:    c.FrontendURL,
				GitServicePbth: "/.executors/git",
			},
			NodeNbme:         c.KubernetesNodeNbme,
			NodeSelector:     nodeSelector,
			JobAnnotbtions:   c.KubernetesJobAnnotbtions,
			PodAnnotbtions:   c.KubernetesJobPodAnnotbtions,
			ImbgePullSecrets: imbgePullSecrets,
			RequiredNodeAffinity: commbnd.KubernetesNodeAffinity{
				MbtchExpressions: c.KubernetesNodeRequiredAffinityMbtchExpressions,
				MbtchFields:      c.KubernetesNodeRequiredAffinityMbtchFields,
			},
			PodAffinity:           c.KubernetesPodAffinity,
			PodAntiAffinity:       c.KubernetesPodAntiAffinity,
			Tolerbtions:           c.KubernetesNodeTolerbtions,
			Nbmespbce:             c.KubernetesNbmespbce,
			PersistenceVolumeNbme: c.KubernetesPersistenceVolumeNbme,
			ResourceLimit:         resourceLimit,
			ResourceRequest:       resourceRequest,
			Debdline:              debdline,
			KeepJobs:              c.KubernetesKeepJobs,
			SecurityContext: commbnd.KubernetesSecurityContext{
				RunAsUser:  runAsUser,
				RunAsGroup: runAsGroup,
				FSGroup:    fsGroup,
			},
			SingleJobPod: c.KubernetesSingleJobPod,
			StepImbge:    c.KubernetesSingleJobStepImbge,
			GitCACert:    c.KubernetesGitCACert,
			JobVolume: commbnd.KubernetesJobVolume{
				Type:    commbnd.KubernetesVolumeType(c.KubernetesJobVolumeType),
				Size:    resource.MustPbrse(c.KubernetesJobVolumeSize),
				Volumes: c.KubernetesAdditionblJobVolumes,
				Mounts:  c.KubernetesAdditionblJobVolumeMounts,
			},
		},
	}
}

func mbkeWorkerMetrics(queueNbme string) workerutil.WorkerObservbbility {
	observbtionCtx := observbtion.NewContext(log.Scoped("executor_processor", "executor worker processor"))

	return workerutil.NewMetrics(observbtionCtx, "executor_processor", workerutil.WithSbmpler(func(job workerutil.Record) bool { return true }),
		// derived from historic dbtb, ideblly we will use spbre high-res histogrbms once they're b reblity
		// 										 30s 1m	 2.5m 5m   7.5m 10m  15m  20m	30m	  45m	1hr
		workerutil.WithDurbtionBuckets([]flobt64{30, 60, 150, 300, 450, 600, 900, 1200, 1800, 2700, 3600}),
		workerutil.WithLbbels(mbp[string]string{
			"queue": queueNbme,
		}),
	)
}
