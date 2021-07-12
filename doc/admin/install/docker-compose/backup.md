# Backup sourcegraph databases in a Docker Compose deployment
The following instructions are specific to backing up and restoring the sourcegraph databases in a Docker Compose deployment. These do not apply to other deployment types. 
If you are not currently running sourcegraph via Docker Compose and would like to migrate, please see the [migration](./migrate.md). 

### Only core data will be backed up

These instructions will only back up core data including user accounts, configuration, repository-metadata, etc. Other data will be regenerated automatically:

* Repositories will be re-cloned
* Search indexes will be rebuilt from scratch

The above may take awhile if you have a lot of repositories. In the meantime, searches may be slow or return incomplete results. Usually this process will not take longer than 6 hours.


## Backup sourcegraph databases

These instuctions will back up the primary `sourcegraph` database the [codeintel](../../../code_intelligence/index.md) database.

1. `ssh` from your local machine into the machine hosting the `sourcegraph` deployment
2. `cd` to the `deploy-sourcegraph-docker/docker-compose` directory on the host
3. Verify the deployment is running:

```bash
docker-compose ps
          Name                         Command                       State                                                           Ports
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
caddy                       caddy run --config /etc/ca ...   Up                      2019/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp
cadvisor                    /usr/bin/cadvisor -logtost ...   Up (health: starting)   8080/tcp
codeinsights-db             docker-entrypoint.sh postgres    Up                      5432/tcp
codeintel-db                /postgres.sh                     Up (healthy)            5432/tcp
github-proxy                /sbin/tini -- /usr/local/b ...   Up
gitserver-0                 /sbin/tini -- /usr/local/b ...   Up
grafana                     /entry.sh                        Up                      3000/tcp, 0.0.0.0:3370->3370/tcp
jaeger                      /go/bin/all-in-one-linux - ...   Up                      0.0.0.0:14250->14250/tcp, 14268/tcp, 0.0.0.0:16686->16686/tcp, 5775/udp, 0.0.0.0:5778->5778/tcp,
                                                                                     0.0.0.0:6831->6831/tcp, 6831/udp, 0.0.0.0:6832->6832/tcp, 6832/udp
minio                       /usr/bin/docker-entrypoint ...   Up (healthy)            9000/tcp
pgsql                       /postgres.sh                     Up (healthy)            5432/tcp
precise-code-intel-worker   /sbin/tini -- /usr/local/b ...   Up (health: starting)   3188/tcp
prometheus                  /bin/prom-wrapper                Up                      0.0.0.0:9090->9090/tcp
query-runner                /sbin/tini -- /usr/local/b ...   Up
redis-cache                 /sbin/tini -- redis-server ...   Up                      6379/tcp
redis-store                 /sbin/tini -- redis-server ...   Up                      6379/tcp
repo-updater                /sbin/tini -- /usr/local/b ...   Up
searcher-0                  /sbin/tini -- /usr/local/b ...   Up (healthy)
symbols-0                   /sbin/tini -- /usr/local/b ...   Up (healthy)            3184/tcp
syntect-server              sh -c /http-server-stabili ...   Up (healthy)            9238/tcp
worker                      /sbin/tini -- /usr/local/b ...   Up                      3189/tcp
zoekt-indexserver-0         /sbin/tini -- zoekt-source ...   Up
zoekt-webserver-0           /sbin/tini -- /bin/sh -c z ...   Up (healthy)
```
4. Stop the deployment, and restart the databases service only to ensure there are no other connections during backup and restore.

```bash
docker-compose down
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

5. Generate the database dumps

```bash
docker exec pgsql sh -c 'pg_dump -C --username sg sg' > sourcegraph_db.out
docker exec codeintel-db -c 'pg_dump -C --username sg sg' > codeintel_db.out
```

6. Ensure the `sourcgraph_db.out` and `codeintel_db.out` files are moved to a safe and secure location. 

## Restore sourcgraph databases

### Restoring sourcegraph databases into a new environment

The following instructions apply only if you are restoring your databases into a new deployment of sourcegraph ie: a new virtual machine 

If you are restoring a previously running environment, see the instructions for [restoring a previously running deployment](#restoring-sourcegraph-databases-into-an-existing-environment)

1. Copy the database dump files, into the `deploy-sourcegraph-docker/docker-compose` directory. 
2. Start the database services

```bash
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

3. Copy the database files into the containers

```bash
docker cp sourcegraph_db.out pgsql:/tmp/sourecgraph_db.out
docker cp codeintel_db.out codeintel-db:/tmp/codeintel_db.out
```

4. Restore the databases

