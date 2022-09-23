package command

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"

	shellquote "github.com/kballard/go-shellquote"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type commandRunner interface {
	RunCommand(ctx context.Context, command command, logger Logger) error
}

const firecrackerContainerDir = "/work"

// cniSubnetCIDR is the CIDR range of the VMs in firecracker. This is the ignite
// default and chosen so that it doesn't interfere with other common applications
// such as docker. It also provides room for a large number of VMs.
var cniSubnetCIDR = mustParseCIDR("10.61.0.0/16")

// formatFirecrackerCommand constructs the command to run on the host via a Firecracker
// virtual machine in order to invoke the given spec. If the spec specifies an image, then
// the command will be run inside of a container inside of the VM. Otherwise, the command
// will be run inside of the VM. The containers are one-shot and subject to the resource
// limits specified in the given options.
//
// The name value supplied here refers to the Firecracker virtual machine, which must have
// also been the name supplied to a successful invocation of setupFirecracker. Additionally,
// the virtual machine must not yet have been torn down (via teardownFirecracker).
func formatFirecrackerCommand(spec CommandSpec, name string, options Options) command {
	rawOrDockerCommand := formatRawOrDockerCommand(spec, firecrackerContainerDir, options)

	innerCommand := shellquote.Join(rawOrDockerCommand.Command...)
	if len(rawOrDockerCommand.Env) > 0 {
		// If we have env vars that are arguments to the command we need to escape them
		quotedEnv := quoteEnv(rawOrDockerCommand.Env)
		innerCommand = fmt.Sprintf("%s %s", strings.Join(quotedEnv, " "), innerCommand)
	}
	if rawOrDockerCommand.Dir != "" {
		innerCommand = fmt.Sprintf("cd %s && %s", rawOrDockerCommand.Dir, innerCommand)
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

const defaultCNIConfigPath = "/etc/cni/net.d/10-ignite.conflist"

// writeDefaultCNIConfig writes the dfault settings for the CNI to disk. This is
// useful so that it's easy to spin up a VM without using the executor for debugging.
// TODO: Use this somewhere.
// TODO: What if we added a command "executor run [NAME]" that just starts a VM
// with all the defaults, like ignite run but doesn't require all these global
// config files.
func writeDefaultCNIConfig() error {
	// Make sure the directory exists.
	if err := os.MkdirAll(path.Dir(defaultCNIConfigPath), os.ModePerm); err != nil {
		return nil
	}
	// Check if the config already exists. If so: quit.
	if _, err := os.Stat(defaultCNIConfigPath); err != nil && !os.IsNotExist(err) {
		return err
	} else if os.IsExist(err) {
		return nil
	}
	// Write the default config file.
	f, err := os.Create(defaultCNIConfigPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return os.WriteFile(f.Name(), []byte(cniConfig(500*1024*1024, 500*1024*1024)), os.ModePerm)
}

// Configures the CNI explicitly and adds the isolation plugin to the chain.
// This is to prevent cross-network communication (which currently doesn't happen
// as we only have 1 bridge).
// We also set the maximum bandwidth usable per VM to the configured value to avoid
// abuse and to make sure multiple VMs on the same host won't starve others.
func cniConfig(maxIngressBandwidth, maxEgressBandwidth int) string {
	return fmt.Sprintf(
		defaultCNIConfig,
		cniSubnetCIDR,
		maxIngressBandwidth,
		2*maxIngressBandwidth,
		maxEgressBandwidth,
		2*maxEgressBandwidth,
	)
}

// firecrackerKernelArgs are the arguments passed to the Linux kernel of our firecracker
// VMs.
//
// Explanation of arguments passed here:
// console: Default
// reboot: Default
// panic: Default
// pci: Default
// ip: Default
// random.trust_cpu: Found in https://github.com/firecracker-microvm/firecracker/blob/main/docs/snapshotting/random-for-clones.md,
// this makes RNG initialization much faster (saves ~1s on startup).
// i8042.X: Makes boot faster, doesn't poll on the i8042 device on boot. See
// https://github.com/firecracker-microvm/firecracker/blob/main/docs/api_requests/actions.md#intel-and-amd-only-sendctrlaltdel.
const firecrackerKernelArgs = "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp random.trust_cpu=on i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd"

// dockerDaemonConfig is a struct that marshals into a valid docker daemon config.
type dockerDaemonConfig struct {
	RegistryMirrors []string `json:"registry-mirrors"`
}

// dockerDaemonConfigFilename is the filename in the firecracker state tmp directory
// for the optional docker daemon config file.
const dockerDaemonConfigFilename = "docker-daemon.json"

func newDockerDaemonConfig(tmpDir, mirrorAddress string) (_ string, err error) {
	c, err := json.Marshal(&dockerDaemonConfig{RegistryMirrors: []string{mirrorAddress}})
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
func setupFirecracker(ctx context.Context, runner commandRunner, logger Logger, name, workspaceDevice string, options Options, operations *Operations) error {
	tmpDir, err := os.MkdirTemp("", "firecracker-vm-state")
	if err != nil {
		return err
	}
	var daemonConfigFile string
	if options.FirecrackerOptions.DockerRegistryMirrorURL != "" {
		var err error
		daemonConfigFile, err = newDockerDaemonConfig(tmpDir, options.FirecrackerOptions.DockerRegistryMirrorURL)
		if err != nil {
			return err
		}
	}

	cniConfigDir := path.Join(tmpDir, "cni")
	err = os.Mkdir(cniConfigDir, os.ModePerm)
	if err != nil {
		return err
	}
	cniConfigFile := path.Join(tmpDir, "10-sourcegraph-executors.conflist")
	err = os.WriteFile(cniConfigFile, []byte(cniConfig(options.ResourceOptions.MaxIngressBandwidth, options.ResourceOptions.MaxEgressBandwidth)), os.ModePerm)
	if err != nil {
		return err
	}

	// Start the VM and wait for the SSH server to become available.
	startCommand := command{
		Key: "setup.firecracker.start",
		// Tell ignite to use our temporary config.
		// TODO: This requires a new ignite release.
		Env: []string{fmt.Sprintf("CNI_CONF_DIR=%s", tmpDir)},
		Command: flatten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "cni",
			firecrackerResourceFlags(options.ResourceOptions),
			firecrackerCopyfileFlags(options.FirecrackerOptions.VMStartupScriptPath, daemonConfigFile),
			firecrackerVolumeFlags(workspaceDevice, firecrackerContainerDir),
			"--ssh",
			"--name", name,
			"--kernel-image", sanitizeImage(options.FirecrackerOptions.KernelImage),
			"--kernel-args", firecrackerKernelArgs,
			"--sandbox-image", options.FirecrackerOptions.SandboxImage,
			sanitizeImage(options.FirecrackerOptions.Image),
		),
		Operation: operations.SetupFirecrackerStart,
	}

	if err := runner.RunCommand(ctx, startCommand, logger); err != nil {
		return errors.Wrap(err, "failed to start firecracker vm")
	}

	if options.FirecrackerOptions.VMStartupScriptPath != "" {
		startupScriptCommand := command{
			Key:       "setup.startup-script",
			Command:   flatten("ignite", "exec", name, "--", options.FirecrackerOptions.VMStartupScriptPath),
			Operation: operations.SetupStartupScript,
		}
		if err := runner.RunCommand(ctx, startupScriptCommand, logger); err != nil {
			return errors.Wrap(err, "failed to run startup script")
		}
	}

	return nil
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

func firecrackerCopyfileFlags(vmStartupScriptPath, daemonConfigFile string) []string {
	copyfiles := make([]string, 0, 2)
	if vmStartupScriptPath != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", vmStartupScriptPath, vmStartupScriptPath))
	}

	if daemonConfigFile != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", daemonConfigFile, "/etc/docker/daemon.json"))
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

func mustParseCIDR(val string) *net.IPNet {
	_, net, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}
	return net
}
