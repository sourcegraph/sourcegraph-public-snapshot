# Quickstart step 3: Start Docker

## macOS

### Option A: Docker for Mac

This is the easy way - just launch Docker.app and wait for it to finish loading.

### Option B: docker-machine

The Docker daemon should be running in the background, which you can test by
running `docker ps`. If you're on OS X and using `docker-machine` instead of
Docker for Mac, you may have to run:

```bash
docker-machine start default
eval $(docker-machine env)
```

## Ubuntu/Linux

The docker daemon might already be running, but if necessary you can use the following commands to start it:

```sh
# as a system service
sudo systemctl enable --now docker

# manually
dockerd
```

If you have issues running Docker, try [adding your user to the docker group][dockerGroup], and/or [updating the socket file permissions][socketPermissions], or try running these commands under `sudo`.

[dockerGroup]: https://stackoverflow.com/a/48957722
[socketPermissions]: https://stackoverflow.com/a/51362528

[< Previous](quickstart_2_clone_repository.md) | [Next >](quickstart_4_initialize_database.md)
