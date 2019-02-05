# Postgres

Sourcegraph uses Postgres as its main internal database. We support any version **starting from 10.6**.

## Upgrades

This section describes the possible Postgres upgrade procedures for the different Sourcegraph deployment types.

### Automatic (recommended)

A few different deployment types can automatically handle Postgres data upgrades for you.

#### Docker host with `sourcegraph/server` image

If a new release of the `sourcegraph/server` image ships with an upgrade version of Postgres,
it'll automatically upgrade your existing Postgres data to be compatible with the new version.

All you need to do is mount the Docker socket the first time you run the new Docker image.

This is needed to run the Postgres upgrade containers defined at https://github.com/tianon/docker-postgres-upgrade in your
Docker host. When the upgrade is done, you can stop the container and start it again without the Docker socket mounted.

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

#### Kubernetes with `sourcegraph/server` image

#### Kubernetes with https://github.com/sourcegraph/deploy-sourcegraph

#### External databases

#### Others

### Manual

If you prefer to manually upgrade Postgres, you can follow these instructions.

#### Docker host with `sourcegraph/server` image

#### Kubernetes with `sourcegraph/server` image

#### Kubernetes with https://github.com/sourcegraph/deploy-sourcegraph

#### External databases

#### Others

