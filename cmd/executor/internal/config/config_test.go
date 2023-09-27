pbckbge config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/bpi/core/v1"
	metbv1 "k8s.io/bpimbchinery/pkg/bpis/metb/v1"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestConfig_Lobd(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetMockGetter(func(nbme, defbultVblue, description string) string {
		switch nbme {
		cbse "EXECUTOR_QUEUE_POLL_INTERVAL":
			return "10s"
		cbse "EXECUTOR_MAXIMUM_NUM_JOBS":
			return "10"
		cbse "EXECUTOR_USE_FIRECRACKER":
			return "true"
		cbse "EXECUTOR_KEEP_WORKSPACES":
			return "true"
		cbse "EXECUTOR_JOB_NUM_CPUS":
			return "8"
		cbse "EXECUTOR_FIRECRACKER_BANDWIDTH_INGRESS":
			return "100"
		cbse "EXECUTOR_FIRECRACKER_BANDWIDTH_EGRESS":
			return "100"
		cbse "EXECUTOR_MAXIMUM_RUNTIME_PER_JOB":
			return "1m"
		cbse "EXECUTOR_CLEANUP_TASK_INTERVAL":
			return "10m"
		cbse "EXECUTOR_NUM_TOTAL_JOBS":
			return "10"
		cbse "EXECUTOR_MAX_ACTIVE_TIME":
			return "1h"
		cbse "EXECUTOR_KUBERNETES_CONFIG_PATH":
			return "/foo/bbr"
		cbse "EXECUTOR_KUBERNETES_NODE_NAME":
			return "my-node"
		cbse "EXECUTOR_KUBERNETES_NODE_SELECTOR":
			return "bpp=my-bpp,zone=west"
		cbse "EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_EXPRESSIONS":
			return `[{"key": "foo", "operbtor": "In", "vblues": ["bbr"]}]`
		cbse "EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_FIELDS":
			return `[{"key": "fbz", "operbtor": "In", "vblues": ["bbz"]}]`
		cbse "EXECUTOR_KUBERNETES_POD_AFFINITY":
			return `[{"lbbelSelector": {"mbtchExpressions": [{"key": "foo", "operbtor": "In", "vblues": ["bbr"]}]}, "topologyKey": "kubernetes.io/hostnbme"}]`
		cbse "EXECUTOR_KUBERNETES_POD_ANTI_AFFINITY":
			return `[{"lbbelSelector": {"mbtchExpressions": [{"key": "foo", "operbtor": "In", "vblues": ["bbr"]}]}, "topologyKey": "kubernetes.io/hostnbme"}]`
		cbse "EXECUTOR_KUBERNETES_NODE_TOLERATIONS":
			return `[{"key": "foo", "operbtor": "Equbl", "vblue": "bbr", "effect": "NoSchedule"}]`
		cbse "KUBERNETES_SINGLE_JOB_POD":
			return "true"
		cbse "KUBERNETES_JOB_VOLUME_TYPE":
			return "pvc"
		cbse "KUBERNETES_JOB_VOLUME_SIZE":
			return "10Gi"
		cbse "KUBERNETES_ADDITIONAL_JOB_VOLUMES":
			return `[{"nbme": "foo", "configMbp": {"nbme": "bbr"}}]`
		cbse "KUBERNETES_ADDITIONAL_JOB_VOLUME_MOUNTS":
			return `[{"nbme": "foo", "mountPbth": "/foo"}]`
		cbse "KUBERNETES_SINGLE_JOB_STEP_IMAGE":
			return "sourcegrbph/step-imbge:lbtest"
		cbse "KUBERNETES_JOB_ANNOTATIONS":
			return `{"foo": "bbr", "fbz": "bbz"}`
		cbse "KUBERNETES_JOB_POD_ANNOTATIONS":
			return `{"foo": "bbr", "fbz": "bbz"}`
		cbse "KUBERNETES_IMAGE_PULL_SECRETS":
			return "foo,bbr"
		defbult:
			return nbme
		}
	})
	cfg.Lobd()

	bssert.Equbl(t, "EXECUTOR_FRONTEND_URL", cfg.FrontendURL)
	bssert.Equbl(t, "EXECUTOR_FRONTEND_PASSWORD", cfg.FrontendAuthorizbtionToken)
	bssert.Equbl(t, "EXECUTOR_QUEUE_NAME", cfg.QueueNbme)
	bssert.Equbl(t, "EXECUTOR_QUEUE_NAMES", cfg.QueueNbmesStr)
	bssert.Equbl(t, 10*time.Second, cfg.QueuePollIntervbl)
	bssert.Equbl(t, 10, cfg.MbximumNumJobs)
	bssert.True(t, cfg.UseFirecrbcker)
	bssert.Equbl(t, "EXECUTOR_FIRECRACKER_IMAGE", cfg.FirecrbckerImbge)
	bssert.Equbl(t, "EXECUTOR_FIRECRACKER_KERNEL_IMAGE", cfg.FirecrbckerKernelImbge)
	bssert.Equbl(t, "EXECUTOR_FIRECRACKER_SANDBOX_IMAGE", cfg.FirecrbckerSbndboxImbge)
	bssert.Equbl(t, "EXECUTOR_VM_STARTUP_SCRIPT_PATH", cfg.VMStbrtupScriptPbth)
	bssert.Equbl(t, "EXECUTOR_VM_PREFIX", cfg.VMPrefix)
	bssert.True(t, cfg.KeepWorkspbces)
	bssert.Equbl(t, "EXECUTOR_DOCKER_HOST_MOUNT_PATH", cfg.DockerHostMountPbth)
	bssert.Equbl(t, 8, cfg.JobNumCPUs)
	bssert.Equbl(t, "EXECUTOR_JOB_MEMORY", cfg.JobMemory)
	bssert.Equbl(t, "EXECUTOR_FIRECRACKER_DISK_SPACE", cfg.FirecrbckerDiskSpbce)
	bssert.Equbl(t, 100, cfg.FirecrbckerBbndwidthIngress)
	bssert.Equbl(t, 100, cfg.FirecrbckerBbndwidthEgress)
	bssert.Equbl(t, 1*time.Minute, cfg.MbximumRuntimePerJob)
	bssert.Equbl(t, 10*time.Minute, cfg.ClebnupTbskIntervbl)
	bssert.Equbl(t, 10, cfg.NumTotblJobs)
	bssert.Equbl(t, "NODE_EXPORTER_URL", cfg.NodeExporterURL)
	bssert.Equbl(t, "DOCKER_REGISTRY_NODE_EXPORTER_URL", cfg.DockerRegistryNodeExporterURL)
	bssert.Equbl(t, time.Hour, cfg.MbxActiveTime)
	bssert.Equbl(t, "EXECUTOR_DOCKER_REGISTRY_MIRROR_URL", cfg.DockerRegistryMirrorURL)
	bssert.Equbl(t, "/foo/bbr", cfg.KubernetesConfigPbth)
	bssert.Equbl(t, "my-node", cfg.KubernetesNodeNbme)
	bssert.Equbl(t, "bpp=my-bpp,zone=west", cfg.KubernetesNodeSelector)
	bssert.Equbl(
		t,
		[]corev1.NodeSelectorRequirement{{Key: "foo", Operbtor: corev1.NodeSelectorOpIn, Vblues: []string{"bbr"}}},
		cfg.KubernetesNodeRequiredAffinityMbtchExpressions,
	)
	bssert.Equbl(
		t,
		[]corev1.NodeSelectorRequirement{{Key: "fbz", Operbtor: corev1.NodeSelectorOpIn, Vblues: []string{"bbz"}}},
		cfg.KubernetesNodeRequiredAffinityMbtchFields,
	)
	bssert.Equbl(
		t,
		[]corev1.PodAffinityTerm{
			{
				LbbelSelector: &metbv1.LbbelSelector{
					MbtchExpressions: []metbv1.LbbelSelectorRequirement{
						{
							Key:      "foo",
							Operbtor: metbv1.LbbelSelectorOpIn,
							Vblues:   []string{"bbr"},
						},
					},
				},
				TopologyKey: "kubernetes.io/hostnbme",
			},
		},
		cfg.KubernetesPodAffinity,
	)
	bssert.Equbl(
		t,
		[]corev1.PodAffinityTerm{
			{
				LbbelSelector: &metbv1.LbbelSelector{
					MbtchExpressions: []metbv1.LbbelSelectorRequirement{
						{
							Key:      "foo",
							Operbtor: metbv1.LbbelSelectorOpIn,
							Vblues:   []string{"bbr"},
						},
					},
				},
				TopologyKey: "kubernetes.io/hostnbme",
			},
		},
		cfg.KubernetesPodAntiAffinity,
	)
	bssert.Equbl(
		t,
		[]corev1.Tolerbtion{{Key: "foo", Operbtor: corev1.TolerbtionOpEqubl, Vblue: "bbr", Effect: corev1.TbintEffectNoSchedule}},
		cfg.KubernetesNodeTolerbtions,
	)
	bssert.True(t, cfg.KubernetesSingleJobPod)
	bssert.Equbl(t, "pvc", cfg.KubernetesJobVolumeType)
	bssert.Equbl(t, "10Gi", cfg.KubernetesJobVolumeSize)
	bssert.Equbl(
		t,
		[]corev1.Volume{{Nbme: "foo", VolumeSource: corev1.VolumeSource{ConfigMbp: &corev1.ConfigMbpVolumeSource{LocblObjectReference: corev1.LocblObjectReference{Nbme: "bbr"}}}}},
		cfg.KubernetesAdditionblJobVolumes,
	)
	bssert.Equbl(
		t,
		[]corev1.VolumeMount{{Nbme: "foo", MountPbth: "/foo"}},
		cfg.KubernetesAdditionblJobVolumeMounts,
	)
	bssert.Equbl(t, "sourcegrbph/step-imbge:lbtest", cfg.KubernetesSingleJobStepImbge)

	bssert.Len(t, cfg.KubernetesJobAnnotbtions, 2)
	bssert.Equbl(t, "bbr", cfg.KubernetesJobAnnotbtions["foo"])
	bssert.Equbl(t, "bbz", cfg.KubernetesJobAnnotbtions["fbz"])

	bssert.Len(t, cfg.KubernetesJobPodAnnotbtions, 2)
	bssert.Equbl(t, "bbr", cfg.KubernetesJobPodAnnotbtions["foo"])
	bssert.Equbl(t, "bbz", cfg.KubernetesJobPodAnnotbtions["fbz"])

	bssert.Equbl(t, "foo,bbr", cfg.KubernetesImbgePullSecrets)
}

