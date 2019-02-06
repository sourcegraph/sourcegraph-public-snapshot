# Postgres

Sourcegraph uses Postgres as its main internal database. We support any version **starting from 10.6**.

## Upgrades

This section describes the possible Postgres upgrade procedures for the different Sourcegraph deployment types.

### Automatic (recommended)

A few different deployment types can automatically handle Postgres data upgrades for you.

#### `sourcegraph/server`

Existing Postgres data will be automatically migrated when a release of the `sourcegraph/server` Docker image ships with a new version of Postgres.

For the upgrade to proceed, the Docker socket must be mounted the first time the new Docker image is ran.

This is needed to run the [Postgres upgrade containers](https://github.com/tianon/docker-postgres-upgrade) in the
Docker host. When the upgrade is done, the container can be restarted without mounting the Docker socket.

Docker host

```console
docker run -p 7080:7080 -p 2633:2633 --rm \
  -v ~/.sourcegraph/config:/etc/sourcegraph \
  -v ~/.sourcegraph/data:/var/opt/sourcegraph \
  -v /var/run/docker.sock:/var/run/docker.sock \
  sourcegraph/server:3.0.1

✱ Sourcegraph is upgrading its internal database. Please don't interrupt this operation.
✱ Sourcegraph finished upgrading its internal database.
01:35:58           postgres | 2019-02-04 01:35:58.341 UTC [77] LOG:  listening on IPv4 address "127.0.0.1", port 5432
01:35:59 management-console | t=2019-02-04T01:35:59+0000 lvl=info msg="management-console: listening" addr=:2633
01:36:01           frontend |
01:36:01           frontend |                    ╓╦╬╬╬╦╖
01:36:01           frontend |                   ╬╬╬╬╬╬╬╬╬
01:36:01           frontend |                  ╞╬╬╬╬╬╬╬╬╬╬           ╓╦╦╦╦┐
01:36:01           frontend |                   ╬╬╬╬╬╬╬╬╬╬╕        ╔╪╪╪╪╪╪╪╪╕
01:36:01           frontend |                   ╘╬╬╬╬╬╬╬╬╬╬      ╔╪╪╪╪╪╪╪╪╪╪╪
01:36:01           frontend |                    ╬╬╬╬╬╬╬╬╬╬╗   ╔╪╪╪╪╪╪╪╪╪╪╪╪┘
01:36:01           frontend |       ╓╦╦╖┐         ╬╬╬╬╬╬╬╬╬╬ ╔╝╪╪╪╪╪╪╪╪╪╪╪╜
01:36:01           frontend |     ╬╪╪╪╪╪╪╪╪╬╗╦╖   ╠╬╬╬╬╬╬╬╬╝╪╪╪╪╪╪╪╪╪╪╪╪╜
01:36:01           frontend |    ╠╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╝╝╬╬╬╬╝╪╪╪╪╪╪╪╪╪╪╪╪╜
01:36:01           frontend |    └╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╩
01:36:01           frontend |      ╙╩╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╖
01:36:01           frontend |           └╙╩╬╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╬╦╗┐
01:36:01           frontend |                  ╙╜╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╗╦╖
01:36:01           frontend |                  ┌╬╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╬
01:36:01           frontend |                ┌╗╪╪╪╪╪╪╪╪╪╪╪╪╬╬╬╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪╪
01:36:01           frontend |              ┌╗╪╪╪╪╪╪╪╪╪╪╪╪╬╬╬╬╬╬╬╬╬ ╙╜╩╪╪╪╪╪╪╪╪╪╪╪╪╜
01:36:01           frontend |            ┌╦╪╪╪╪╪╪╪╪╪╪╪╝╙╬╬╬╬╬╬╬╬╬╬       └╙╩╬╪╪╝╜
01:36:01           frontend |           ╦╪╪╪╪╪╪╪╪╪╪╪╪┘  ╠╬╬╬╬╬╬╬╬╬╬
01:36:01           frontend |          ╬╪╪╪╪╪╪╪╪╪╪╪┘     ╬╬╬╬╬╬╬╬╬╬┐
01:36:01           frontend |          ╙╪╪╪╪╪╪╪╪╪╜       ╘╬╬╬╬╬╬╬╬╬╬
01:36:01           frontend |           └╩╪╪╪╪╝╙          ╬╬╬╬╬╬╬╬╬╬╕
01:36:01           frontend |                             └╬╬╬╬╬╬╬╬╬╛
01:36:01           frontend |                               ╩╬╬╬╬╬╬┘
01:36:01           frontend |
01:36:01           frontend | ✱ Sourcegraph is ready at: http://127.0.0.1:7080
```

##### On Kubernetes


#### Kubernetes with https://github.com/sourcegraph/deploy-sourcegraph



#### External databases

#### Others

### Manual

If you prefer to manually upgrade Postgres, you can follow these instructions.

#### Docker host with `sourcegraph/server` image

```shell
set -euo pipefail

docker run -p 7080:7080 -p 2633:2633 --rm \
  -v ~/.sourcegraph/config:/etc/sourcegraph \
  -v ~/.sourcegraph/data:/var/opt/sourcegraph \
  -v /var/run/docker.sock:/var/run/docker.sock \
  tianon/postgres-upgrade:$OLD_VERSION-to-$NEW_VERSION
```

#### Kubernetes with https://github.com/sourcegraph/deploy-sourcegraph

#### External databases

#### Others

