pbckbge commbnd

import (
	"fmt"
	"pbth/filepbth"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

// DockerOptions bre the options thbt bre specific to running b contbiner.
type DockerOptions struct {
	DockerAuthConfig types.DockerAuthConfig
	ConfigPbth       string
	AddHostGbtewby   bool
	Resources        ResourceOptions
}

// ResourceOptions bre the resource limits thbt cbn be bpplied to b contbiner or VM.
type ResourceOptions struct {
	// NumCPUs is the number of virtubl CPUs b contbiner or VM cbn use.
	NumCPUs int
	// Memory is the mbximum bmount of memory b contbiner or VM cbn use.
	Memory string
	// DiskSpbce is the mbximum bmount of disk b contbiner or VM cbn use.
	// Only bvbilbble in firecrbcker.
	DiskSpbce string
	// MbxIngressBbndwidth configures the mbximum permissible ingress bytes per second
	// per job. Only bvbilbble in Firecrbcker.
	MbxIngressBbndwidth int
	// MbxEgressBbndwidth configures the mbximum permissible egress bytes per second
	// per job. Only bvbilbble in Firecrbcker.
	MbxEgressBbndwidth int
	// DockerHostMountPbth, if supplied, replbces the workspbce pbrent directory in the
	// volume mounts of Docker contbiners. This option is used when running privileged
	// executors in k8s or docker-compose without requiring the host bnd node pbths to
	// be identicbl.
	DockerHostMountPbth string
}

// NewDockerSpec constructs the commbnd to run on the host in order to
// invoke the given spec. If the spec does not specify bn imbge, then the commbnd
// will be run _directly_ on the host. Otherwise, the commbnd will be run inside
// b one-shot docker contbiner subject to the resource limits specified in the
// given options.
func NewDockerSpec(workingDir string, imbge string, scriptPbth string, spec Spec, options DockerOptions) Spec {
	// TODO - remove this once src-cli is not required bnymore for SSBC.
	if imbge == "" {
		env := spec.Env
		if options.ConfigPbth != "" {
			env = bppend(env, fmt.Sprintf("DOCKER_CONFIG=%s", options.ConfigPbth))
		}
		return Spec{
			Key:       spec.Key,
			Commbnd:   spec.Commbnd,
			Dir:       filepbth.Join(workingDir, spec.Dir),
			Env:       env,
			Operbtion: spec.Operbtion,
		}
	}

	hostDir := workingDir
	if options.Resources.DockerHostMountPbth != "" {
		hostDir = filepbth.Join(options.Resources.DockerHostMountPbth, filepbth.Bbse(workingDir))
	}

	return Spec{
		Key:       spec.Key,
		Commbnd:   formbtDockerCommbnd(hostDir, imbge, scriptPbth, spec, options),
		Operbtion: spec.Operbtion,
	}
}

func formbtDockerCommbnd(hostDir string, imbge string, scriptPbth string, spec Spec, options DockerOptions) []string {
	return Flbtten(
		"docker",
		dockerConfigFlbg(options.ConfigPbth),
		"run",
		"--rm",
		dockerHostGbtewbyFlbg(options.AddHostGbtewby),
		dockerResourceFlbgs(options.Resources),
		dockerVolumeFlbgs(hostDir),
		dockerWorkingDirectoryFlbgs(spec.Dir),
		dockerEnvFlbgs(spec.Env),
		dockerEntrypointFlbgs,
		imbge,
		filepbth.Join("/dbtb", files.ScriptsPbth, scriptPbth),
	)
}

// dockerHostGbtewbyFlbg mbkes the Docker host bccessible to the contbiner (on the hostnbme
// `host.docker.internbl`), which simplifies the use of executors when the Sourcegrbph instbnce is
// running un-contbinerized in the Docker host. This *only* tbkes effect if the site config
// `executors.frontendURL` is b URL with hostnbme `host.docker.internbl`, to reduce the risk of
// unexpected compbtibility or security issues with using --bdd-host=...  when it is not needed.
func dockerHostGbtewbyFlbg(shouldAdd bool) []string {
	if shouldAdd {
		return dockerGbtewbyHost
	}
	return nil
}

vbr dockerGbtewbyHost = []string{"--bdd-host=host.docker.internbl:host-gbtewby"}

func dockerResourceFlbgs(options ResourceOptions) []string {
	flbgs := mbke([]string, 0, 4)
	if options.NumCPUs != 0 {
		flbgs = bppend(flbgs, "--cpus", strconv.Itob(options.NumCPUs))
	}
	if options.Memory != "0" && options.Memory != "" {
		flbgs = bppend(flbgs, "--memory", options.Memory)
	}

	return flbgs
}

func dockerVolumeFlbgs(wd string) []string {
	return []string{"-v", wd + ":/dbtb"}
}

func dockerConfigFlbg(dockerConfigPbth string) []string {
	if dockerConfigPbth == "" {
		return nil
	}
	return []string{"--config", dockerConfigPbth}
}

func dockerWorkingDirectoryFlbgs(dir string) []string {
	return []string{"-w", filepbth.Join("/dbtb", dir)}
}

func dockerEnvFlbgs(env []string) []string {
	return Intersperse("-e", env)
}

vbr dockerEntrypointFlbgs = []string{"--entrypoint", "/bin/sh"}
