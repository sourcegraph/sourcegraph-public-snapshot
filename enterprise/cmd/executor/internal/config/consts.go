package config

import (
	"fmt"
	"net"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

const (
	// DefaultIgniteVersion is the sourcegraph/ignite version to be used by this
	// executor build.
	// If this is changed, update the documentation in doc/admin/executors/deploy_executors_binary_offline.md.
	DefaultIgniteVersion = "v0.10.5"
	// DefaultFirecrackerKernelImage is the kernel source image to extract the vmlinux
	// image from.
	// If this is changed, update the documentation in doc/admin/executors/deploy_executors_binary_offline.md.
	DefaultFirecrackerKernelImage = "sourcegraph/ignite-kernel:5.10.135-amd64"
	// CNIBinDir is the dir where ignite expects the CNI plugins to be installed to.
	CNIBinDir = "/opt/cni/bin"
	// FirecrackerKernelArgs are the arguments passed to the Linux kernel of our firecracker
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
	FirecrackerKernelArgs = "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp random.trust_cpu=on i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd"
)

var (
	// DefaultFirecrackerSandboxImage is the isolation image used to run firecracker
	// from ignite.
	DefaultFirecrackerSandboxImage = fmt.Sprintf("sourcegraph/ignite:%s", DefaultIgniteVersion)
	// DefaultFirecrackerImage is the VM image to use with firecracker. Will be imported
	// from the docker image.
	DefaultFirecrackerImage = func() string {
		tag := version.Version()
		// In dev, just use insiders for convenience.
		if version.IsDev(tag) {
			tag = "insiders"
		}
		return fmt.Sprintf("sourcegraph/executor-vm:%s", tag)
	}()
	// RequiredCNIPlugins is the list of CNI binaries that are expected to exist when using
	// firecracker.
	RequiredCNIPlugins = []string{
		// Used to throttle bandwidth per VM so that none can drain the host completely.
		"bandwidth",
		"bridge",
		"firewall",
		"host-local",
		// Used to isolate the ignite bridge from other bridges.
		"isolation",
		"loopback",
		// Needed by ignite, but we don't actually do port mapping.
		"portmap",
	}
	// RequiredCLITools contains all the programs that are expected to exist in
	// PATH when running the executor and a help text on installation.
	RequiredCLITools = map[string]string{
		"docker": "Check out https://docs.docker.com/get-docker/ on how to install.",
		"git":    "Use your package manager, or build from source.",
		"src":    "Run executor install src-cli, or refer to https://github.com/sourcegraph/src-cli to install src-cli yourself.",
	}
	// RequiredCLIToolsFirecracker contains all the programs that are expected to
	// exist in PATH when running the executor with firecracker enabled.
	RequiredCLIToolsFirecracker = []string{"dmsetup", "losetup", "mkfs.ext4", "strings"}
	// CNISubnetCIDR is the CIDR range of the VMs in firecracker. This is the ignite
	// default and chosen so that it doesn't interfere with other common applications
	// such as docker. It also provides room for a large number of VMs.
	CNISubnetCIDR = mustParseCIDR("10.61.0.0/16")
	// MinGitVersionConstraint is the minimum version of git required by the executor.
	MinGitVersionConstraint = mustParseConstraint(">= 2.26")
)

func mustParseConstraint(constraint string) *semver.Constraints {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return c
}

func mustParseCIDR(val string) *net.IPNet {
	_, ipNetwork, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}
	return ipNetwork
}
