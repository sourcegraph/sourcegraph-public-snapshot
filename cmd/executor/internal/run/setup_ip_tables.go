//go:build !windows
// +build !windows

package run

import (
	"github.com/coreos/go-iptables/iptables"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func SetupIPTables(runner util.CmdRunner, recreateChain bool) error {
	found, err := util.ExistsPath(runner, "iptables")
	if err != nil {
		return errors.Wrap(err, "failed to look up iptables")
	}
	if !found {
		return errors.Newf("iptables not found, is it installed?")
	}

	// TODO: Use config.CNISubnetCIDR below instead of hard coded CIDRs.

	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return err
	}

	if recreateChain {
		if err = ipt.DeleteChain("filter", "CNI-ADMIN"); err != nil {
			return err
		}
	}

	// Ensure the chain exists.
	if ok, err := ipt.ChainExists("filter", "CNI-ADMIN"); err != nil {
		return err
	} else if !ok {
		if err = ipt.NewChain("filter", "CNI-ADMIN"); err != nil {
			return err
		}
	}

	// Explicitly allow DNS traffic (currently, the DNS server lives in the private
	// networks for GCP and AWS. Ideally we'd want to use an internet-only DNS server
	// to prevent leaking any network details).
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-p", "udp", "--dport", "53", "-j", "ACCEPT"); err != nil {
		return err
	}

	// Disallow any host-VM network traffic from the guests, except connections made
	// FROM the host (to ssh into the guest).
	if err = ipt.AppendUnique("filter", "INPUT", "-d", "10.61.0.0/16", "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "INPUT", "-s", "10.61.0.0/16", "-j", "DROP"); err != nil {
		return err
	}

	// Disallow any inter-VM traffic.
	// But allow to reach the gateway for internet access.
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.1/32", "-d", "10.61.0.0/16", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-d", "10.61.0.0/16", "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "10.61.0.0/16", "-j", "DROP"); err != nil {
		return err
	}

	// Disallow local networks access.
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "10.0.0.0/8", "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "192.168.0.0/16", "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "172.16.0.0/12", "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	// Disallow link-local traffic, too. This usually contains cloud provider
	// resources that we don't want to expose.
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "169.254.0.0/16", "-j", "DROP"); err != nil {
		return err
	}

	return nil
}
