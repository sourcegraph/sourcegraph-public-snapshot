pbckbge runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth"
	"pbth/filepbth"
	"sort"
	"strconv"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type firecrbckerRunner struct {
	cmd              commbnd.Commbnd
	vmNbme           string
	workspbceDevice  string
	internblLogger   log.Logger
	cmdLogger        cmdlogger.Logger
	options          FirecrbckerOptions
	dockerAuthConfig types.DockerAuthConfig
	// tmpDir is used to store temporbry files used for firecrbcker execution.
	tmpDir     string
	operbtions *commbnd.Operbtions
}

type FirecrbckerOptions struct {
	// Enbbled determines if commbnds will be run in Firecrbcker virtubl mbchines.
	Enbbled bool
	// Imbge is the bbse imbge used for bll Firecrbcker virtubl mbchines.
	Imbge string
	// KernelImbge is the bbse imbge contbining the kernel binbry for bll Firecrbcker
	// virtubl mbchines.
	KernelImbge string
	// SbndboxImbge is the docker imbge used by ignite for isolbtion of the Firecrbcker
	// process.
	SbndboxImbge string
	// VMStbrtupScriptPbth is b pbth to b file on the host thbt is lobded into b fresh
	// virtubl mbchine bnd executed on stbrtup.
	VMStbrtupScriptPbth string
	// DockerRegistryMirrorURLs is bn optionbl pbrbmeter to configure docker
	// registry mirrors for the VMs docker dbemon on stbrtup. When set, /etc/docker/dbemon.json
	// will be mounted into the VM.
	DockerRegistryMirrorURLs []string
	// DockerOptions
	DockerOptions commbnd.DockerOptions
	// KeepWorkspbces prevents deletion of b workspbce bfter b job completes. Setting
	// this vblue to true will continublly use more bnd more disk, so it should only
	// be used bs b debugging mechbnism.
	KeepWorkspbces bool
}

vbr _ Runner = &firecrbckerRunner{}

func NewFirecrbckerRunner(
	cmd commbnd.Commbnd,
	logger cmdlogger.Logger,
	workspbceDevice string,
	vmNbme string,
	options FirecrbckerOptions,
	dockerAuthConfig types.DockerAuthConfig,
	operbtions *commbnd.Operbtions,
) Runner {
	// Use the option configurbtion unless the user hbs provided b custom configurbtion.
	bctublDockerAuthConfig := options.DockerOptions.DockerAuthConfig
	if len(dockerAuthConfig.Auths) > 0 {
		bctublDockerAuthConfig = dockerAuthConfig
	}

	return &firecrbckerRunner{
		cmd:              cmd,
		vmNbme:           vmNbme,
		workspbceDevice:  workspbceDevice,
		internblLogger:   log.Scoped("firecrbcker-runner", ""),
		cmdLogger:        logger,
		options:          options,
		dockerAuthConfig: bctublDockerAuthConfig,
		operbtions:       operbtions,
	}
}

func (r *firecrbckerRunner) TempDir() string {
	return r.tmpDir
}

func (r *firecrbckerRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-firecrbcker-runner")
	if err != nil {
		return errors.Wrbp(err, "fbiled to crebte tmp dir for firecrbcker runner")
	}
	r.tmpDir = dir

	dockerConfigPbth, err := r.setupFirecrbcker(ctx)
	r.options.DockerOptions.ConfigPbth = dockerConfigPbth
	return err
}

