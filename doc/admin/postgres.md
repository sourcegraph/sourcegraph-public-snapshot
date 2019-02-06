# Postgres

Sourcegraph uses Postgres as its main internal database. We support any version **starting from 10.6**.

## Upgrades

This section describes the possible Postgres upgrade procedures for the different Sourcegraph deployment types.

### Automatic (recommended)

A few different deployment types can automatically handle Postgres data upgrades for you.

#### `sourcegraph/server` automatic upgrades

Existing Postgres data will be automatically migrated when a release of the `sourcegraph/server` Docker image ships with a new version of Postgres.

For the upgrade to proceed, the Docker socket must be mounted the first time the new Docker image is ran.

This is needed to run the [Postgres upgrade containers](https://github.com/tianon/docker-postgres-upgrade) in the
Docker host. When the upgrade is done, the container can be restarted without mounting the Docker socket.

Here's the full invocation when using Docker:

```console
# Add "--env=SRC_LOG_LEVEL=dbug" below for verbose logging.
docker run -p 7080:7080 -p 2633:2633 --rm \
  -v ~/.sourcegraph/config:/etc/sourcegraph \
  -v ~/.sourcegraph/data:/var/opt/sourcegraph \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  sourcegraph/server:3.0.1
```

When using the `sourcegraph/server` image in other environments (e.g. Kubernetes), please refer to official documentation on how to mount the Docker socket for the upgrade procedure.

Alternatively, Postgres can be [upgraded manually](#manual).

#### Kubernetes with https://github.com/sourcegraph/deploy-sourcegraph

### Manual

These instructions can be followed when manual Postgres upgrades are preferred.

#### `sourcegraph/server` manual upgrades

Assuming Postgres data must be upgraded from `9.6` to `11`, here's how it'd be done:

```bash
#!/usr/bin/env bash

set -xeuo pipefail

export OLD=9.6 NEW=11 SRC_DIR="$HOME/.sourcegraph"

docker run \
  -w /tmp/upgrade \
  -v "$SRC_DIR/data/postgres-$NEW-upgrade:/tmp/upgrade" \
  -v "$SRC_DIR/data/postgresql:/var/lib/postgresql/$OLD/data" \
  -v "$SRC_DIR/data/postgresql-$NEW:/var/lib/postgresql/$NEW/data" \
  "tianon/postgres-upgrade:$OLD-to-$NEW"

mv "$SRC_DIR/data/"{postgresql,postgresql-$OLD}
mv "$SRC_DIR/data/"{postgresql-$NEW,postgresql}

curl -fsSL -o "$SRC_DIR/data/postgres-$NEW-upgrade/optimize.sh" https://raw.githubusercontent.com/sourcegraph/sourcegraph/master/cmd/server/rootfs/postgres-optimize.sh

docker run \
  --entrypoint "/bin/bash" \
  -w /tmp/upgrade \
  -v "$SRC_DIR/data/postgres-$NEW-upgrade:/tmp/upgrade" \
  -v "$SRC_DIR/data/postgresql:/var/lib/postgresql/data" \
  "postgres:$NEW" \
  -c 'chown -R postgres $PGDATA && gosu postgres ./optimize.sh $PGDATA'
```

#### Kubernetes with https://github.com/sourcegraph/deploy-sourcegraph
