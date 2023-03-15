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
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/kballard/go-shellquote"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type commandRunner interface {
	RunCommand(ctx context.Context, command command, logger Logger) error
}

const (
	firecrackerContainerDir  = "/work"
	firecrackerDockerConfDir = "/etc/docker/cli"
)

// formatFirecrackerCommand constructs the command to run on the host via a Firecracker
// virtual machine in order to invoke the given spec. If the spec specifies an image, then
// the command will be run inside of a container inside of the VM. Otherwise, the command
// will be run inside of the VM. The containers are one-shot and subject to the resource
// limits specified in the given options.
//
// The name value supplied here refers to the Firecracker virtual machine, which must have
// also been the name supplied to a successful invocation of setupFirecracker. Additionally,
// the virtual machine must not yet have been torn down (via teardownFirecracker).
func formatFirecrackerCommand(spec CommandSpec, name string, options Options, dockerConfigPath string) command {
	rawOrDockerCommand := formatRawOrDockerCommand(spec, firecrackerContainerDir, options, dockerConfigPath)

	innerCommand := shellquote.Join(rawOrDockerCommand.Command...)

	// Note: src-cli run commands don't receive env vars in firecracker so we
	// have to prepend them inline to the script.
	// TODO: This branch should disappear when we make src-cli a non-special cased
	// thing.
	if spec.Image == "" && len(rawOrDockerCommand.Env) > 0 {
		innerCommand = fmt.Sprintf("%s %s", strings.Join(quoteEnv(rawOrDockerCommand.Env), " "), innerCommand)
	}

	if rawOrDockerCommand.Dir != "" {
		innerCommand = fmt.Sprintf("cd %s && %s", shellquote.Join(rawOrDockerCommand.Dir), innerCommand)
	}

	return command{
		Key:       spec.Key,
		Command:   []string{"ignite", "exec", name, "--", innerCommand},
		Operation: spec.Operation,
	}
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

// dockerDaemonConfig is a struct that marshals into a valid docker daemon config.
type dockerDaemonConfig struct {
	RegistryMirrors []string `json:"registry-mirrors"`
}

// dockerDaemonConfigFilename is the filename in the firecracker state tmp directory
// for the optional docker daemon config file.
const dockerDaemonConfigFilename = "docker-daemon.json"

func newDockerDaemonConfig(tmpDir string, mirrorAddresses []string) (_ string, err error) {
	c, err := json.Marshal(&dockerDaemonConfig{RegistryMirrors: mirrorAddresses})
	if err != nil {
		return "", errors.Wrap(err, "marshalling docker daemon config")
	}

	tmpFilePath := path.Join(tmpDir, dockerDaemonConfigFilename)
	err = os.WriteFile(tmpFilePath, c, os.ModePerm)
	return tmpFilePath, errors.Wrap(err, "writing docker daemon config file")
}

// setupFirecracker invokes a set of commands to provision and prepare a Firecracker virtual
// machine instance. If a startup script path (an executable file on the host) is supplied,
// it will be mounted into the new virtual machine instance and executed.
func setupFirecracker(ctx context.Context, runner commandRunner, logger Logger, name, workspaceDevice, tmpDir string, options Options, operations *Operations) (_ string, err error) {
	var daemonConfigFile string
	if len(options.FirecrackerOptions.DockerRegistryMirrorURLs) > 0 {
		var err error
		daemonConfigFile, err = newDockerDaemonConfig(tmpDir, options.FirecrackerOptions.DockerRegistryMirrorURLs)
		if err != nil {
			return "", err
		}
	}

	// If docker auth config is present, write it.
	var dockerConfigPath string
	if len(options.DockerOptions.DockerAuthConfig.Auths) > 0 {
		d, err := json.Marshal(options.DockerOptions.DockerAuthConfig)
		if err != nil {
			return "", err
		}
		dockerConfigPath, err = os.MkdirTemp(tmpDir, "docker_auth")
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(dockerConfigPath, "config.json"), d, os.ModePerm); err != nil {
			return "", err
		}
	}

	// Make subdirectory called "cni" to store CNI config in. All files from a directory
	// will be considered so this has to be it's own directory with just our config file.
	cniConfigDir := path.Join(tmpDir, "cni")
	err = os.Mkdir(cniConfigDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	cniConfigFile := path.Join(cniConfigDir, "10-sourcegraph-executors.conflist")
	err = os.WriteFile(cniConfigFile, []byte(cniConfig(options.ResourceOptions.MaxIngressBandwidth, options.ResourceOptions.MaxEgressBandwidth)), os.ModePerm)
	if err != nil {
		return "", err
	}

	// Start the VM and wait for the SSH server to become available.
	startCommand := command{
		Key: "setup.firecracker.start",
		// Tell ignite to use our temporary config file for maximum isolation of
		// envs.
		Env: []string{fmt.Sprintf("CNI_CONF_DIR=%s", cniConfigDir)},
		Command: flatten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "cni",
			firecrackerResourceFlags(options.ResourceOptions),
			firecrackerCopyfileFlags(options.FirecrackerOptions.VMStartupScriptPath, daemonConfigFile, dockerConfigPath),
			firecrackerVolumeFlags(workspaceDevice, firecrackerContainerDir),
			"--ssh",
			"--name", name,
			"--kernel-image", sanitizeImage(options.FirecrackerOptions.KernelImage),
			"--kernel-args", config.FirecrackerKernelArgs,
			"--sandbox-image", sanitizeImage(options.FirecrackerOptions.SandboxImage),
			sanitizeImage(options.FirecrackerOptions.Image),
		),
		Operation: operations.SetupFirecrackerStart,
	}

	if err := runner.RunCommand(ctx, startCommand, logger); err != nil {
		return "", errors.Wrap(err, "failed to start firecracker vm")
	}

	if options.FirecrackerOptions.VMStartupScriptPath != "" {
		startupScriptCommand := command{
			Key:       "setup.startup-script",
			Command:   flatten("ignite", "exec", name, "--", options.FirecrackerOptions.VMStartupScriptPath),
			Operation: operations.SetupStartupScript,
		}
		if err := runner.RunCommand(ctx, startupScriptCommand, logger); err != nil {
			return "", errors.Wrap(err, "failed to run startup script")
		}
	}

	if dockerConfigPath != "" {
		return firecrackerDockerConfDir, nil
	}

	return "", nil
}

