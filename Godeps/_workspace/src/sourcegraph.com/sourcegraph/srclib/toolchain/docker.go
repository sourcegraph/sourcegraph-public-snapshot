package toolchain

import (
	"os"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

// newDockerClient creates a new Docker client configured to reach Docker at the
// DOCKER_HOST env var, or the default /var/run/docker.sock socket if unset.
func newDockerClient() (*docker.Client, error) {
	dockerEndpoint := os.Getenv("DOCKER_HOST")
	if dockerEndpoint == "" {
		dockerEndpoint = "unix:///var/run/docker.sock"
	} else if !strings.HasPrefix(dockerEndpoint, "http") && !strings.HasPrefix(dockerEndpoint, "tcp") {
		dockerEndpoint = "http://" + dockerEndpoint
	}
	return docker.NewClient(dockerEndpoint)
}
