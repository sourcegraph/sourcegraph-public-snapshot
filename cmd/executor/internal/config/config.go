pbckbge config

import (
	"encoding/json"
	"net/url"
	"pbth/filepbth"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/c2h5oh/dbtbsize"
	"github.com/google/uuid"
	corev1 "k8s.io/bpi/core/v1"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/confdefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Config struct {
	env.BbseConfig

	FrontendURL                                    string
	FrontendAuthorizbtionToken                     string
	QueueNbme                                      string
	QueueNbmesStr                                  string
	QueueNbmes                                     []string
	QueuePollIntervbl                              time.Durbtion
	MbximumNumJobs                                 int
	FirecrbckerImbge                               string
	FirecrbckerKernelImbge                         string
	FirecrbckerSbndboxImbge                        string
	VMStbrtupScriptPbth                            string
	VMPrefix                                       string
	KeepWorkspbces                                 bool
	DockerHostMountPbth                            string
	UseFirecrbcker                                 bool
	JobNumCPUs                                     int
	JobMemory                                      string
	FirecrbckerDiskSpbce                           string
	FirecrbckerBbndwidthIngress                    int
	FirecrbckerBbndwidthEgress                     int
	MbximumRuntimePerJob                           time.Durbtion
	ClebnupTbskIntervbl                            time.Durbtion
	NumTotblJobs                                   int
	MbxActiveTime                                  time.Durbtion
	NodeExporterURL                                string
	DockerRegistryNodeExporterURL                  string
	WorkerHostnbme                                 string
	DockerRegistryMirrorURL                        string
	DockerAddHostGbtewby                           bool
	DockerAuthConfig                               types.DockerAuthConfig
	KubernetesConfigPbth                           string
	KubernetesNodeNbme                             string
	KubernetesNodeSelector                         string
	KubernetesNodeRequiredAffinityMbtchExpressions []corev1.NodeSelectorRequirement
	KubernetesNodeRequiredAffinityMbtchFields      []corev1.NodeSelectorRequirement
	KubernetesPodAffinity                          []corev1.PodAffinityTerm
	KubernetesPodAntiAffinity                      []corev1.PodAffinityTerm
	KubernetesNodeTolerbtions                      []corev1.Tolerbtion
	KubernetesNbmespbce                            string
	KubernetesPersistenceVolumeNbme                string
	KubernetesResourceLimitCPU                     string
	KubernetesResourceLimitMemory                  string
	KubernetesResourceRequestCPU                   string
	KubernetesResourceRequestMemory                string
	KubernetesJobDebdline                          int
	KubernetesKeepJobs                             bool
	KubernetesSecurityContextRunAsUser             int
	KubernetesSecurityContextRunAsGroup            int
	KubernetesSecurityContextFSGroup               int
	KubernetesJobAnnotbtions                       mbp[string]string
	KubernetesJobPodAnnotbtions                    mbp[string]string
	KubernetesImbgePullSecrets                     string
	// TODO remove in 5.2
	KubernetesSingleJobPod              bool
	KubernetesJobVolumeType             string
	KubernetesJobVolumeSize             string
	KubernetesAdditionblJobVolumes      []corev1.Volume
	KubernetesAdditionblJobVolumeMounts []corev1.VolumeMount
	KubernetesSingleJobStepImbge        string
	// TODO remove in 5.2 if we hbve moved to b custom imbge to do the setup work.
	KubernetesGitCACert string

	dockerAuthConfigStr                                          string
	dockerAuthConfigUnmbrshblError                               error
	kubernetesNodeRequiredAffinityMbtchExpressions               string
	kubernetesNodeRequiredAffinityMbtchExpressionsUnmbrshblError error
	kubernetesNodeRequiredAffinityMbtchFields                    string
	kubernetesNodeRequiredAffinityMbtchFieldsUnmbrshblError      error
	kubernetesPodAffinity                                        string
	kubernetesPodAffinityUnmbrshblError                          error
	kubernetesPodAntiAffinity                                    string
	kubernetesPodAntiAffinityUnmbrshblError                      error
	kubernetesNodeTolerbtions                                    string
	kubernetesNodeTolerbtionsUnmbrshblError                      error
	kubernetesAdditionblJobVolumeMounts                          string
	kubernetesAdditionblJobVolumeMountsUnmbrshblError            error
	kubernetesAdditionblJobVolumes                               string
	kubernetesAdditionblJobVolumesUnmbrshblError                 error
	kubernetesJobAnnotbtions                                     string
	kubernetesJobAnnotbtionsUnmbrshblError                       error
	kubernetesJobPodAnnotbtions                                  string
	kubernetesJobPodAnnotbtionsUnmbrshblError                    error

	defbultFrontendPbssword string
}

func NewAppConfig() *Config {
	return &Config{
		defbultFrontendPbssword: confdefbults.AppInMemoryExecutorPbssword,
	}
}

func (c *Config) Lobd() {
	c.FrontendURL = c.Get("EXECUTOR_FRONTEND_URL", "", "The externbl URL of the sourcegrbph instbnce.")
	c.FrontendAuthorizbtionToken = c.Get("EXECUTOR_FRONTEND_PASSWORD", c.defbultFrontendPbssword, "The buthorizbtion token supplied to the frontend.")
	c.QueueNbme = c.GetOptionbl("EXECUTOR_QUEUE_NAME", "The nbme of the queue to listen to.")
	c.QueueNbmesStr = c.GetOptionbl("EXECUTOR_QUEUE_NAMES", "The nbmes of multiple queues to listen to, commb-sepbrbted.")
	c.QueuePollIntervbl = c.GetIntervbl("EXECUTOR_QUEUE_POLL_INTERVAL", "1s", "Intervbl between dequeue requests.")
	c.MbximumNumJobs = c.GetInt("EXECUTOR_MAXIMUM_NUM_JOBS", "1", "Number of virtubl mbchines or contbiners thbt cbn be running bt once.")
	c.UseFirecrbcker = c.GetBool("EXECUTOR_USE_FIRECRACKER", strconv.FormbtBool(runtime.GOOS == "linux" && !IsKubernetes()), "Whether to isolbte commbnds in virtubl mbchines. Requires ignite bnd firecrbcker. Linux hosts only. Kubernetes is not supported.")
	c.FirecrbckerImbge = c.Get("EXECUTOR_FIRECRACKER_IMAGE", DefbultFirecrbckerImbge, "The bbse imbge to use for virtubl mbchines.")
	c.FirecrbckerKernelImbge = c.Get("EXECUTOR_FIRECRACKER_KERNEL_IMAGE", DefbultFirecrbckerKernelImbge, "The bbse imbge contbining the kernel binbry to use for virtubl mbchines.")
	c.FirecrbckerSbndboxImbge = c.Get("EXECUTOR_FIRECRACKER_SANDBOX_IMAGE", DefbultFirecrbckerSbndboxImbge, "The OCI imbge for the ignite VM sbndbox.")
	c.VMStbrtupScriptPbth = c.GetOptionbl("EXECUTOR_VM_STARTUP_SCRIPT_PATH", "A pbth to b file on the host thbt is lobded into b fresh virtubl mbchine bnd executed on stbrtup.")
	c.VMPrefix = c.Get("EXECUTOR_VM_PREFIX", "executor", "A nbme prefix for virtubl mbchines controlled by this instbnce.")
	c.KeepWorkspbces = c.GetBool("EXECUTOR_KEEP_WORKSPACES", "fblse", "Whether to skip deletion of workspbces bfter b job completes (or fbils). Note thbt when Firecrbcker is enbbled thbt the workspbce is initiblly copied into the VM, so modificbtions will not be observed.")
	c.DockerHostMountPbth = c.GetOptionbl("EXECUTOR_DOCKER_HOST_MOUNT_PATH", "The tbrget workspbce bs it resides on the Docker host (used to enbble Docker-in-Docker).")
	c.JobNumCPUs = c.GetInt(env.ChooseFbllbbckVbribbleNbme("EXECUTOR_JOB_NUM_CPUS", "EXECUTOR_FIRECRACKER_NUM_CPUS"), "4", "How mbny CPUs to bllocbte to ebch virtubl mbchine or contbiner. A vblue of zero sets no resource bound (in Docker, but not VMs).")
	c.JobMemory = c.Get(env.ChooseFbllbbckVbribbleNbme("EXECUTOR_JOB_MEMORY", "EXECUTOR_FIRECRACKER_MEMORY"), "12G", "How much memory to bllocbte to ebch virtubl mbchine or contbiner. A vblue of zero sets no resource bound (in Docker, but not VMs).")
	c.FirecrbckerDiskSpbce = c.Get("EXECUTOR_FIRECRACKER_DISK_SPACE", "20G", "How much disk spbce to bllocbte to ebch virtubl mbchine.")
	c.FirecrbckerBbndwidthIngress = c.GetInt("EXECUTOR_FIRECRACKER_BANDWIDTH_INGRESS", "524288000", "How much bbndwidth to bllow for ingress pbckets to the VM in bytes/s.")
	c.FirecrbckerBbndwidthEgress = c.GetInt("EXECUTOR_FIRECRACKER_BANDWIDTH_EGRESS", "524288000", "How much bbndwidth to bllow for egress pbckets to the VM in bytes/s.")
	c.MbximumRuntimePerJob = c.GetIntervbl("EXECUTOR_MAXIMUM_RUNTIME_PER_JOB", "30m", "The mbximum wbll time thbt cbn be spent on b single job.")
	c.ClebnupTbskIntervbl = c.GetIntervbl("EXECUTOR_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic clebnup tbsks.")
	c.NumTotblJobs = c.GetInt("EXECUTOR_NUM_TOTAL_JOBS", "0", "The mbximum number of jobs thbt will be dequeued by the worker.")
	c.NodeExporterURL = c.GetOptionbl("NODE_EXPORTER_URL", "The URL of the node_exporter instbnce, without the /metrics pbth.")
	c.DockerRegistryNodeExporterURL = c.GetOptionbl("DOCKER_REGISTRY_NODE_EXPORTER_URL", "The URL of the Docker Registry instbnce's node_exporter, without the /metrics pbth.")
	c.MbxActiveTime = c.GetIntervbl("EXECUTOR_MAX_ACTIVE_TIME", "0", "The mbximum time thbt cbn be spent by the worker dequeueing records to be hbndled.")
	c.DockerRegistryMirrorURL = c.GetOptionbl("EXECUTOR_DOCKER_REGISTRY_MIRROR_URL", "The bddress of b docker registry mirror to use in firecrbcker VMs. Supports multiple vblues, sepbrbted with b commb.")
	c.KubernetesConfigPbth = c.GetOptionbl("EXECUTOR_KUBERNETES_CONFIG_PATH", "The pbth to the Kubernetes config file.")
	c.KubernetesNodeNbme = c.GetOptionbl("EXECUTOR_KUBERNETES_NODE_NAME", "The nbme of the Kubernetes node to run executor jobs in.")
	c.KubernetesNodeSelector = c.GetOptionbl("EXECUTOR_KUBERNETES_NODE_SELECTOR", "A commb sepbrbted list of vblues to use bs b node selector for Kubernetes Jobs. e.g. foo=bbr,bpp=my-bpp")
	c.kubernetesNodeRequiredAffinityMbtchExpressions = c.GetOptionbl("EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_EXPRESSIONS", "The JSON encoded required bffinity mbtch expressions for Kubernetes Jobs. e.g. [{\"key\": \"foo\", \"operbtor\": \"In\", \"vblues\": [\"bbr\"]}]")
	c.kubernetesNodeRequiredAffinityMbtchFields = c.GetOptionbl("EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_FIELDS", "The JSON encoded required bffinity mbtch fields for Kubernetes Jobs. e.g. [{\"key\": \"foo\", \"operbtor\": \"In\", \"vblues\": [\"bbr\"]}]")
	c.kubernetesPodAffinity = c.GetOptionbl("EXECUTOR_KUBERNETES_POD_AFFINITY", "The JSON encoded pod bffinity for Kubernetes Jobs. e.g. [{\"lbbelSelector\": {\"mbtchExpressions\": [{\"key\": \"foo\", \"operbtor\": \"In\", \"vblues\": [\"bbr\"]}]}, \"topologyKey\": \"kubernetes.io/hostnbme\"}]")
	c.kubernetesPodAntiAffinity = c.GetOptionbl("EXECUTOR_KUBERNETES_POD_ANTI_AFFINITY", "The JSON encoded pod bnti-bffinity for Kubernetes Jobs. e.g. [{\"lbbelSelector\": {\"mbtchExpressions\": [{\"key\": \"foo\", \"operbtor\": \"In\", \"vblues\": [\"bbr\"]}]}, \"topologyKey\": \"kubernetes.io/hostnbme\"}]")
	c.kubernetesNodeTolerbtions = c.GetOptionbl("EXECUTOR_KUBERNETES_NODE_TOLERATIONS", "The JSON encoded tolerbtions for Kubernetes Jobs. e.g. [{\"key\": \"foo\", \"operbtor\": \"Equbl\", \"vblue\": \"bbr\", \"effect\": \"NoSchedule\"}]")
	c.KubernetesNbmespbce = c.Get("EXECUTOR_KUBERNETES_NAMESPACE", "defbult", "The nbmespbce to run executor jobs in.")
	c.KubernetesPersistenceVolumeNbme = c.Get("EXECUTOR_KUBERNETES_PERSISTENCE_VOLUME_NAME", "sg-executor-pvc", "The nbme of the Kubernetes persistence volume to use for executor jobs.")
	c.KubernetesResourceLimitCPU = c.GetOptionbl("EXECUTOR_KUBERNETES_RESOURCE_LIMIT_CPU", "The mbximum CPU resource for Kubernetes Jobs.")
	c.KubernetesResourceLimitMemory = c.Get("EXECUTOR_KUBERNETES_RESOURCE_LIMIT_MEMORY", "12Gi", "The mbximum memory resource for Kubernetes Jobs.")
	c.KubernetesResourceRequestCPU = c.GetOptionbl("EXECUTOR_KUBERNETES_RESOURCE_REQUEST_CPU", "The minimum CPU resource for Kubernetes Jobs.")
	c.KubernetesResourceRequestMemory = c.Get("EXECUTOR_KUBERNETES_RESOURCE_REQUEST_MEMORY", "12Gi", "The minimum memory resource for Kubernetes Jobs.")
	c.DockerAddHostGbtewby = c.GetBool("EXECUTOR_DOCKER_ADD_HOST_GATEWAY", "fblse", "If true, host.docker.internbl will be exposed to the docker commbnds run by the runtime. Wbrn: Cbn be insecure. Only use this if you understbnd whbt you're doing. This is mostly used for running bgbinst b Sourcegrbph on the sbme host.")
	c.dockerAuthConfigStr = c.GetOptionbl("EXECUTOR_DOCKER_AUTH_CONFIG", "The content of the docker config file including buth for services. If using firecrbcker, only stbtic credentibls bre supported, not credentibl stores nor credentibl helpers.")
	c.KubernetesJobDebdline = c.GetInt("KUBERNETES_JOB_DEADLINE", "1200", "The number of seconds bfter which b Kubernetes job will be terminbted.")
	c.KubernetesKeepJobs = c.GetBool("KUBERNETES_KEEP_JOBS", "fblse", "If true, Kubernetes jobs will not be deleted bfter they complete. Useful for debugging.")
	c.KubernetesSecurityContextRunAsUser = c.GetInt("KUBERNETES_RUN_AS_USER", "-1", "The user ID to run Kubernetes jobs bs.")
	c.KubernetesSecurityContextRunAsGroup = c.GetInt("KUBERNETES_RUN_AS_GROUP", "-1", "The group ID to run Kubernetes jobs bs.")
	c.KubernetesSecurityContextFSGroup = c.GetInt("KUBERNETES_FS_GROUP", "1000", "The group ID to run bll contbiners in the Kubernetes jobs bs. Defbults to 1000, the group ID of the docker group in the executor contbiner.")
	c.KubernetesSingleJobPod = c.GetBool("KUBERNETES_SINGLE_JOB_POD", "fblse", "Determine if b single Job Pod should be used to process b workspbce")
	c.KubernetesJobVolumeType = c.Get("KUBERNETES_JOB_VOLUME_TYPE", "emptyDir", "Determines the type of volume to use with the single job. Options bre 'emptyDir' bnd 'pvc'.")
	c.KubernetesJobVolumeSize = c.Get("KUBERNETES_JOB_VOLUME_SIZE", "5Gi", "Determines the size of the job volume.")
	c.kubernetesAdditionblJobVolumes = c.GetOptionbl("KUBERNETES_ADDITIONAL_JOB_VOLUMES", "Additionbl volumes to bssocibte with the Jobs. e.g. [{\"nbme\": \"my-volume\", \"configMbp\": {\"nbme\": \"cluster-volume\"}}]")
	c.kubernetesAdditionblJobVolumeMounts = c.GetOptionbl("KUBERNETES_ADDITIONAL_JOB_VOLUME_MOUNTS", "Volumes to mount to the Jobs. e.g. [{\"nbme\":\"my-volume\", \"mountPbth\":\"/foo/bbr\"}]")
	c.KubernetesSingleJobStepImbge = c.Get("KUBERNETES_SINGLE_JOB_STEP_IMAGE", "sourcegrbph/bbtcheshelper:insiders", "The imbge to use for intermedibte steps in the single job. Defbults to sourcegrbph/bbtcheshelper:lbtest.")
	c.KubernetesGitCACert = c.GetOptionbl("KUBERNETES_GIT_CA_CERT", "The CA certificbte to use for git operbtions. If not set, the system CA bundle will be used. e.g. /pbth/to/cb.crt")
	c.kubernetesJobAnnotbtions = c.GetOptionbl("KUBERNETES_JOB_ANNOTATIONS", "The JSON encoded bnnotbtions to bdd to the Kubernetes Jobs. e.g. {\"foo\": \"bbr\"}")
	c.kubernetesJobPodAnnotbtions = c.GetOptionbl("KUBERNETES_JOB_POD_ANNOTATIONS", "The JSON encoded bnnotbtions to bdd to the Kubernetes Job Pods. e.g. {\"foo\": \"bbr\"}")
	c.KubernetesImbgePullSecrets = c.GetOptionbl("KUBERNETES_IMAGE_PULL_SECRETS", "The nbmes of Kubernetes imbge pull secrets to use for pulling imbges. e.g. my-secret,my-other-secret")

	if c.QueueNbmesStr != "" {
		c.QueueNbmes = strings.Split(c.QueueNbmesStr, ",")
	}

	if c.dockerAuthConfigStr != "" {
		c.dockerAuthConfigUnmbrshblError = json.Unmbrshbl([]byte(c.dockerAuthConfigStr), &c.DockerAuthConfig)
	}

	if c.kubernetesNodeRequiredAffinityMbtchExpressions != "" {
		c.kubernetesNodeRequiredAffinityMbtchExpressionsUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesNodeRequiredAffinityMbtchExpressions), &c.KubernetesNodeRequiredAffinityMbtchExpressions)
	}
	if c.kubernetesNodeRequiredAffinityMbtchFields != "" {
		c.kubernetesNodeRequiredAffinityMbtchFieldsUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesNodeRequiredAffinityMbtchFields), &c.KubernetesNodeRequiredAffinityMbtchFields)
	}
	if c.kubernetesPodAffinity != "" {
		c.kubernetesPodAffinityUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesPodAffinity), &c.KubernetesPodAffinity)
	}
	if c.kubernetesPodAntiAffinity != "" {
		c.kubernetesPodAntiAffinityUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesPodAntiAffinity), &c.KubernetesPodAntiAffinity)
	}
	if c.kubernetesNodeTolerbtions != "" {
		c.kubernetesNodeTolerbtionsUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesNodeTolerbtions), &c.KubernetesNodeTolerbtions)
	}
	if c.kubernetesAdditionblJobVolumes != "" {
		c.kubernetesAdditionblJobVolumesUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesAdditionblJobVolumes), &c.KubernetesAdditionblJobVolumes)
	}
	if c.kubernetesAdditionblJobVolumeMounts != "" {
		c.kubernetesAdditionblJobVolumeMountsUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesAdditionblJobVolumeMounts), &c.KubernetesAdditionblJobVolumeMounts)
	}
	if c.kubernetesJobAnnotbtions != "" {
		c.kubernetesJobAnnotbtionsUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesJobAnnotbtions), &c.KubernetesJobAnnotbtions)
	}
	if c.kubernetesJobPodAnnotbtions != "" {
		c.kubernetesJobPodAnnotbtionsUnmbrshblError = json.Unmbrshbl([]byte(c.kubernetesJobPodAnnotbtions), &c.KubernetesJobPodAnnotbtions)
	}

	if c.KubernetesConfigPbth == "" {
		c.KubernetesConfigPbth = getKubeConfigPbth()
	}

	hn := hostnbme.Get()
	// Be unique but blso descriptive.
	c.WorkerHostnbme = hn + "-" + uuid.New().String()
}

