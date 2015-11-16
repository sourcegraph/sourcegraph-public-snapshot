+++
title = "Deploying with Docker"
navtitle = "with Docker"
+++

Sourcegraph is available as the
[sourcegraph/sourcegraph image on Docker Hub](https://hub.docker.com/r/sourcegraph/sourcegraph/).

# Running a container

To run Sourcegraph in a Docker container:

```
docker run \
  --name src \
  --detach \
  --hostname src.example.com \
  --publish 80:3080 \
  --publish 443:3443 \
  --restart on-failure:10 \
  --volume /etc/sourcegraph:/etc/sourcegraph \
  --volume /var/lib/sourcegraph:/home/sourcegraph/.sourcegraph \
  sourcegraph/sourcegraph:latest
```

This will download and run Sourcegraph. Once started, you can access
your Sourcegraph server on the host's HTTP port (80).

NOTE: In order to register your Sourcegraph server, you need to set
the AppURL to its externally accessible URL. See below for how to edit
the server's configuration.


## Storage

Sourcegraph's configuration and data is persisted on the host using
[Docker volumes](https://docs.docker.com/userguide/dockervolumes/):

* Configuration: the host directory `/etc/sourcegraph` is mounted at
  `/etc/sourcegraph` in the container.
* Data (repositories, builds, users, etc.): the host directory
  `/var/lib/sourcegraph` is mounted at
  `/home/sourcegraph/.sourcegraph` in the container.


## Configuration & administration

* Edit configuration: run `docker exec -it src vi
  /etc/sourcegraph/config.ini` or edit `/etc/sourcegraph/config.ini`
  directly on the host (assuming you used the volume mapping suggested
  in the `docker run` command above).
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


# Known issues

* The Sourcegraph server version is not set correctly in the Docker
  image. As a result, the Web app erroneously displays a message in
  the footer about upgrading.