```bash
docker exec pgsql sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
docker exec codeintel-db sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

5. Start the remaining sourcegraph services

```bash
docker-compose up -d
```

6. Verify the deployment has started 

```bash 
docker-compose ps
          Name                         Command                       State                                                           Ports
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
caddy                       caddy run --config /etc/ca ...   Up                      2019/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp
cadvisor                    /usr/bin/cadvisor -logtost ...   Up (health: starting)   8080/tcp
codeinsights-db             docker-entrypoint.sh postgres    Up                      5432/tcp
codeintel-db                /postgres.sh                     Up (healthy)            5432/tcp
github-proxy                /sbin/tini -- /usr/local/b ...   Up
gitserver-0                 /sbin/tini -- /usr/local/b ...   Up
grafana                     /entry.sh                        Up                      3000/tcp, 0.0.0.0:3370->3370/tcp
jaeger                      /go/bin/all-in-one-linux - ...   Up                      0.0.0.0:14250->14250/tcp, 14268/tcp, 0.0.0.0:16686->16686/tcp, 5775/udp, 0.0.0.0:5778->5778/tcp,
                                                                                     0.0.0.0:6831->6831/tcp, 6831/udp, 0.0.0.0:6832->6832/tcp, 6832/udp
minio                       /usr/bin/docker-entrypoint ...   Up (healthy)            9000/tcp
pgsql                       /postgres.sh                     Up (healthy)            5432/tcp
precise-code-intel-worker   /sbin/tini -- /usr/local/b ...   Up (health: starting)   3188/tcp
prometheus                  /bin/prom-wrapper                Up                      0.0.0.0:9090->9090/tcp
query-runner                /sbin/tini -- /usr/local/b ...   Up
redis-cache                 /sbin/tini -- redis-server ...   Up                      6379/tcp
redis-store                 /sbin/tini -- redis-server ...   Up                      6379/tcp
repo-updater                /sbin/tini -- /usr/local/b ...   Up
searcher-0                  /sbin/tini -- /usr/local/b ...   Up (healthy)
symbols-0                   /sbin/tini -- /usr/local/b ...   Up (healthy)            3184/tcp
syntect-server              sh -c /http-server-stabili ...   Up (healthy)            9238/tcp
worker                      /sbin/tini -- /usr/local/b ...   Up                      3189/tcp
zoekt-indexserver-0         /sbin/tini -- zoekt-source ...   Up
zoekt-webserver-0           /sbin/tini -- /bin/sh -c z ...   Up (healthy)> docker-compose ps
```

7. Browse to your sourcegraph deployment, login and verify your existing configuration has been restored


### Restoring sourcegraph databases into an existing environment

1. `cd` to the `deploy-sourcegraph-docker/docker-compose` and stop the previous deployment and remove any existing volumes
```bash
docker-compose down
docker volume rm docker-compose_pgsql
docker volume rm docker-compose_codeintel-db
```

2. Start the databases services only
```bash
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

3. Copy the database files into the containers

```bash
docker cp sourcegraph_db.out pgsql:/tmp/sourecgraph_db.out
docker cp codeintel_db.out codeintel-db:/tmp/codeintel_db.out
```

4. Restore the databases

```bash
docker exec pgsql sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
docker exec codeintel-db sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

5. Start the remaining sourcegraph services

```bash
docker-compose up -d
```

6. Verify the deployment has started 

```bash 
docker-compose ps
          Name                         Command                       State                                                           Ports
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
caddy                       caddy run --config /etc/ca ...   Up                      2019/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp
cadvisor                    /usr/bin/cadvisor -logtost ...   Up (health: starting)   8080/tcp
codeinsights-db             docker-entrypoint.sh postgres    Up                      5432/tcp
codeintel-db                /postgres.sh                     Up (healthy)            5432/tcp
github-proxy                /sbin/tini -- /usr/local/b ...   Up
gitserver-0                 /sbin/tini -- /usr/local/b ...   Up
grafana                     /entry.sh                        Up                      3000/tcp, 0.0.0.0:3370->3370/tcp
jaeger                      /go/bin/all-in-one-linux - ...   Up                      0.0.0.0:14250->14250/tcp, 14268/tcp, 0.0.0.0:16686->16686/tcp, 5775/udp, 0.0.0.0:5778->5778/tcp,
                                                                                     0.0.0.0:6831->6831/tcp, 6831/udp, 0.0.0.0:6832->6832/tcp, 6832/udp
minio                       /usr/bin/docker-entrypoint ...   Up (healthy)            9000/tcp
pgsql                       /postgres.sh                     Up (healthy)            5432/tcp
precise-code-intel-worker   /sbin/tini -- /usr/local/b ...   Up (health: starting)   3188/tcp
prometheus                  /bin/prom-wrapper                Up                      0.0.0.0:9090->9090/tcp
query-runner                /sbin/tini -- /usr/local/b ...   Up
redis-cache                 /sbin/tini -- redis-server ...   Up                      6379/tcp
redis-store                 /sbin/tini -- redis-server ...   Up                      6379/tcp
repo-updater                /sbin/tini -- /usr/local/b ...   Up
searcher-0                  /sbin/tini -- /usr/local/b ...   Up (healthy)
symbols-0                   /sbin/tini -- /usr/local/b ...   Up (healthy)            3184/tcp
syntect-server              sh -c /http-server-stabili ...   Up (healthy)            9238/tcp
worker                      /sbin/tini -- /usr/local/b ...   Up                      3189/tcp
zoekt-indexserver-0         /sbin/tini -- zoekt-source ...   Up
zoekt-webserver-0           /sbin/tini -- /bin/sh -c z ...   Up (healthy)> docker-compose ps
```

7. Browse to your sourcegraph deployment, login and verify your existing configuration has been restored