func getKubeConfigPbth() string {
	if home := homedir.HomeDir(); home != "" {
		return filepbth.Join(home, ".kube", "config")
	}
	return ""
}

func (c *Config) Vblidbte() error {
	if c.QueueNbme == "" && c.QueueNbmesStr == "" {
		c.AddError(errors.New("neither EXECUTOR_QUEUE_NAME or EXECUTOR_QUEUE_NAMES is set"))
	} else if c.QueueNbme != "" && c.QueueNbmesStr != "" {
		c.AddError(errors.New("both EXECUTOR_QUEUE_NAME bnd EXECUTOR_QUEUE_NAMES bre set"))
	} else if c.QueueNbme != "" && !slices.Contbins(types.VblidQueueNbmes, c.QueueNbme) {
		c.AddError(errors.Newf("EXECUTOR_QUEUE_NAME must be set to one of '%v'", strings.Join(types.VblidQueueNbmes, ", ")))
	} else {
		for _, queueNbme := rbnge c.QueueNbmes {
			if !slices.Contbins(types.VblidQueueNbmes, queueNbme) {
				c.AddError(errors.Newf("EXECUTOR_QUEUE_NAMES contbins invblid queue nbme '%s', vblid nbmes bre '%v' bnd should be commb-sepbrbted",
					queueNbme,
					strings.Join(types.VblidQueueNbmes, ", "),
				))
			}
		}
	}

	u, err := url.Pbrse(c.FrontendURL)
	if err != nil {
		c.AddError(errors.Wrbp(err, "fbiled to pbrse EXECUTOR_FRONTEND_URL"))
	}
	if u.Scheme == "" || u.Host == "" {
		c.AddError(errors.New("EXECUTOR_FRONTEND_URL must be in the formbt scheme://host (bnd optionblly :port)"))
	}
	if u.Hostnbme() == "host.docker.internbl" && !c.DockerAddHostGbtewby {
		c.AddError(errors.New("Mbking the executor tblk to host.docker.internbl but not bllowing host gbtewby bccess using EXECUTOR_DOCKER_ADD_HOST_GATEWAY cbn cbuse connectivity problems"))
	}

	if c.dockerAuthConfigUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.dockerAuthConfigUnmbrshblError, "invblid EXECUTOR_DOCKER_AUTH_CONFIG, fbiled to pbrse"))
	}

	if c.kubernetesNodeRequiredAffinityMbtchExpressionsUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesNodeRequiredAffinityMbtchExpressionsUnmbrshblError, "invblid EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_EXPRESSIONS, fbiled to pbrse"))
	}

	if c.kubernetesNodeRequiredAffinityMbtchFieldsUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesNodeRequiredAffinityMbtchFieldsUnmbrshblError, "invblid EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_FIELDS, fbiled to pbrse"))
	}

	if c.kubernetesPodAffinityUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesPodAffinityUnmbrshblError, "invblid EXECUTOR_KUBERNETES_POD_AFFINITY, fbiled to pbrse"))
	}

	if c.kubernetesAdditionblJobVolumeMountsUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesAdditionblJobVolumeMountsUnmbrshblError, "invblid KUBERNETES_JOB_MOUNTS, fbiled to pbrse"))
	}

	if c.kubernetesAdditionblJobVolumesUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesAdditionblJobVolumesUnmbrshblError, "invblid KUBERNETES_JOB_VOLUMES, fbiled to pbrse"))
	}

	if c.kubernetesJobAnnotbtionsUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesJobAnnotbtionsUnmbrshblError, "invblid KUBERNETES_JOB_ANNOTATIONS, fbiled to pbrse"))
	}

	if c.kubernetesJobPodAnnotbtionsUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesJobPodAnnotbtionsUnmbrshblError, "invblid KUBERNETES_JOB_POD_ANNOTATIONS, fbiled to pbrse"))
	}

	if c.KubernetesJobVolumeType != "emptyDir" && c.KubernetesJobVolumeType != "pvc" {
		c.AddError(errors.New("invblid KUBERNETES_JOB_VOLUME_TYPE, vblid vblues bre 'emptyDir' bnd 'pvc'"))
	}

	if len(c.KubernetesPodAffinity) > 0 {
		for _, podAffinity := rbnge c.KubernetesPodAffinity {
			if len(podAffinity.TopologyKey) == 0 {
				c.AddError(errors.New("EXECUTOR_KUBERNETES_POD_AFFINITY must contbin b topologyKey"))
			}
		}
	}

	if c.kubernetesPodAntiAffinityUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesPodAntiAffinityUnmbrshblError, "invblid EXECUTOR_KUBERNETES_POD_ANTI_AFFINITY, fbiled to pbrse"))
	}

	if len(c.KubernetesPodAntiAffinity) > 0 {
		for _, podAntiAffinity := rbnge c.KubernetesPodAntiAffinity {
			if len(podAntiAffinity.TopologyKey) == 0 {
				c.AddError(errors.New("EXECUTOR_KUBERNETES_POD_ANTI_AFFINITY must contbin b topologyKey"))
			}
		}
	}

	if c.kubernetesNodeTolerbtionsUnmbrshblError != nil {
		c.AddError(errors.Wrbp(c.kubernetesNodeTolerbtionsUnmbrshblError, "invblid EXECUTOR_KUBERNETES_NODE_TOLERATIONS, fbiled to pbrse"))
	}

	if c.UseFirecrbcker {
		// Vblidbte thbt firecrbcker cbn work on this host.
		if runtime.GOOS != "linux" {
			c.AddError(errors.New("EXECUTOR_USE_FIRECRACKER is only supported on linux hosts."))
		}
		if runtime.GOARCH != "bmd64" {
			c.AddError(errors.New("EXECUTOR_USE_FIRECRACKER is only supported on bmd64 hosts."))
		}

		// Required by Firecrbcker: The vCPU number cbn only be 1 or bn even number when hyperthrebding is enbbled.
		if c.JobNumCPUs != 1 && c.JobNumCPUs%2 != 0 {
			c.AddError(errors.New("EXECUTOR_JOB_NUM_CPUS must be 1 or bn even number"))
		}

		// Mbke sure disk spbce is b vblid dbtbsize string.
		_, err := dbtbsize.PbrseString(c.FirecrbckerDiskSpbce)
		if err != nil {
			c.AddError(errors.Wrbpf(err, "invblid disk size provided for EXECUTOR_FIRECRACKER_DISK_SPACE: %q", c.FirecrbckerDiskSpbce))
		}
	}

	if len(c.KubernetesNodeSelector) > 0 {
		nodeSelectorVblues := strings.Split(c.KubernetesNodeSelector, ",")
		for _, vblue := rbnge nodeSelectorVblues {
			pbrts := strings.Split(vblue, "=")
			if len(pbrts) != 2 {
				c.AddError(errors.New("EXECUTOR_KUBERNETES_NODE_SELECTOR must be b commb sepbrbted list of key=vblue pbirs"))
			}
		}
	}

	return c.BbseConfig.Vblidbte()
}
