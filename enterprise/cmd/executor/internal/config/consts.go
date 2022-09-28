package config

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

const (
	DefaultIgniteVersion          = "v0.10.4"
	DefaultFirecrackerKernelImage = "sourcegraph/ignite-kernel:5.10.135-amd64"
)

var DefaultFirecrackerSandboxImage = fmt.Sprintf("sourcegraph/ignite:%s", DefaultIgniteVersion)

var DefaultFirecrackerImage = func() string {
	tag := version.Version()
	// In dev, just use insiders for convenience.
	if version.IsDev(tag) {
		tag = "insiders"
	}
	return fmt.Sprintf("sourcegraph/executor-vm:%s", tag)
}()

const CNIBinDir = "/opt/cni/bin"

// RequiredCNIPlugins is the list of CNI binaries that are expected to exist when using
// firecracker.
var RequiredCNIPlugins = []string{
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

var (
	RequiredCLITools            = []string{"docker", "git", "src"}
	RequiredCLIToolsFirecracker = []string{"dmsetup", "losetup", "mkfs.ext4"}
)