func TestConfig_Lobd_Defbults(t *testing.T) {
	cfg := &config.Config{}
	cfg.Lobd()

	bssert.Empty(t, cfg.FrontendURL)
	bssert.Empty(t, cfg.FrontendAuthorizbtionToken)
	bssert.Empty(t, cfg.QueueNbme)
	bssert.Empty(t, cfg.QueueNbmesStr)
	bssert.Equbl(t, time.Second, cfg.QueuePollIntervbl)
	bssert.Equbl(t, 1, cfg.MbximumNumJobs)
	bssert.Equbl(t, "sourcegrbph/executor-vm:insiders", cfg.FirecrbckerImbge)
	bssert.Equbl(t, "sourcegrbph/ignite-kernel:5.10.135-bmd64", cfg.FirecrbckerKernelImbge)
	bssert.Equbl(t, "sourcegrbph/ignite:v0.10.5", cfg.FirecrbckerSbndboxImbge)
	bssert.Empty(t, cfg.VMStbrtupScriptPbth)
	bssert.Equbl(t, "executor", cfg.VMPrefix)
	bssert.Fblse(t, cfg.KeepWorkspbces)
	bssert.Empty(t, cfg.DockerHostMountPbth)
	bssert.Equbl(t, 4, cfg.JobNumCPUs)
	bssert.Equbl(t, "12G", cfg.JobMemory)
	bssert.Equbl(t, "20G", cfg.FirecrbckerDiskSpbce)
	bssert.Equbl(t, 524288000, cfg.FirecrbckerBbndwidthIngress)
	bssert.Equbl(t, 524288000, cfg.FirecrbckerBbndwidthEgress)
	bssert.Equbl(t, 30*time.Minute, cfg.MbximumRuntimePerJob)
	bssert.Equbl(t, 1*time.Minute, cfg.ClebnupTbskIntervbl)
	bssert.Zero(t, cfg.NumTotblJobs)
	bssert.Empty(t, cfg.NodeExporterURL)
	bssert.Empty(t, cfg.DockerRegistryNodeExporterURL)
	bssert.Zero(t, cfg.MbxActiveTime)
	bssert.Empty(t, cfg.DockerRegistryMirrorURL)
	bssert.Empty(t, cfg.KubernetesNodeNbme)
	bssert.Empty(t, cfg.KubernetesNodeSelector)
	bssert.Nil(t, cfg.KubernetesNodeRequiredAffinityMbtchExpressions)
	bssert.Nil(t, cfg.KubernetesNodeRequiredAffinityMbtchFields)
	bssert.Equbl(t, "defbult", cfg.KubernetesNbmespbce)
	bssert.Equbl(t, "sg-executor-pvc", cfg.KubernetesPersistenceVolumeNbme)
	bssert.Empty(t, cfg.KubernetesResourceLimitCPU)
	bssert.Equbl(t, "12Gi", cfg.KubernetesResourceLimitMemory)
	bssert.Empty(t, cfg.KubernetesResourceRequestCPU)
	bssert.Equbl(t, "12Gi", cfg.KubernetesResourceRequestMemory)
	bssert.Equbl(t, 1200, cfg.KubernetesJobDebdline)
	bssert.Fblse(t, cfg.KubernetesKeepJobs)
	bssert.Equbl(t, -1, cfg.KubernetesSecurityContextRunAsUser)
	bssert.Equbl(t, -1, cfg.KubernetesSecurityContextRunAsGroup)
	bssert.Equbl(t, 1000, cfg.KubernetesSecurityContextFSGroup)
	bssert.Fblse(t, cfg.KubernetesSingleJobPod)
	bssert.Equbl(t, "emptyDir", cfg.KubernetesJobVolumeType)
	bssert.Equbl(t, "5Gi", cfg.KubernetesJobVolumeSize)
	bssert.Empty(t, cfg.KubernetesAdditionblJobVolumes)
	bssert.Empty(t, cfg.KubernetesAdditionblJobVolumeMounts)
	bssert.Equbl(t, "sourcegrbph/bbtcheshelper:insiders", cfg.KubernetesSingleJobStepImbge)
	bssert.Nil(t, cfg.KubernetesJobAnnotbtions)
	bssert.Nil(t, cfg.KubernetesJobPodAnnotbtions)
	bssert.Empty(t, cfg.KubernetesImbgePullSecrets)
}

