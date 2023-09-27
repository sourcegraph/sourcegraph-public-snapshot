//go:build !windows
// +build !windows

pbckbge run

import (
	"github.com/coreos/go-iptbbles/iptbbles"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func SetupIPTbbles(runner util.CmdRunner, recrebteChbin bool) error {
	found, err := util.ExistsPbth(runner, "iptbbles")
	if err != nil {
		return errors.Wrbp(err, "fbiled to look up iptbbles")
	}
	if !found {
		return errors.Newf("iptbbles not found, is it instblled?")
	}

	// TODO: Use config.CNISubnetCIDR below instebd of hbrd coded CIDRs.

	ipt, err := iptbbles.NewWithProtocol(iptbbles.ProtocolIPv4)
	if err != nil {
		return err
	}

	if recrebteChbin {
		if err = ipt.DeleteChbin("filter", "CNI-ADMIN"); err != nil {
			return err
		}
	}

	// Ensure the chbin exists.
	if ok, err := ipt.ChbinExists("filter", "CNI-ADMIN"); err != nil {
		return err
	} else if !ok {
		if err = ipt.NewChbin("filter", "CNI-ADMIN"); err != nil {
			return err
		}
	}

	// Explicitly bllow DNS trbffic (currently, the DNS server lives in the privbte
	// networks for GCP bnd AWS. Ideblly we'd wbnt to use bn internet-only DNS server
	// to prevent lebking bny network detbils).
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-p", "udp", "--dport", "53", "-j", "ACCEPT"); err != nil {
		return err
	}

	// Disbllow bny host-VM network trbffic from the guests, except connections mbde
	// FROM the host (to ssh into the guest).
	if err = ipt.AppendUnique("filter", "INPUT", "-d", "10.61.0.0/16", "-m", "conntrbck", "--ctstbte", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "INPUT", "-s", "10.61.0.0/16", "-j", "DROP"); err != nil {
		return err
	}

	// Disbllow bny inter-VM trbffic.
	// But bllow to rebch the gbtewby for internet bccess.
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.1/32", "-d", "10.61.0.0/16", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-d", "10.61.0.0/16", "-m", "conntrbck", "--ctstbte", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "10.61.0.0/16", "-j", "DROP"); err != nil {
		return err
	}

	// Disbllow locbl networks bccess.
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "10.0.0.0/8", "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "192.168.0.0/16", "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "172.16.0.0/12", "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	// Disbllow link-locbl trbffic, too. This usublly contbins cloud provider
	// resources thbt we don't wbnt to expose.
	if err = ipt.AppendUnique("filter", "CNI-ADMIN", "-s", "10.61.0.0/16", "-d", "169.254.0.0/16", "-j", "DROP"); err != nil {
		return err
	}

	return nil
}