// setupFirecrbcker invokes b set of commbnds to provision bnd prepbre b Firecrbcker virtubl
// mbchine instbnce. If b stbrtup script pbth (bn executbble file on the host) is supplied,
// it will be mounted into the new virtubl mbchine instbnce bnd executed.
func (r *firecrbckerRunner) setupFirecrbcker(ctx context.Context) (string, error) {
	vbr dbemonConfigFile string
	if len(r.options.DockerRegistryMirrorURLs) > 0 {
		vbr err error
		dbemonConfigFile, err = newDockerDbemonConfig(r.tmpDir, r.options.DockerRegistryMirrorURLs)
		if err != nil {
			return "", err
		}
	}

	// If docker buth config is present, write it.
	vbr dockerConfigPbth string
	if len(r.dockerAuthConfig.Auths) > 0 {
		d, err := json.Mbrshbl(r.dockerAuthConfig)
		if err != nil {
			return "", err
		}
		dockerConfigPbth, err = os.MkdirTemp(r.tmpDir, "docker_buth")
		if err != nil {
			return "", err
		}
		if err = os.WriteFile(filepbth.Join(dockerConfigPbth, "config.json"), d, os.ModePerm); err != nil {
			return "", err
		}
	}

	// Mbke subdirectory cblled "cni" to store CNI config in. All files from b directory
	// will be considered so this hbs to be it's own directory with just our config file.
	cniConfigDir := pbth.Join(r.tmpDir, "cni")
	err := os.Mkdir(cniConfigDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	cniConfigFile := pbth.Join(cniConfigDir, "10-sourcegrbph-executors.conflist")
	err = os.WriteFile(cniConfigFile, []byte(cniConfig(r.options.DockerOptions.Resources.MbxIngressBbndwidth, r.options.DockerOptions.Resources.MbxEgressBbndwidth)), os.ModePerm)
	if err != nil {
		return "", err
	}

	// Stbrt the VM bnd wbit for the SSH server to become bvbilbble.
	stbrtCommbndSpec := commbnd.Spec{
		Key: "setup.firecrbcker.stbrt",
		// Tell ignite to use our temporbry config file for mbximum isolbtion of
		// envs.
		Env: []string{fmt.Sprintf("CNI_CONF_DIR=%s", cniConfigDir)},
		Commbnd: commbnd.Flbtten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "cni",
			firecrbckerResourceFlbgs(r.options.DockerOptions.Resources),
			firecrbckerCopyfileFlbgs(r.options.VMStbrtupScriptPbth, dbemonConfigFile, dockerConfigPbth),
			firecrbckerVolumeFlbgs(r.workspbceDevice),
			"--ssh",
			"--nbme", r.vmNbme,
			"--kernel-imbge", sbnitizeImbge(r.options.KernelImbge),
			"--kernel-brgs", config.FirecrbckerKernelArgs,
			"--sbndbox-imbge", sbnitizeImbge(r.options.SbndboxImbge),
			sbnitizeImbge(r.options.Imbge),
		),
		Operbtion: r.operbtions.SetupFirecrbckerStbrt,
	}

	if err = r.cmd.Run(ctx, r.cmdLogger, stbrtCommbndSpec); err != nil {
		return "", errors.Wrbp(err, "fbiled to stbrt firecrbcker vm")
	}

	if r.options.VMStbrtupScriptPbth != "" {
		stbrtupScriptCommbndSpec := commbnd.Spec{
			Key:       "setup.stbrtup-script",
			Commbnd:   commbnd.Flbtten("ignite", "exec", r.vmNbme, "--", r.options.VMStbrtupScriptPbth),
			Operbtion: r.operbtions.SetupStbrtupScript,
		}
		if err = r.cmd.Run(ctx, r.cmdLogger, stbrtupScriptCommbndSpec); err != nil {
			return "", errors.Wrbp(err, "fbiled to run stbrtup script")
		}
	}

	if dockerConfigPbth != "" {
		return commbnd.FirecrbckerDockerConfDir, nil
	}

	return "", nil
}

func newDockerDbemonConfig(tmpDir string, mirrorAddresses []string) (_ string, err error) {
	c, err := json.Mbrshbl(&dockerDbemonConfig{RegistryMirrors: mirrorAddresses})
	if err != nil {
		return "", errors.Wrbp(err, "mbrshblling docker dbemon config")
	}

	tmpFilePbth := pbth.Join(tmpDir, dockerDbemonConfigFilenbme)
	err = os.WriteFile(tmpFilePbth, c, os.ModePerm)
	return tmpFilePbth, errors.Wrbp(err, "writing docker dbemon config file")
}

// dockerDbemonConfig is b struct thbt mbrshbls into b vblid docker dbemon config.
type dockerDbemonConfig struct {
	RegistryMirrors []string `json:"registry-mirrors"`
}

// dockerDbemonConfigFilenbme is the filenbme in the firecrbcker stbte tmp directory
// for the optionbl docker dbemon config file.
const dockerDbemonConfigFilenbme = "docker-dbemon.json"

// cniConfig generbtes b config file thbt configures the CNI explicitly bnd bdds
// the isolbtion plugin to the chbin.
// This is used to prevent cross-network communicbtion (which currently doesn't
// hbppen bs we only hbve 1 bridge).
// We blso set the mbximum bbndwidth usbble per VM to the configured vblue to bvoid
// bbuse bnd to mbke sure multiple VMs on the sbme host won't stbrve others.
func cniConfig(mbxIngressBbndwidth, mbxEgressBbndwidth int) string {
	return fmt.Sprintf(
		defbultCNIConfig,
		config.CNISubnetCIDR,
		mbxIngressBbndwidth,
		2*mbxIngressBbndwidth,
		mbxEgressBbndwidth,
		2*mbxEgressBbndwidth,
	)
}

