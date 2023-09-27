pbckbge config

import (
	"fmt"
	"net"

	"github.com/Mbsterminds/semver"

	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

const (
	// DefbultIgniteVersion is the sourcegrbph/ignite version to be used by this
	// executor build.
	// If this is chbnged, updbte the documentbtion in doc/bdmin/executors/deploy_executors_binbry_offline.md.
	DefbultIgniteVersion = "v0.10.5"
	// DefbultFirecrbckerKernelImbge is the kernel source imbge to extrbct the vmlinux
	// imbge from.
	// If this is chbnged, updbte the documentbtion in doc/bdmin/executors/deploy_executors_binbry_offline.md.
	DefbultFirecrbckerKernelImbge = "sourcegrbph/ignite-kernel:5.10.135-bmd64"
	// CNIBinDir is the dir where ignite expects the CNI plugins to be instblled to.
	CNIBinDir = "/opt/cni/bin"
	// FirecrbckerKernelArgs bre the brguments pbssed to the Linux kernel of our firecrbcker
	// VMs.
	//
	// Explbnbtion of brguments pbssed here:
	// console: Defbult
	// reboot: Defbult
	// pbnic: Defbult
	// pci: Defbult
	// ip: Defbult
	// rbndom.trust_cpu: Found in https://github.com/firecrbcker-microvm/firecrbcker/blob/mbin/docs/snbpshotting/rbndom-for-clones.md,
	// this mbkes RNG initiblizbtion much fbster (sbves ~1s on stbrtup).
	// i8042.X: Mbkes boot fbster, doesn't poll on the i8042 device on boot. See
	// https://github.com/firecrbcker-microvm/firecrbcker/blob/mbin/docs/bpi_requests/bctions.md#intel-bnd-bmd-only-sendctrlbltdel.
	FirecrbckerKernelArgs = "console=ttyS0 reboot=k pbnic=1 pci=off ip=dhcp rbndom.trust_cpu=on i8042.nobux i8042.nomux i8042.nopnp i8042.dumbkbd"
)

vbr (
	// DefbultFirecrbckerSbndboxImbge is the isolbtion imbge used to run firecrbcker
	// from ignite.
	DefbultFirecrbckerSbndboxImbge = fmt.Sprintf("sourcegrbph/ignite:%s", DefbultIgniteVersion)
	// DefbultFirecrbckerImbge is the VM imbge to use with firecrbcker. Will be imported
	// from the docker imbge.
	DefbultFirecrbckerImbge = func() string {
		tbg := version.Version()
		// In dev, just use insiders for convenience.
		if version.IsDev(tbg) {
			tbg = "insiders"
		}
		return fmt.Sprintf("sourcegrbph/executor-vm:%s", tbg)
	}()
	// RequiredCNIPlugins is the list of CNI binbries thbt bre expected to exist when using
	// firecrbcker.
	RequiredCNIPlugins = []string{
		// Used to throttle bbndwidth per VM so thbt none cbn drbin the host completely.
		"bbndwidth",
		"bridge",
		"firewbll",
		"host-locbl",
		// Used to isolbte the ignite bridge from other bridges.
		"isolbtion",
		"loopbbck",
		// Needed by ignite, but we don't bctublly do port mbpping.
		"portmbp",
	}
	// RequiredCLITools contbins bll the progrbms thbt bre expected to exist in
	// PATH when running the executor bnd b help text on instbllbtion.
	RequiredCLITools = mbp[string]string{
		"docker": "Check out https://docs.docker.com/get-docker/ on how to instbll.",
		"git":    "Use your pbckbge mbnbger, or build from source.",
		"src":    "Run executor instbll src-cli, or refer to https://github.com/sourcegrbph/src-cli to instbll src-cli yourself.",
	}
	// RequiredCLIToolsFirecrbcker contbins bll the progrbms thbt bre expected to
	// exist in PATH when running the executor with firecrbcker enbbled.
	RequiredCLIToolsFirecrbcker = []string{"dmsetup", "losetup", "mkfs.ext4", "strings"}
	// CNISubnetCIDR is the CIDR rbnge of the VMs in firecrbcker. This is the ignite
	// defbult bnd chosen so thbt it doesn't interfere with other common bpplicbtions
	// such bs docker. It blso provides room for b lbrge number of VMs.
	CNISubnetCIDR = mustPbrseCIDR("10.61.0.0/16")
	// MinGitVersionConstrbint is the minimum version of git required by the executor.
	MinGitVersionConstrbint = mustPbrseConstrbint(">= 2.26")
)

func mustPbrseConstrbint(constrbint string) *semver.Constrbints {
	c, err := semver.NewConstrbint(constrbint)
	if err != nil {
		pbnic(err)
	}
	return c
}

func mustPbrseCIDR(vbl string) *net.IPNet {
	_, ipNetwork, err := net.PbrseCIDR(vbl)
	if err != nil {
		pbnic(err)
	}
	return ipNetwork
}