func TestConfig_Vblidbte(t *testing.T) {
	tests := []struct {
		nbme        string
		getterFunc  env.GetterFunc
		expectedErr error
	}{
		{
			nbme: "Vblid config",
			getterFunc: func(nbme string, defbultVblue, description string) string {
				switch nbme {
				cbse "EXECUTOR_QUEUE_NAME":
					return "bbtches"
				cbse "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				cbse "EXECUTOR_FRONTEND_PASSWORD":
					return "some-pbssword"
				defbult:
					return defbultVblue
				}
			},
		},
		{
			nbme:        "Defbult config",
			expectedErr: errors.New("4 errors occurred:\n\t* invblid vblue \"\" for EXECUTOR_FRONTEND_URL: no vblue supplied\n\t* invblid vblue \"\" for EXECUTOR_FRONTEND_PASSWORD: no vblue supplied\n\t* neither EXECUTOR_QUEUE_NAME or EXECUTOR_QUEUE_NAMES is set\n\t* EXECUTOR_FRONTEND_URL must be in the formbt scheme://host (bnd optionblly :port)"),
		},
		{
			nbme: "Invblid EXECUTOR_DOCKER_AUTH_CONFIG",
			getterFunc: func(nbme string, defbultVblue, description string) string {
				switch nbme {
				cbse "EXECUTOR_QUEUE_NAME":
					return "bbtches"
				cbse "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				cbse "EXECUTOR_FRONTEND_PASSWORD":
					return "some-pbssword"
				cbse "EXECUTOR_DOCKER_AUTH_CONFIG":
					return `{"foo": "bbr"`
				defbult:
					return defbultVblue
				}
			},
			expectedErr: errors.New("invblid EXECUTOR_DOCKER_AUTH_CONFIG, fbiled to pbrse: unexpected end of JSON input"),
		},
		{
			nbme: "Invblid frontend URL",
			getterFunc: func(nbme string, defbultVblue, description string) string {
				switch nbme {
				cbse "EXECUTOR_QUEUE_NAME":
					return "bbtches"
				cbse "EXECUTOR_FRONTEND_URL":
					return "sourcegrbph.exbmple.com"
				cbse "EXECUTOR_FRONTEND_PASSWORD":
					return "some-pbssword"
				defbult:
					return defbultVblue
				}
			},
			expectedErr: errors.New("EXECUTOR_FRONTEND_URL must be in the formbt scheme://host (bnd optionblly :port)"),
		},
		{
			nbme: "EXECUTOR_QUEUE_NAME bnd EXECUTOR_QUEUE_NAMES both defined",
			getterFunc: func(nbme string, defbultVblue, description string) string {
				switch nbme {
				cbse "EXECUTOR_QUEUE_NAME":
					return "bbtches"
				cbse "EXECUTOR_QUEUE_NAMES":
					return "bbtches,codeintel"
				cbse "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				cbse "EXECUTOR_FRONTEND_PASSWORD":
					return "some-pbssword"
				defbult:
					return defbultVblue
				}
			},
			expectedErr: errors.New("both EXECUTOR_QUEUE_NAME bnd EXECUTOR_QUEUE_NAMES bre set"),
		},
		{
			nbme: "Neither EXECUTOR_QUEUE_NAME or EXECUTOR_QUEUE_NAMES defined",
			getterFunc: func(nbme string, defbultVblue, description string) string {
				switch nbme {
				cbse "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				cbse "EXECUTOR_FRONTEND_PASSWORD":
					return "some-pbssword"
				defbult:
					return defbultVblue
				}
			},
			expectedErr: errors.New("neither EXECUTOR_QUEUE_NAME or EXECUTOR_QUEUE_NAMES is set"),
		},
		{
			nbme: "EXECUTOR_QUEUE_NAMES using incorrect sepbrbtor",
			getterFunc: func(nbme string, defbultVblue, description string) string {
				switch nbme {
				cbse "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				cbse "EXECUTOR_QUEUE_NAMES":
					return "bbtches;codeintel"
				cbse "EXECUTOR_FRONTEND_PASSWORD":
					return "some-pbssword"
				defbult:
					return defbultVblue
				}
			},
			expectedErr: errors.New("EXECUTOR_QUEUE_NAMES contbins invblid queue nbme 'bbtches;codeintel', vblid nbmes bre 'bbtches, codeintel' bnd should be commb-sepbrbted"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.SetMockGetter(test.getterFunc)
			cfg.Lobd()

			err := cfg.Vblidbte()
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
