package docker

import "github.com/samalba/dockerclient"

// Client is a wrapper around the default Docker client
// that tracks all created containers ensures some default
// configurations are in place.
type Client struct {
	dockerclient.Client
	info  *dockerclient.ContainerInfo
	names []string // names of created containers
}

func NewClient(docker dockerclient.Client) (*Client, error) {
	// creates an ambassador container
	conf := &dockerclient.ContainerConfig{}
	conf.HostConfig = dockerclient.HostConfig{
		MemorySwappiness: -1,
	}
	conf.Entrypoint = []string{"/bin/sleep"}
	conf.Cmd = []string{"86400"}
	conf.Image = "gliderlabs/alpine:3.1"
	conf.Volumes = map[string]struct{}{}
	conf.Volumes["/drone"] = struct{}{}
	info, err := Start(docker, conf, nil, false)
	if err != nil {
		return nil, err
	}

	return &Client{Client: docker, info: info}, nil
}

// CreateContainer creates a container and internally
// caches its container id.
func (c *Client) CreateContainer(conf *dockerclient.ContainerConfig, name string, auth *dockerclient.AuthConfig) (string, error) {
	conf.Env = append(conf.Env, "affinity:container=="+c.info.Id)
	id, err := c.Client.CreateContainer(conf, name, auth)
	if err == nil {
		c.names = append(c.names, id)
	}
	return id, err
}

// StartContainer starts a container and links to an
// ambassador container sharing the build machiens volume.
func (c *Client) StartContainer(id string, conf *dockerclient.HostConfig) error {
	conf.VolumesFrom = append(conf.VolumesFrom, c.info.Id)
	if len(conf.NetworkMode) == 0 {
		conf.NetworkMode = "container:" + c.info.Id
	}
	return c.Client.StartContainer(id, conf)
}

// Destroy will terminate and destroy all containers that
// were created by this client.
func (c *Client) Destroy() error {
	for _, id := range c.names {
		c.Client.KillContainer(id, "9")
		c.Client.RemoveContainer(id, true, true)
	}
	c.Client.KillContainer(c.info.Id, "9")
	return c.Client.RemoveContainer(c.info.Id, true, true)
}
