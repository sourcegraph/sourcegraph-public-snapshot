package toolchain

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

// newTLSDockerClient creates a new TLS docker client using the certificates
// from the the given docker certificate path (i.e. $DOCKER_CERT_PATH)
// connecting to the given docker host (i.e. $DOCKER_HOST).
//
// This should only be used when $DOCKER_CERT_PATH is set and is not an empty
// string.
//
func newTLSDockerClient(certPath, host string) (*docker.Client, error) {
	h, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	h.Scheme = "tcp"

	// Create docker client.
	return docker.NewTLSClient(h.String(), 
		filepath.Join(certPath, "cert.pem"), 
		filepath.Join(certPath, "key.pem"),
		filepath.Join(certPath, "ca.pem"))
}

// newDockerClient creates a new Docker client configured to reach Docker at the
// DOCKER_HOST env var, or the default /var/run/docker.sock socket if unset.
func newDockerClient() (*docker.Client, error) {
	dockerEndpoint := os.Getenv("DOCKER_HOST")
	if strings.HasPrefix(dockerEndpoint, "tcp") {
		if certPath := os.Getenv("DOCKER_CERT_PATH"); certPath != "" {
			return newTLSDockerClient(certPath, dockerEndpoint)
		}
	} else if dockerEndpoint == "" {
		dockerEndpoint = "unix:///var/run/docker.sock"
	} else if !strings.HasPrefix(dockerEndpoint, "http") {
		dockerEndpoint = "http://" + dockerEndpoint
	}
	return docker.NewClient(dockerEndpoint)
}
