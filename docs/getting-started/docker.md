+++
title = "Deploying with Docker"
linktitle = "with Docker"
+++

Sourcegraph is available as the
[sourcegraph/sourcegraph image on Docker Hub](https://hub.docker.com/r/sourcegraph/sourcegraph/).

# Running a container

To run Sourcegraph in a Docker container:

```
docker run \
  --name src \
  --detach \
  --publish 80:80 \
  --publish 443:443 \
  --restart on-failure:10 \
  --volume /var/lib/sourcegraph:/root/.sourcegraph \
  --volume /var/run/docker.sock:/var/run/docker.sock \
  sourcegraph/sourcegraph:latest \
  serve \
  --http-addr=:80 \
  --https-addr=:443
```

This will download and run Sourcegraph. Once started, you can access
your Sourcegraph server on the host's HTTP port (80).

Note that Sourcegraph requires the ability to run Docker
containers. If you're unable to mount `/var/run/docker.sock` (as in
the command above), you may pass the Docker host address with (for
example) `--env DOCKER_HOST=tcp://1.2.3.4:2376`.


## Storage

Sourcegraph's data (repositories, builds, users, etc.) is persisted on
the host using a
[Docker volume](https://docs.docker.com/userguide/dockervolumes/). The
command above uses the `/var/lib/sourcegraph` directory on the host to
store this data.


## Configuration & administration

* Edit configuration: modify the command-line flags after the `serve` command.
* Restart the Sourcegraph server (required after config changes): run
  `docker restart src`
* Stop the Sourcegraph server: run `docker stop src`
* Upgrade: run `docker exec -it src selfupdate` then restart the server
* View logs: run `docker logs src`
* Access a shell prompt in the container: run `docker exec -it src
  /bin/bash`


# Advanced

## Rebuilding the Docker image

The Docker image is built from the
[top-level `Dockerfile` in the Sourcegraph repository](https://src.sourcegraph.com/sourcegraph/.tree/Dockerfile):

```
docker build -t sourcegraph/sourcegraph:latest .
```

{{< ads_conversion >}}