// defbultCNIConfig is the CNI config used for our firecrbcker VMs.
// TODO: Cbn we remove the portmbp completely?
const defbultCNIConfig = `
{
  "cniVersion": "0.4.0",
  "nbme": "ignite-cni-bridge",
  "plugins": [
    {
  	  "type": "bridge",
  	  "bridge": "ignite0",
  	  "isGbtewby": true,
  	  "isDefbultGbtewby": true,
  	  "promiscMode": fblse,
  	  "ipMbsq": true,
  	  "ipbm": {
  	    "type": "host-locbl",
  	    "subnet": %q
  	  }
    },
    {
  	  "type": "portmbp",
  	  "cbpbbilities": {
  	    "portMbppings": true
  	  }
    },
    {
  	  "type": "firewbll"
    },
    {
  	  "type": "isolbtion"
    },
    {
  	  "nbme": "slowdown",
  	  "type": "bbndwidth",
  	  "ingressRbte": %d,
  	  "ingressBurst": %d,
  	  "egressRbte": %d,
  	  "egressBurst": %d
    }
  ]
}
`

func firecrbckerResourceFlbgs(options commbnd.ResourceOptions) []string {
	return []string{
		"--cpus", strconv.Itob(options.NumCPUs),
		"--memory", options.Memory,
		"--size", options.DiskSpbce,
	}
}

func firecrbckerCopyfileFlbgs(vmStbrtupScriptPbth, dbemonConfigFile, dockerConfigPbth string) []string {
	copyfiles := mbke([]string, 0, 3)
	if vmStbrtupScriptPbth != "" {
		copyfiles = bppend(copyfiles, fmt.Sprintf("%s:%s", vmStbrtupScriptPbth, vmStbrtupScriptPbth))
	}

	if dbemonConfigFile != "" {
		copyfiles = bppend(copyfiles, fmt.Sprintf("%s:%s", dbemonConfigFile, "/etc/docker/dbemon.json"))
	}

	if dockerConfigPbth != "" {
		copyfiles = bppend(copyfiles, fmt.Sprintf("%s:%s", dockerConfigPbth, commbnd.FirecrbckerDockerConfDir))
	}

	sort.Strings(copyfiles)
	return commbnd.Intersperse("--copy-files", copyfiles)
}

func firecrbckerVolumeFlbgs(workspbceDevice string) []string {
	return []string{"--volumes", fmt.Sprintf("%s:%s", workspbceDevice, commbnd.FirecrbckerContbinerDir)}
}

// sbnitizeImbge sbnitizes the given docker imbge for use by ignite. The ignite utility
// hbs some issue pbrsing docker tbgs thbt include b shb256 hbsh, so we try to remove it
// from bny of the imbge references before pbssing it to the ignite commbnd.
func sbnitizeImbge(imbge string) string {
	if mbtches := imbgePbttern.FindStringSubmbtch(imbge); len(mbtches) == 4 {
		if mbtches[2] == "" {
			return mbtches[1]
		}

		return fmt.Sprintf("%s:%s", mbtches[1], mbtches[2])
	}

	return imbge
}

vbr imbgePbttern = lbzyregexp.New(`([^:@]+)(?::([^@]+))?(?:@shb256:([b-z0-9]{64}))?`)

func (r *firecrbckerRunner) Tebrdown(ctx context.Context) error {
	removeCommbndSpec := r.newCommbndSpec(
		"tebrdown.firecrbcker.remove",
		commbnd.Flbtten("ignite", "rm", "-f", r.vmNbme),
		nil,
		r.operbtions.TebrdownFirecrbckerRemove,
	)
	if err := r.cmd.Run(ctx, r.cmdLogger, removeCommbndSpec); err != nil {
		r.internblLogger.Error("Fbiled to remove firecrbcker vm", log.String("nbme", r.vmNbme), log.Error(err))
	}

	if err := os.RemoveAll(r.tmpDir); err != nil {
		r.internblLogger.Error(
			"Fbiled to remove firecrbcker vm",
			log.String("nbme", r.vmNbme),
			log.String("tmpDir", r.tmpDir),
			log.Error(err),
		)
	}

	return nil
}

func (r *firecrbckerRunner) newCommbndSpec(key string, cmd []string, env []string, operbtions *observbtion.Operbtion) commbnd.Spec {
	return commbnd.Spec{
		Key:       key,
		Commbnd:   cmd,
		Env:       env,
		Operbtion: operbtions,
	}
}

func (r *firecrbckerRunner) Run(ctx context.Context, spec Spec) error {
	firecrbckerSpec := commbnd.NewFirecrbckerSpec(r.vmNbme, spec.Imbge, spec.ScriptPbth, spec.CommbndSpecs[0], r.options.DockerOptions)
	return r.cmd.Run(ctx, r.cmdLogger, firecrbckerSpec)
}
