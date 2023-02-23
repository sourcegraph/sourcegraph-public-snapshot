package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type firecrackerRunner struct {
	vmName          string
	workspaceDevice string
	cmdLogger       command.Logger
	options         FirecrackerOptions
	// tmpDir is used to store temporary files used for firecracker execution.
	tmpDir     string
	operations *Operations
}

type FirecrackerOptions struct {
	// Enabled determines if commands will be run in Firecracker virtual machines.
	Enabled bool
	// Image is the base image used for all Firecracker virtual machines.
	Image string
	// KernelImage is the base image containing the kernel binary for all Firecracker
	// virtual machines.
	KernelImage string
	// SandboxImage is the docker image used by ignite for isolation of the Firecracker
	// process.
	SandboxImage string
	// VMStartupScriptPath is a path to a file on the host that is loaded into a fresh
	// virtual machine and executed on startup.
	VMStartupScriptPath string
	// DockerRegistryMirrorURLs is an optional parameter to configure docker
	// registry mirrors for the VMs docker daemon on startup. When set, /etc/docker/daemon.json
	// will be mounted into the VM.
	DockerRegistryMirrorURLs []string
	// DockerOptions
	DockerOptions command.DockerOptions
}

var _ Runner = &firecrackerRunner{}

func NewFirecrackerRunner(dir string, logger command.Logger, vmName string, options FirecrackerOptions, operations *Operations) Runner {
	return &firecrackerRunner{
		vmName:          vmName,
		workspaceDevice: dir,
		cmdLogger:       logger,
		options:         options,
		operations:      operations,
	}
}

func (r *firecrackerRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-firecracker-runner")
	if err != nil {
		return errors.Wrap(err, "failed to create tmp dir for firecracker runner")
	}
	r.tmpDir = dir

	dockerConfigPath, err := r.setupFirecracker(ctx)
	r.options.DockerOptions.ConfigPath = dockerConfigPath
	return err
}

// setupFirecracker invokes a set of commands to provision and prepare a Firecracker virtual
// machine instance. If a startup script path (an executable file on the host) is supplied,
// it will be mounted into the new virtual machine instance and executed.
func (r *firecrackerRunner) setupFirecracker(ctx context.Context) (string, error) {
	var daemonConfigFile string
	if len(r.options.DockerRegistryMirrorURLs) > 0 {
		var err error
		daemonConfigFile, err = newDockerDaemonConfig(r.tmpDir, r.options.DockerRegistryMirrorURLs)
		if err != nil {
			return "", err
		}
	}

	// If docker auth config is present, write it.
	var dockerConfigPath string
	if len(r.options.DockerOptions.DockerAuthConfig.Auths) > 0 {
		d, err := json.Marshal(r.options.DockerOptions.DockerAuthConfig)
		if err != nil {
			return "", err
		}
		dockerConfigPath, err = os.MkdirTemp(r.tmpDir, "docker_auth")
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(dockerConfigPath, "config.json"), d, os.ModePerm); err != nil {
			return "", err
		}
	}

	// Make subdirectory called "cni" to store CNI config in. All files from a directory
	// will be considered so this has to be it's own directory with just our config file.
	cniConfigDir := path.Join(r.tmpDir, "cni")
	err := os.Mkdir(cniConfigDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	cniConfigFile := path.Join(cniConfigDir, "10-sourcegraph-executors.conflist")
	err = os.WriteFile(cniConfigFile, []byte(cniConfig(r.options.DockerOptions.Resources.MaxIngressBandwidth, r.options.DockerOptions.Resources.MaxEgressBandwidth)), os.ModePerm)
	if err != nil {
		return "", err
	}

	// Start the VM and wait for the SSH server to become available.
	startCommand := r.newCommand(
		"setup.firecracker.start",
		command.Flatten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "cni",
			firecrackerResourceFlags(r.options.DockerOptions.Resources),
			firecrackerCopyfileFlags(r.options.VMStartupScriptPath, daemonConfigFile, dockerConfigPath),
			firecrackerVolumeFlags(r.workspaceDevice),
			"--ssh",
			"--name", r.vmName,
			"--kernel-image", sanitizeImage(r.options.KernelImage),
			"--kernel-args", config.FirecrackerKernelArgs,
			"--sandbox-image", sanitizeImage(r.options.SandboxImage),
			sanitizeImage(r.options.Image),
		),
		[]string{fmt.Sprintf("CNI_CONF_DIR=%s", cniConfigDir)},
	)

	if err = startCommand.Run(ctx); err != nil {
		return "", errors.Wrap(err, "failed to start firecracker vm")
	}

	if r.options.VMStartupScriptPath != "" {
		startupScriptCommand := r.newCommand(
			"setup.startup-script",
			command.Flatten("ignite", "exec", r.vmName, "--", r.options.VMStartupScriptPath),
			nil,
		)
		if err = startupScriptCommand.Run(ctx); err != nil {
			return "", errors.Wrap(err, "failed to run startup script")
		}
	}

	if dockerConfigPath != "" {
		return command.FirecrackerDockerConfDir, nil
	}

	return "", nil
}

