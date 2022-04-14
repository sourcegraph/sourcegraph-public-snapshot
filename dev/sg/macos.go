package main

import "github.com/urfave/cli/v2"

var (
	addToMacOSFirewall     bool
	addToMacOSFirewallFlag = &cli.BoolFlag{
		Name:        "add-to-macos-firewall",
		Usage:       "OSX only; Add required exceptions to the firewall",
		Value:       true,
		Destination: &addToMacOSFirewall,
	}
)
