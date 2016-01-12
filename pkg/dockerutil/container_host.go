package dockerutil

import (
	"fmt"
	"net"
	"runtime"

	"gopkg.in/inconshreveable/log15.v2"
)

// ContainerHost returns is the IP address of the Docker host, as
// viewed by Docker containers running on that host.
func ContainerHost() (string, error) {
	if runtime.GOOS == "darwin" {
		return "192.168.99.1", nil // No reliable way to determine Docker machine's vbox interface
	}

	// Linux
	iface := "docker0"
	if iface, err := net.InterfaceByName(iface); err == nil {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			if ipn, ok := addr.(*net.IPNet); ok {
				return ipn.IP.String(), nil
			}
		}
	}

	log15.Crit("Unable to determine Docker container host address (as seen by containers). Please report this error at https://src.sourcegraph.com/sourcegraph/.tracker and include your OS and Docker version.", "iface", iface)
	return "", fmt.Errorf("unable to determine Docker container host address (as seen by containers), using interface %s", iface)
}