func newDockerDaemonConfig(tmpDir string, mirrorAddresses []string) (_ string, err error) {
	c, err := json.Marshal(&dockerDaemonConfig{RegistryMirrors: mirrorAddresses})
	if err != nil {
		return "", errors.Wrap(err, "marshalling docker daemon config")
	}

	tmpFilePath := path.Join(tmpDir, dockerDaemonConfigFilename)
	err = os.WriteFile(tmpFilePath, c, os.ModePerm)
	return tmpFilePath, errors.Wrap(err, "writing docker daemon config file")
}

// dockerDaemonConfig is a struct that marshals into a valid docker daemon config.
type dockerDaemonConfig struct {
	RegistryMirrors []string `json:"registry-mirrors"`
}

// dockerDaemonConfigFilename is the filename in the firecracker state tmp directory
// for the optional docker daemon config file.
const dockerDaemonConfigFilename = "docker-daemon.json"

// cniConfig generates a config file that configures the CNI explicitly and adds
// the isolation plugin to the chain.
// This is used to prevent cross-network communication (which currently doesn't
// happen as we only have 1 bridge).
// We also set the maximum bandwidth usable per VM to the configured value to avoid
// abuse and to make sure multiple VMs on the same host won't starve others.
func cniConfig(maxIngressBandwidth, maxEgressBandwidth int) string {
	return fmt.Sprintf(
		defaultCNIConfig,
		config.CNISubnetCIDR,
		maxIngressBandwidth,
		2*maxIngressBandwidth,
		maxEgressBandwidth,
		2*maxEgressBandwidth,
	)
}

// defaultCNIConfig is the CNI config used for our firecracker VMs.
// TODO: Can we remove the portmap completely?
const defaultCNIConfig = `
{
  "cniVersion": "0.4.0",
  "name": "ignite-cni-bridge",
  "plugins": [
    {
  	  "type": "bridge",
  	  "bridge": "ignite0",
  	  "isGateway": true,
  	  "isDefaultGateway": true,
  	  "promiscMode": false,
  	  "ipMasq": true,
  	  "ipam": {
  	    "type": "host-local",
  	    "subnet": %q
  	  }
    },
    {
  	  "type": "portmap",
  	  "capabilities": {
  	    "portMappings": true
  	  }
    },
    {
  	  "type": "firewall"
    },
    {
  	  "type": "isolation"
    },
    {
  	  "name": "slowdown",
  	  "type": "bandwidth",
  	  "ingressRate": %d,
  	  "ingressBurst": %d,
  	  "egressRate": %d,
  	  "egressBurst": %d
    }
  ]
}
`

func firecrackerResourceFlags(options command.ResourceOptions) []string {
	return []string{
		"--cpus", strconv.Itoa(options.NumCPUs),
		"--memory", options.Memory,
		"--size", options.DiskSpace,
	}
}

func firecrackerCopyfileFlags(vmStartupScriptPath, daemonConfigFile, dockerConfigPath string) []string {
	copyfiles := make([]string, 0, 2)
	if vmStartupScriptPath != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", vmStartupScriptPath, vmStartupScriptPath))
	}

	if daemonConfigFile != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", daemonConfigFile, "/etc/docker/daemon.json"))
	}

	if dockerConfigPath != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", dockerConfigPath, command.FirecrackerDockerConfDir))
	}

	sort.Strings(copyfiles)
	return command.Intersperse("--copy-files", copyfiles)
}

func firecrackerVolumeFlags(workspaceDevice string) []string {
	return []string{"--volumes", fmt.Sprintf("%s:%s", workspaceDevice, command.FirecrackerContainerDir)}
}

// sanitizeImage sanitizes the given docker image for use by ignite. The ignite utility
// has some issue parsing docker tags that include a sha256 hash, so we try to remove it
// from any of the image references before passing it to the ignite command.
func sanitizeImage(image string) string {
	if matches := imagePattern.FindStringSubmatch(image); len(matches) == 4 {
		if matches[2] == "" {
			return matches[1]
		}

		return fmt.Sprintf("%s:%s", matches[1], matches[2])
	}

	return image
}

var imagePattern = lazyregexp.New(`([^:@]+)(?::([^@]+))?(?:@sha256:([a-z0-9]{64}))?`)

func (r *firecrackerRunner) Teardown(ctx context.Context) error {
	return r.teardownFirecracker(ctx)
}

// teardownFirecracker issues a stop and a remove request for the Firecracker VM with
// the given name and removes the tmpDir.
func (r *firecrackerRunner) teardownFirecracker(ctx context.Context) error {
	removeCommand := r.newCommand{
		"teardown.firecracker.remove",
		command.Flatten("ignite", "rm", "-f", r.vmName),
	}
	if err := runner.RunCommand(ctx, removeCommand, logger); err != nil {
		log15.Error("Failed to remove firecracker vm", "name", name, "err", err)
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		log15.Error("Failed to remove firecracker state tmp dir", "name", name, "tmpDir", tmpDir, "err", err)
	}

	return nil
}

func (r *firecrackerRunner) newCommand(key string, cmd []string, env []string) command.Command {
	return command.Command{
		Key:       key,
		Command:   cmd,
		Env:       env,
		CmdRunner: nil,
		Logger:    r.cmdLogger,
	}
}

func (r *firecrackerRunner) Run(ctx context.Context) error {
	firecrackerCommand := command.NewFirecrackerCommand(r.cmdLogger, nil, r.vmName, r.options.DockerOptions)
	return firecrackerCommand.Run(ctx)
}
