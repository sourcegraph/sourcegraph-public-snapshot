package dockerutil

import (
	"fmt"
	"net"
	"runtime"

	"gopkg.in/inconshreveable/log15.v2"
)

// ContainerHost returns is the IP address of the Docker host, as
// viewed by Docker containers running on that host.
var ContainerHost = func() (string, error) {
	iface := "docker0"
	if runtime.GOOS == "darwin" {
		iface = "en0"
	}
	if iface, err := net.InterfaceByName(iface); err == nil {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			if ipn, ok := addr.(*net.IPNet); ok {
				if ip := ipn.IP.To4(); ip != nil {
					return ip.String(), nil
				}
			}
		}
	}

	log15.Crit("Unable to determine Docker container host address (as seen by containers). Please report this error to support@sourcegraph.com and include your OS and Docker version.", "iface", iface)
	return "", fmt.Errorf("unable to determine Docker container host address (as seen by containers), using interface %s", iface)
}