// teardownFirecracker issues a stop and a remove request for the Firecracker VM with
// the given name and removes the tmpDir.
func teardownFirecracker(ctx context.Context, runner commandRunner, logger Logger, name, tmpDir string, operations *Operations) error {
	removeCommand := command{
		Key:       "teardown.firecracker.remove",
		Command:   flatten("ignite", "rm", "-f", name),
		Operation: operations.TeardownFirecrackerRemove,
	}
	if err := runner.RunCommand(ctx, removeCommand, logger); err != nil {
		log15.Error("Failed to remove firecracker vm", "name", name, "err", err)
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		log15.Error("Failed to remove firecracker state tmp dir", "name", name, "tmpDir", tmpDir, "err", err)
	}

	return nil
}

func firecrackerResourceFlags(options ResourceOptions) []string {
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
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", dockerConfigPath, firecrackerDockerConfDir))
	}

	sort.Strings(copyfiles)
	return intersperse("--copy-files", copyfiles)
}

func firecrackerVolumeFlags(workspaceDevice, firecrackerContainerDir string) []string {
	return []string{"--volumes", fmt.Sprintf("%s:%s", workspaceDevice, firecrackerContainerDir)}
}

var imagePattern = lazyregexp.New(`([^:@]+)(?::([^@]+))?(?:@sha256:([a-z0-9]{64}))?`)

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
