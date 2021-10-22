# Single-container Sourcegraph with Docker operations guides

Operations guides specific to managing [single-container Sourcegraph with Docker](./index.md) installations.

Trying to deploy single-container Sourcegraph with Docker? Refer to our [installation guide](./index.md#installation).

## Upgrade

Before upgrading, refer to the [update notes for single-container Sourcegraph with Docker](../../updates/pure_docker.md).

To update, just use the newer `sourcegraph/server:N.N.N` Docker image (where `N.N.N` is the version number) in place of the older one, using the same Docker volumes. Your server's data will be migrated automatically if needed.

You can always find the version number of the latest release at [docs.sourcegraph.com](https://docs.sourcegraph.com) in the `docker run` command's image tag.

- As a precaution, before updating, we recommend backing up the contents of the Docker volumes used by Sourcegraph.
- If you need a HA deployment, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph).
- There is currently no automated way to downgrade to an older version after you have updated. [Contact support](https://about.sourcegraph.com/contact) for help.

## Configure exposed Sourcegraph port

Change the `docker` `--publish` argument to make it listen on the specific interface and port on your host machine. For example, `docker run ... --publish 0.0.0.0:80:7080 ...` would make it accessible on port 80 of your machine. For more information, see "[Publish or expose port](https://docs.docker.com/engine/reference/commandline/run/#publish-or-expose-port--p---expose)" in the Docker documentation.

The other option is to deploy and run Sourcegraph on a cloud provider. For an example, see documentation to [deploy to Google Cloud](./google_cloud.md).

## Access the database

> NOTE: To execute an SQL query against the database without first creating an interactive session (as below), append `--command "SELECT * FROM users;"` to the `docker container exec` command.

Get the Docker container ID for Sourcegraph:

```bash
docker ps
CONTAINER ID        IMAGE
d039ec989761        sourcegraph/server:VERSION
```

Open a PostgreSQL interactive terminal:

```bash
docker container exec -it d039ec989761 psql -U postgres sourcegraph
```

Run your SQL query:

```sql
SELECT * FROM users;
```

## Git configuration and authentication

For single-container environments, upon the Sourcegraph Docker image container start, it copies all files from `/etc/sourcegraph/{ssh,gitconfig,netrc}` into its own `$HOME` directory, via the `--volume /mnt/sourcegraph/config:/etc/sourcegraph` in the `docker run` command.

For example, to mount a `.gitconfig`, create a file `/mnt/sourcegraph/config/gitconfig` on your host containing your configuration:

```
# example .gitconfig

[url "example.url.com:"]
  insteadOf = "ssh://example.url.com"
```

Alternatively you can create a new Docker image which inherits from Sourcegraph and then mutates the environment:

```dockerfile
FROM sourcegraph/server:3.33.0

COPY gitconfig /etc/gitconfig
COPY ssh /root/.ssh
RUN	find /root/.ssh -type f -exec chmod 600 '{}' ';'
RUN	find /root/.ssh -type d -exec chmod 700 '{}' ';'
```

This approach can also be used for `sourcegraph/gitserver` images in cluster environments.

Learn more about Git [configuration](../../repo/git_config.md) and [authentication](../../repo/auth.md).

### SSH authentication (config, keys, `known_hosts`)

The container consults its own file system (in the standard locations) for SSH configuration, private keys, and `known_hosts`. Upon container start, it copies all files from `/etc/sourcegraph/ssh` into its own `$HOME/.ssh` directory.

To provide SSH authentication configuration to the container, assuming you're using the default `--volume $HOME/.sourcegraph/config:/etc/sourcegraph`, follow these steps:

1. Create files at `$HOME/.sourcegraph/config/ssh/config`, `$HOME/.sourcegraph/config/ssh/known_hosts`, etc., on the host machine as desired to configure SSH.
1. Start (or restart) the container.

To configure the container to use the same SSH as your user account on the host machine, you can also run `cp -R $HOME/.ssh $HOME/.sourcegraph/config/ssh`.

### HTTP(S) authentication via netrc

The easiest way to specify HTTP(S) authentication for repositories is to include the username and password in the clone URL itself, such as `https://user:password@example.com/my/repo`. These credentials won't be displayed to non-admin users.

Otherwise, the container consults the `$HOME/.netrc` files on its own file system for HTTP(S) authentication. The `.netrc` file is a standard way to specify authentication used to connect to external hosts.

To provide HTTP(S) authentication, assuming you're using the default `--volume $HOME/.sourcegraph/config:/etc/sourcegraph`, follow these steps:

1. Create a file at `$HOME/.sourcegraph/config/netrc` on the host machine that contains lines of the form `machine example.com login alice password mypassword` (replacing `example.com`, `alice`, and `mypassword` with the actual values).
1. Start (or restart) the container.

## Expose debug port

This is required to [collect debug data](../../pprof.md).

The docker run command for single-container Sourcegraph needs an additional publish flag to expose the debug port:

```bash script
docker run --publish 7080:7080 --publish 127.0.0.1:3370:3370 --publish 127.0.0.1:6060:6060 --rm --volume ~/.sourcegraph/config:/etc/sourcegraph --volume ~/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:3.33.0
```

If Sourcegraph is deployed to a remote server, then access via an SSH tunnel using a tool
such as [sshuttle](https://github.com/sshuttle/sshuttle) is required to establish a secure connection.
To access the remote server using `sshuttle` from your local machine:

```bash script
sshuttle -r user@host 0/0
```

## Environment variables

Add the following to your docker run command:

```
docker run [...]
-e (YOUR CODE)
sourcegraph/server:3.33.0
```
