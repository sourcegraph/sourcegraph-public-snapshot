# High level overview

Sourcegraph's user data can be migrated from the single Docker image (`sourcegraph/server`) to the Docker Compose deployment by dumping and restoring the Postgres database.

## Notes

### Version requirements

* This migration can only be done with Sourcegraph `3.13.0` and above (e.g. `sourcegraph/server:3.13.0` and [v3.13.0 (FILL IN RELEASE HERE) Docker Compose](TODO) ).
* Sourcegraph's user data can only be transferred between deployments that are running the same Sourcegraph verion (e.g. `sourcegraph/server:3.13.0` can only transfer its data to `v3.13.0` of the Docker Compose definition). If you're running a version of Sourcegraph server that's older than the Docker Compose deployment version, you **must** upgrade to a newer `sourcegraph/server` version before continuing.

### Storage Location

Note that after this process, Sourcegraph's data will be stored in Docker volumes instead of -  `~/.sourcegraph/.` (which defaulted to `~/sourcegraph-docker`). See  [https://docs.docker.com/storage/volumes/](https://docs.docker.com/storage/volumes/) for more information about Docker volumes. By default on Linux, local Docker volumes are stored at `/var/lib/docker/volumes`.

## Notes

### External database

These instructions only apply if your single Docker image is running the built-in Postgres server. If you're running an external database, simply follow the [CLOUD - PROVIDER  DOCUmENTATION] to setup your new Docker compose instance and following the [DOCKER -COMPOSE EXTERNAL DATABAE /i]
* While this process will migrate your user data, the new Docker Compose deployment will need to regenerate all the other ephemberal data: all repositories will need to be re-cloned, search indicies will need to be recreated, etc.
* If you're using an [external database (LINK HERE)]()
* After this process, Sourcegraph's data will be stored in Docker volumes instead of -  `$DATA_ROOT` (which defaulted to `~/sourcegraph-docker`). See  [https://docs.docker.com/storage/volumes/](https://docs.docker.com/storage/volumes/) for more information about Docker volumes. By default on Linux, local Docker volumes are stored at `/var/lib/docker/volumes`.

1. Backup Postgres database
2. Upgrade Postgres container only
3. Restore Postgres database
4. Upgrade all other containers

Note that after this process, Sourcegraph's data will be stored in Docker volumes instead of -  `$DATA_ROOT` (which defaulted to `~/sourcegraph-docker`). See  [https://docs.docker.com/storage/volumes/](https://docs.docker.com/storage/volumes/) for more information about Docker volumes. By default on Linux, local Docker volumes are stored at `/var/lib/docker/volumes`.

## Version requirements

The `sourcegraph/server` image that that you're migrating from **must** match the same Docker Compose version that you're upgrading to. If you're running an older `sourcegraph 


## Backup Postgres

Find the `CONTAINER_ID` of the `sourcegraph/server` image from the `docker ps` output:

```bash

> docker ps
CONTAINER ID        IMAGE
...                 sourcegraph/server
```

Generate Postgres dump inside sourcegraph/server container 

```bash
# Open a shell inside sourcegraph/server using the CONTAINER_ID found in the previous step
> docker exec -it "${CONTAINER_ID}" /bin/sh

> docker exec -it pgsql /bin/sh

# Dump Postgres database to db.out file 
> pg_dumpall --verbose --username=postgres > /tmp/db.out

# Exit container shell session
> exit
```

### Copy Postgres dump from the container onto host machine

```bash
> docker cp "${CONTAINER_ID}":/tmp/db.out ~/db.out

# You can run "less ~/db.out" to verify that it has the contents that you expect
```

### Tear down the existing `sourcegraph/server` container

```
> docker stop "${CONTAINER_ID}"
```

### Initialize the Docker Compose database 

```bash
> docker-compose up -d
> docker-compose down 
```

### Bring up the postgres database on it's own 

```
docker-compose-f pgsql-only-migrate.docker-compose.yaml up -d
```

```
docker cp ~/db.out pgsql:/tmp/db.out
docker exec -it pgsql /bin/sh
psql --username=sg -f /tmp/db.out postgres
```

```psql
> DROP DATABASE sg;

> ALTER DATABASE sourcegraph RENAME TO sg;

> ALERT DATABASE sg OWNER TO sg;
```

```
docker-compose -f docker-compose.yaml up -d 
```

### Copy Postgres dump from the container onto host machine

    > docker cp "${CONTAINER_ID}":/tmp/db.out ~/db.out
    
    # You can run "less ~/db.out" to verify that it has the contents that you expect

### Tear down the existing `sourcegraph/server` container

    > docker stop "${CONTAINER_ID}"

### Generate Postgres dump

```bash
# Login to the pgsql container:
> docker exec -it pgsql /bin/sh

# Dump database to db.out file
> pg_dumpall --verbose --username=sg > /tmp/db.out

# Exit container shell session
> exit
```

### Copy Postgres dump from the container onto host machine

```bash

> docker cp pgsql:/tmp/db.out ~/db.out

# You can run "less ~/db.out" to verify that it has the contents that you expect

```

### Stop the exiting `sourcegraph/server` image

```bash
> docker stop $
```

### Tear down the existing docker-compose deployment

    > docker-compose down 

## Upgrade Postgres

### Grab deploy-sourcegraph-docker [v3.12.3](https://github.com/sourcegraph/deploy-sourcegraph-docker/releases/tag/v3.12.3)

    > git pull
    > git checkout v3.12.3

### Deploy the postgres container on its own

1. Create a new `pgsql-only-docker-compose.yaml`, that has all services besides `pgsql` commented out, eg:

        version: '2.4'
        services:
          # # Description: Acts as a reverse proxy for all of the sourcegraph-frontend instances
          # #
          # # Disk: none
          # # Ports exposed to other Sourcegraph services: none
          # # Ports exposed to the public internet: 7080 (HTTP) and/or 7443 (HTTPS)
          # #
          # # Note: nginx/nginx/sourcegraph_backend.conf lists all of the sourcegraph-frontend
          # # network addresses that Nginx should proxy. nginx/nginx/sourcegraph_backend.conf
          # # needs to be updated with new addresses when scaling the number of soucegraph-frontend
          # # replicas.
          # nginx:
          #   container_name: nginx
          #   image: 'index.docker.io/library/nginx:1.17.7@sha256:8aa7f6a9585d908a63e5e418dc5d14ae7467d2e36e1ab4f0d8f9d059a3d071ce'
          #   cpus: 1
          #   mem_limit: '1g'
          #   volumes:
          #     - '../nginx:/etc/nginx'
          #   ports:
          #     - '0.0.0.0:80:7080'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Serves the frontend of Sourcegraph via HTTP(S).
          # #
          # # Disk: 128GB / non-persistent SSD
          # # Ports exposed to other Sourcegraph services: 6060/TCP, 3080 (HTTP), and/or 3443 (HTTPS)
          # # Ports exposed to the public internet: none
          # #
          # # Note: SRC_GIT_SERVERS, SEARCHER_URL, and SYMBOLS_URL are space-separated
          # # lists which each allow you to specify more container instances for scaling
          # # purposes. Be sure to also apply such a change here to the frontend-internal
          # # service.
          # sourcegraph-frontend-0:
          #   container_name: sourcegraph-frontend-0
          #   image: 'index.docker.io/sourcegraph/frontend:3.12.3@sha256:d5c1d50ce370b3c730f33efe60f46b5106b39551c4771fb036fd89e5bba3a16d'
          #   cpus: 4
          #   mem_limit: '8g'
          #   environment:
          #     - GOMAXPROCS=12
          #     - JAEGER_AGENT_HOST=jaeger-agent
          #     - PGHOST=pgsql
          #     - 'SRC_GIT_SERVERS=gitserver-0:3178'
          #     - 'SRC_SYNTECT_SERVER=http://syntect-server:9238'
          #     - 'SEARCHER_URL=http://searcher-0:3181'
          #     - 'SYMBOLS_URL=http://symbols-0:3184'
          #     - 'INDEXED_SEARCH_SERVERS=zoekt-webserver-0:6070'
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - 'REPO_UPDATER_URL=http://repo-updater:3182'
          #     - 'REPLACER_URL=http://replacer:3185'
          #     - 'LSIF_SERVER_URL=http://lsif-server:3186'
          #     - 'GRAFANA_SERVER_URL=http://grafana:3370'
          #   healthcheck:
          #     test: "wget -q 'http://127.0.0.1:3080/healthz' -O /dev/null || exit 1"
          #     interval: 5s
          #     timeout: 10s
          #     retries: 3
          #     start_period: 300s
          #   volumes:
          #     - 'sourcegraph-frontend-0:/mnt/cache'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Serves the internal Sourcegraph frontend API.
          # #
          # # Disk: 128GB / non-persistent SSD
          # # Ports exposed to other Sourcegraph services: 3090/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # sourcegraph-frontend-internal:
          #   container_name: sourcegraph-frontend-internal
          #   image: 'index.docker.io/sourcegraph/frontend:3.12.3@sha256:d5c1d50ce370b3c730f33efe60f46b5106b39551c4771fb036fd89e5bba3a16d'
          #   cpus: 4
          #   mem_limit: '8g'
          #   environment:
          #     - GOMAXPROCS=4
          #     - PGHOST=pgsql
          #     - 'SRC_GIT_SERVERS=gitserver-0:3178'
          #     - 'SRC_SYNTECT_SERVER=http://syntect-server:9238'
          #     - 'SEARCHER_URL=http://searcher-0:3181'
          #     - 'SYMBOLS_URL=http://symbols-0:3184'
          #     - 'INDEXED_SEARCH_SERVERS=zoekt-webserver-0:6070'
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - 'REPO_UPDATER_URL=http://repo-updater:3182'
          #     - 'REPLACER_URL=http://replacer:3185'
          #     - 'LSIF_SERVER_URL=http://lsif-server:3186'
          #     - 'GRAFANA_SERVER_URL=http://grafana:3000'
          #   volumes:
          #     - 'sourcegraph-frontend-internal-0:/mnt/cache'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Stores clones of repositories to perform Git operations.
          # #
          # # Disk: 200GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: 3178/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # gitserver-0:
          #   container_name: gitserver-0
          #   image: 'index.docker.io/sourcegraph/gitserver:3.12.3@sha256:46de0ef5d67888ad7c7c883f46f9c537a1e2efd5cc8bc3e1bfda5c37b7ddb0e7'
          #   cpus: 4
          #   mem_limit: '8g'
          #   environment:
          #     - GOMAXPROCS=4
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - JAEGER_AGENT_HOST=jaeger-agent
          #   volumes:
          #     - 'gitserver-0:/data/repos'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Backend for indexed text search operations.
          # #
          # # Disk: 200GB / persistent SSD
          # # Network: 100mbps
          # # Liveness probe: n/a
          # # Ports exposed to other Sourcegraph services: 6072/TCP
          # # Ports exposed to the public internet: none
          # #
          # zoekt-indexserver-0:
          #   container_name: zoekt-indexserver-0
          #   image: 'index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200124185115-83b89a5@sha256:efd1fb37fc62bfab963f12e95f69778b0e2e6a253caed5be9025840072ea85b5'
          #   mem_limit: '16g'
          #   environment:
          #     - GOMAXPROCS=8
          #     - 'HOSTNAME=zoekt-webserver-0:6070'
          #     - 'SRC_FRONTEND_INTERNAL=http://sourcegraph-frontend-internal:3090'
          #   volumes:
          #     - 'zoekt-0-shared:/data/index'
          #   networks:
          #     - sourcegraph
          #   restart: always
          #   hostname: zoekt-indexserver-0
          # # Description: Backend for indexed text search operations.
          # #
          # # Disk: 200GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: 6070/TCP
          # # Ports exposed to the public internet: none
          # #
          # zoekt-webserver-0:
          #   container_name: zoekt-webserver-0
          #   image: 'index.docker.io/sourcegraph/zoekt-webserver:0.0.20200124185328-83b89a5@sha256:cde27ee7db0fe6c293a8c9df47b529fb01b5a898e6cbeea4c18d80fe218563db'
          #   cpus: 8
          #   mem_limit: '50g'
          #   environment:
          #     - GOMAXPROCS=8
          #     - 'HOSTNAME=zoekt-webserver-0:6070'
          #   healthcheck:
          #     test: "wget -q 'http://127.0.0.1:6070/healthz' -O /dev/null || exit 1"
          #     interval: 1s
          #     timeout: 10s
          #     retries: 1
          #   volumes:
          #     - 'zoekt-0-shared:/data/index'
          #   networks:
          #     - sourcegraph
          #   restart: always
          #   hostname: zoekt-webserver-0
        
          # # Description: Backend for text search operations.
          # #
          # # Disk: 128GB / non-persistent SSD
          # # Ports exposed to other Sourcegraph services: 3181/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # searcher-0:
          #   container_name: searcher-0
          #   image: 'index.docker.io/sourcegraph/searcher:3.12.3@sha256:c31deb621158d1fc0f7d19ca50cca8b2c0aba957736437f485e206fb10367bf7'
          #   cpus: 2
          #   mem_limit: '2g'
          #   environment:
          #     - GOMAXPROCS=2
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - JAEGER_AGENT_HOST=jaeger-agent
          #   healthcheck:
          #     test: "wget -q 'http://127.0.0.1:3181/healthz' -O /dev/null || exit 1"
          #     interval: 1s
          #     timeout: 10s
          #     retries: 1
          #   volumes:
          #     - 'searcher-0:/mnt/cache'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Rate-limiting proxy for the GitHub API.
          # #
          # # CPU: 1
          # # Memory: 1GB
          # # Disk: 1GB / non-persistent SSD (only for read-only config file)
          # # Ports exposed to other Sourcegraph services: 3180/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # github-proxy:
          #   container_name: github-proxy
          #   image: 'index.docker.io/sourcegraph/github-proxy:3.12.3@sha256:969d5107c7314e196c7e1ac223b7c65d6bfcfe1e4ccb4a08da67432827b670b2'
          #   cpus: 1
          #   mem_limit: '1g'
          #   environment:
          #     - GOMAXPROCS=1
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - JAEGER_AGENT_HOST=jaeger-agent
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: LSIF HTTP server for code intelligence.
          # #
          # # Disk: 200GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: 3186/TCP (server) 3187/TCP (worker)
          # # Ports exposed to the public internet: none
          # #
          # lsif-server:
          #   container_name: lsif-server
          #   image: 'index.docker.io/sourcegraph/lsif-server:3.12.3@sha256:078ed33647571a21d6ae3393faf46b976a08aa3c1312b6f5272ec6812b1f0f6e'
          #   cpus: 2
          #   mem_limit: '2g'
          #   environment:
          #     - GOMAXPROCS=2
          #     - LSIF_STORAGE_ROOT=/lsif-storage
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #   healthcheck:
          #     test: "wget -q 'http://127.0.0.1:3186/ping' -O /dev/null || exit 1"
          #     interval: 5s
          #     timeout: 5s
          #     retries: 1
          #     start_period: 60s
          #   volumes:
          #     - 'lsif-server:/lsif-storage'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Saved search query runner / notification service.
          # #
          # # Disk: 1GB / non-persistent SSD (only for read-only config file)
          # # Network: 100mbps
          # # Liveness probe: n/a
          # # Ports exposed to other Sourcegraph services: 3183/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # query-runner:
          #   container_name: query-runner
          #   image: 'index.docker.io/sourcegraph/query-runner:3.12.3@sha256:f0fc1c1b8b19937708687eed6bb2456fcc8433083127f9f9f5ce531edcf39e2b'
          #   cpus: 1
          #   mem_limit: '1g'
          #   environment:
          #     - GOMAXPROCS=1
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - JAEGER_AGENT_HOST=jaeger-agent
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Backend for replace operations.
          # #
          # # Disk: 128GB / non-persistent SSD
          # # Ports exposed to other Sourcegraph services: 3185/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # replacer:
          #   container_name: replacer
          #   image: 'index.docker.io/sourcegraph/replacer:3.12.3@sha256:f5a40b66614cbb72602e7c31f46232c16919477800a965f5bc5a568c235e7578'
          #   cpus: 1
          #   mem_limit: '512m'
          #   environment:
          #     - GOMAXPROCS=1
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #   healthcheck:
          #     test: "wget -q 'http://127.0.0.1:3185/healthz' -O /dev/null || exit 1"
          #     interval: 1s
          #     timeout: 10s
          #     retries: 1
          #   volumes:
          #     - 'replacer:/mnt/cache'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Handles repository metadata (not Git data) lookups and updates from external code hosts and other similar services.
          # #
          # # Disk: 128GB / non-persistent SSD
          # # Ports exposed to other Sourcegraph services: 3182/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # repo-updater:
          #   container_name: repo-updater
          #   image: 'index.docker.io/sourcegraph/repo-updater:3.12.3@sha256:1a7a29bebd1d142401051c3c0e8d9b0763d0c7d10df776f8ff342cad66af1d88'
          #   cpus: 4
          #   mem_limit: '4g'
          #   environment:
          #     - GOMAXPROCS=1
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - JAEGER_AGENT_HOST=jaeger-agent
          #     - 'GITHUB_BASE_URL=http://github-proxy:3180'
          #   volumes:
          #     - 'repo-updater:/mnt/cache'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Backend for syntax highlighting operations.
          # #
          # # Disk: none
          # # Ports exposed to other Sourcegraph services: 9238/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # syntect-server:
          #   container_name: syntect-server
          #   image: 'index.docker.io/sourcegraph/syntect_server:2b5a3fb@sha256:ef5529cafdc68d5a21edea472ee8ad966878b173044aa5c3db93bc3d84765b1f'
          #   cpus: 4
          #   mem_limit: '6g'
          #   healthcheck:
          #     test: "wget -q 'http://127.0.0.1:9238/health' -O /dev/null || exit 1"
          #     interval: 1s
          #     timeout: 5s
          #     retries: 1
          #     start_period: 5s
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Backend for symbols operations.
          # #
          # # Disk: 128GB / non-persistent SSD
          # # Ports exposed to other Sourcegraph services: 3184/TCP 6060/TCP
          # # Ports exposed to the public internet: none
          # #
          # symbols-0:
          #   container_name: symbols-0
          #   image: 'index.docker.io/sourcegraph/symbols:3.12.3@sha256:29861594850db42724443e4c58a824696314d7566ec508f6ea85e38f693bf0fb'
          #   cpus: 2
          #   mem_limit: '4g'
          #   environment:
          #     - GOMAXPROCS=2
          #     - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
          #     - JAEGER_AGENT_HOST=jaeger-agent
          #   healthcheck:
          #     test: "wget -q 'http://127.0.0.1:3184/healthz' -O /dev/null || exit 1"
          #     interval: 5s
          #     timeout: 5s
          #     retries: 1
          #     start_period: 60s
          #   volumes:
          #     - 'symbols-0:/mnt/cache'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Prometheus collects metrics and aggregates them into graphs.
          # #
          # # Disk: 200GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: none
          # # Ports exposed to the public internet: none (HTTP 9090 should be exposed to admins only)
          # #
          # prometheus:
          #   container_name: prometheus
          #   image: 'index.docker.io/sourcegraph/prometheus:10.0.7@sha256:22d54f27c7df8733a06c7ae8c2e851b61b1ed42f1f5621d493ef58ebd8d815e0'
          #   cpus: 4
          #   mem_limit: '8g'
          #   volumes:
          #     - 'prometheus-v2:/prometheus'
          #     - '../prometheus:/sg_prometheus_add_ons'
          #   ports:
          #     - '0.0.0.0:9090:9090'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Dashboards and graphs for Prometheus metrics.
          # #
          # # Disk: 100GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: none
          # # Ports exposed to the public internet: none (HTTP 3000 should be exposed to admins only)
          # #
          # # Add the following environment variables if you wish to use an auth proxy with Grafana:
          # #
          # # 'GF_AUTH_PROXY_ENABLED=true'
          # # 'GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User'
          # # 'GF_SERVER_ROOT_URL='https://grafana.example.com'
          # grafana:
          #   container_name: grafana
          #   image: 'index.docker.io/sourcegraph/grafana:10.0.9@sha256:0132e5602030145803753468497a2d17640164b9c34df4ce2532dd93e4b1f6fc'
          #   cpus: 1
          #   mem_limit: '1g'
          #   volumes:
          #     - 'grafana:/var/lib/grafana'
          #     - '../grafana/datasources:/sg_config_grafana/provisioning/datasources'
          #     - '../grafana/dashboards:/sg_grafana_additional_dashboards'
          #   ports:
          #     - '0.0.0.0:3370:3370'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Publishes Prometheus metrics about Docker containers.
          # #
          # # Disk: none
          # # Ports exposed to other Sourcegraph services: 8080/TCP
          # # Ports exposed to the public internet: none
          # #
          # cadvisor:
          #   container_name: cadvisor
          #   image: 'google/cadvisor:v0.33.0'
          #   cpus: 1
          #   mem_limit: '1g'
          #   volumes:
          #     - '/:/rootfs:ro'
          #     - '/var/run:/var/run:ro'
          #     - '/sys:/sys:ro'
          #     - '/var/lib/docker/:/var/lib/docker:ro'
          #     - '/dev/disk/:/dev/disk:ro'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # # Description: Jaeger agent which is local to the host machine (containers on
          # # # the machine send trace information to it and it relays to the collector).
          # # #
          # # # Disk: none
          # # # Ports exposed to other Sourcegraph services: 5775/UDP 6831/UDP 6832/UDP (on the same host machine)
          # # # Ports exposed to the public internet: none
          # # #
          # # jaeger-agent:
          # #   container_name: jaeger-agent
          # #   image: 'index.docker.io/jaegertracing/jaeger-agent@sha256:7ad33c19fd66307f2a3c07c95eb07c335ddce1b487f6b6128faa75d042c496cb'
          # #   cpus: 1
          # #   mem_limit: '1g'
          # #   environment:
          # #     - "COLLECTOR_HOST_PORT='jaeger-collector:14267'"
          # #   networks:
          # #     - sourcegraph
          # #   restart: always
        
          # # # Description: Jaeger's Cassandra database for storing traces.
          # # #
          # # # Disk: 128GB / persistent SSD
          # # # Ports exposed to other Sourcegraph services: 9042/TCP
          # # # Ports exposed to the public internet: none
          # # #
          # # jaeger-cassandra:
          # #   container_name: jaeger-cassandra
          # #   image: 'index.docker.io/library/cassandra:3.11.4@sha256:9f1d47fd23261c49f226546fe0134e6d4ad0570b7ea3a169c521005cb8369a32'
          # #   cpus: 4
          # #   mem_limit: '8g'
          # #   environment:
          # #     - 'HEAP_NEWSIZE=1G'
          # #     - 'MAX_HEAP_SIZE=6G'
          # #     - 'CASSANDRA_DC=sourcegraph'
          # #     - 'CASSANDRA_RACK=rack1'
          # #     - 'CASSANDRA_ENDPOINT_SNITCH=GossipingPropertyFileSnitch'
          # #   volumes:
          # #     - 'jaeger-cassandra:/var/lib/cassandra'
          # #   networks:
          # #     - sourcegraph
          # #   restart: always
        
          # # # Description: Receives traces from Jaeger agents.
          # # #
          # # # Disk: none
          # # # Ports exposed to other Sourcegraph services: 14267/TCP
          # # # Ports exposed to the public internet: none
          # # #
          # # jaeger-collector:
          # #   container_name: jaeger-collector
          # #   image: 'index.docker.io/jaegertracing/jaeger-collector:1.11@sha256:0b6d28bb52410f7b50c0f0fc16d7ee391e2e3eca47b713ac88d0891ca8a63cb9'
          # #   cpus: 1
          # #   mem_limit: '1g'
          # #   environment:
          # #     - SPAN_STORAGE_TYPE=cassandra
          # #     - CASSANDRA_SERVERS=jaeger-cassandra
          # #     - CASSANDRA_KEYSPACE=jaeger_v1_sourcegraph
          # #   networks:
          # #     - sourcegraph
          # #   restart: always
        
          # # # Description: Jaeger frontend for querying traces.
          # # #
          # # # Disk: none
          # # # Ports exposed to other Sourcegraph services: none
          # # # Ports exposed to the public internet: none (HTTP 16686 should be exposed to admins only)
          # # #
          # # jaeger-query:
          # #   container_name: jaeger-query
          # #   image: 'index.docker.io/jaegertracing/jaeger-query:1.11@sha256:cddc521d0166c868931282685a863368ae2c14c4de0c1be38e388ece3080439e'
          # #   cpus: 1
          # #   mem_limit: '1g'
          # #   environment:
          # #     - SPAN_STORAGE_TYPE=cassandra
          # #     - CASSANDRA_SERVERS=jaeger-cassandra
          # #     - CASSANDRA_KEYSPACE=jaeger_v1_sourcegraph
          # #     - CASSANDRA_LOCAL_DC=sourcegraph
          # #   ports:
          # #     - '0.0.0.0:16686:16686'
          # #   networks:
          # #     - sourcegraph
          # #   restart: always
        
          # Description: PostgreSQL database for various data.
          #
          # Disk: 128GB / persistent SSD
          # Ports exposed to other Sourcegraph services: 5432/TCP 9187/TCP
          # Ports exposed to the public internet: none
          #
          pgsql:
            container_name: pgsql
            image: 'index.docker.io/sourcegraph/postgres-11.4:19-11-14_b084311b@sha256:072481559d559cfd9a53ad77c3688b5cf583117457fd452ae238a20405923297'
            cpus: 4
            mem_limit: '2g'
            healthcheck:
              test: '/liveness.sh'
              interval: 10s
              timeout: 1s
              retries: 3
              start_period: 15s
            volumes:
              - 'pgsql:/data/'
            networks:
              - sourcegraph
            restart: always
        
          # # Description: Redis for storing short-lived caches.
          # #
          # # Disk: 128GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: 6379/TCP 9121/TCP
          # # Ports exposed to the public internet: none
          # #
          # redis-cache:
          #   container_name: redis-cache
          #   image: index.docker.io/sourcegraph/redis-cache:19-04-16_6891de82@sha256:4cbfac8af0abb673899250d4fd859cc477d6426de519e9deb71e454e18322499
          #   cpus: 1
          #   mem_limit: '6g'
          #   volumes:
          #     - 'redis-cache:/redis-data'
          #   networks:
          #     - sourcegraph
          #   restart: always
          # # Description: Redis for storing semi-persistent data like user sessions.
          # #
          # # Disk: 128GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: 6379/TCP 9121/TCP
          # # Ports exposed to the public internet: none
          # #
          # redis-store:
          #   container_name: redis-store
          #   image: 'index.docker.io/sourcegraph/redis-store:19-04-16_6891de821@sha256:56426d601ce1f6d63088fea1cefa61f69a2e809c7d90fc1d157cca63cf81b277'
          #   cpus: 1
          #   mem_limit: '6g'
          #   volumes:
          #     - 'redis-store:/redis-data'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
        volumes:
          gitserver-0:
          grafana:
          #jaeger-cassandra:
          lsif-server:
          pgsql:
          prometheus-v2:
          redis-cache:
          redis-store:
          replacer:
          repo-updater:
          searcher-0:
          sourcegraph-frontend-0:
          sourcegraph-frontend-internal-0:
          symbols-0:
          zoekt-0-shared:
        networks:
          sourcegraph:

1. Deploy the `pgsql` container 

        > docker-compose up -f pgsql-only-docker-compose.yaml -d 
        # docker ps should show that only the pgsql container is running

## Restore the Postgres database

    # Copy the Postgres dump into the new container 
    > docker cp ~/db.out pgsql:/tmp/db.out
    
    # Login to the pgsql container
    > docker exec -it pgsql /bin/sh
    
    # Restore the Postgres dump 
    > psql --username=sg -f /tmp/db.out postgres
    
    # Exit the "pgsql" shell session
    > exit 

## Upgrade the other containers

1. Uncomment the other services in the docker-compose.yaml file, eg: 

        version: '2.4'
        services:
          # Description: Acts as a reverse proxy for all of the sourcegraph-frontend instances
          #
          # Disk: none
          # Ports exposed to other Sourcegraph services: none
          # Ports exposed to the public internet: 7080 (HTTP) and/or 7443 (HTTPS)
          #
          # Note: nginx/nginx/sourcegraph_backend.conf lists all of the sourcegraph-frontend
          # network addresses that Nginx should proxy. nginx/nginx/sourcegraph_backend.conf
          # needs to be updated with new addresses when scaling the number of soucegraph-frontend
          # replicas.
          nginx:
            container_name: nginx
            image: 'index.docker.io/library/nginx:1.17.7@sha256:8aa7f6a9585d908a63e5e418dc5d14ae7467d2e36e1ab4f0d8f9d059a3d071ce'
            cpus: 1
            mem_limit: '1g'
            volumes:
              - '../nginx:/etc/nginx'
            ports:
              - '0.0.0.0:80:7080'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Serves the frontend of Sourcegraph via HTTP(S).
          #
          # Disk: 128GB / non-persistent SSD
          # Ports exposed to other Sourcegraph services: 6060/TCP, 3080 (HTTP), and/or 3443 (HTTPS)
          # Ports exposed to the public internet: none
          #
          # Note: SRC_GIT_SERVERS, SEARCHER_URL, and SYMBOLS_URL are space-separated
          # lists which each allow you to specify more container instances for scaling
          # purposes. Be sure to also apply such a change here to the frontend-internal
          # service.
          sourcegraph-frontend-0:
            container_name: sourcegraph-frontend-0
            image: 'index.docker.io/sourcegraph/frontend:3.12.3@sha256:d5c1d50ce370b3c730f33efe60f46b5106b39551c4771fb036fd89e5bba3a16d'
            cpus: 4
            mem_limit: '8g'
            environment:
              - GOMAXPROCS=12
              - JAEGER_AGENT_HOST=jaeger-agent
              - PGHOST=pgsql
              - 'SRC_GIT_SERVERS=gitserver-0:3178'
              - 'SRC_SYNTECT_SERVER=http://syntect-server:9238'
              - 'SEARCHER_URL=http://searcher-0:3181'
              - 'SYMBOLS_URL=http://symbols-0:3184'
              - 'INDEXED_SEARCH_SERVERS=zoekt-webserver-0:6070'
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - 'REPO_UPDATER_URL=http://repo-updater:3182'
              - 'REPLACER_URL=http://replacer:3185'
              - 'LSIF_SERVER_URL=http://lsif-server:3186'
              - 'GRAFANA_SERVER_URL=http://grafana:3370'
            healthcheck:
              test: "wget -q 'http://127.0.0.1:3080/healthz' -O /dev/null || exit 1"
              interval: 5s
              timeout: 10s
              retries: 3
              start_period: 300s
            volumes:
              - 'sourcegraph-frontend-0:/mnt/cache'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Serves the internal Sourcegraph frontend API.
          #
          # Disk: 128GB / non-persistent SSD
          # Ports exposed to other Sourcegraph services: 3090/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          sourcegraph-frontend-internal:
            container_name: sourcegraph-frontend-internal
            image: 'index.docker.io/sourcegraph/frontend:3.12.3@sha256:d5c1d50ce370b3c730f33efe60f46b5106b39551c4771fb036fd89e5bba3a16d'
            cpus: 4
            mem_limit: '8g'
            environment:
              - GOMAXPROCS=4
              - PGHOST=pgsql
              - 'SRC_GIT_SERVERS=gitserver-0:3178'
              - 'SRC_SYNTECT_SERVER=http://syntect-server:9238'
              - 'SEARCHER_URL=http://searcher-0:3181'
              - 'SYMBOLS_URL=http://symbols-0:3184'
              - 'INDEXED_SEARCH_SERVERS=zoekt-webserver-0:6070'
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - 'REPO_UPDATER_URL=http://repo-updater:3182'
              - 'REPLACER_URL=http://replacer:3185'
              - 'LSIF_SERVER_URL=http://lsif-server:3186'
              - 'GRAFANA_SERVER_URL=http://grafana:3000'
            volumes:
              - 'sourcegraph-frontend-internal-0:/mnt/cache'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Stores clones of repositories to perform Git operations.
          #
          # Disk: 200GB / persistent SSD
          # Ports exposed to other Sourcegraph services: 3178/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          gitserver-0:
            container_name: gitserver-0
            image: 'index.docker.io/sourcegraph/gitserver:3.12.3@sha256:46de0ef5d67888ad7c7c883f46f9c537a1e2efd5cc8bc3e1bfda5c37b7ddb0e7'
            cpus: 4
            mem_limit: '8g'
            environment:
              - GOMAXPROCS=4
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - JAEGER_AGENT_HOST=jaeger-agent
            volumes:
              - 'gitserver-0:/data/repos'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Backend for indexed text search operations.
          #
          # Disk: 200GB / persistent SSD
          # Network: 100mbps
          # Liveness probe: n/a
          # Ports exposed to other Sourcegraph services: 6072/TCP
          # Ports exposed to the public internet: none
          #
          zoekt-indexserver-0:
            container_name: zoekt-indexserver-0
            image: 'index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200124185115-83b89a5@sha256:efd1fb37fc62bfab963f12e95f69778b0e2e6a253caed5be9025840072ea85b5'
            mem_limit: '16g'
            environment:
              - GOMAXPROCS=8
              - 'HOSTNAME=zoekt-webserver-0:6070'
              - 'SRC_FRONTEND_INTERNAL=http://sourcegraph-frontend-internal:3090'
            volumes:
              - 'zoekt-0-shared:/data/index'
            networks:
              - sourcegraph
            restart: always
            hostname: zoekt-indexserver-0
          # Description: Backend for indexed text search operations.
          #
          # Disk: 200GB / persistent SSD
          # Ports exposed to other Sourcegraph services: 6070/TCP
          # Ports exposed to the public internet: none
          #
          zoekt-webserver-0:
            container_name: zoekt-webserver-0
            image: 'index.docker.io/sourcegraph/zoekt-webserver:0.0.20200124185328-83b89a5@sha256:cde27ee7db0fe6c293a8c9df47b529fb01b5a898e6cbeea4c18d80fe218563db'
            cpus: 8
            mem_limit: '50g'
            environment:
              - GOMAXPROCS=8
              - 'HOSTNAME=zoekt-webserver-0:6070'
            healthcheck:
              test: "wget -q 'http://127.0.0.1:6070/healthz' -O /dev/null || exit 1"
              interval: 1s
              timeout: 10s
              retries: 1
            volumes:
              - 'zoekt-0-shared:/data/index'
            networks:
              - sourcegraph
            restart: always
            hostname: zoekt-webserver-0
        
          # Description: Backend for text search operations.
          #
          # Disk: 128GB / non-persistent SSD
          # Ports exposed to other Sourcegraph services: 3181/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          searcher-0:
            container_name: searcher-0
            image: 'index.docker.io/sourcegraph/searcher:3.12.3@sha256:c31deb621158d1fc0f7d19ca50cca8b2c0aba957736437f485e206fb10367bf7'
            cpus: 2
            mem_limit: '2g'
            environment:
              - GOMAXPROCS=2
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - JAEGER_AGENT_HOST=jaeger-agent
            healthcheck:
              test: "wget -q 'http://127.0.0.1:3181/healthz' -O /dev/null || exit 1"
              interval: 1s
              timeout: 10s
              retries: 1
            volumes:
              - 'searcher-0:/mnt/cache'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Rate-limiting proxy for the GitHub API.
          #
          # CPU: 1
          # Memory: 1GB
          # Disk: 1GB / non-persistent SSD (only for read-only config file)
          # Ports exposed to other Sourcegraph services: 3180/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          github-proxy:
            container_name: github-proxy
            image: 'index.docker.io/sourcegraph/github-proxy:3.12.3@sha256:969d5107c7314e196c7e1ac223b7c65d6bfcfe1e4ccb4a08da67432827b670b2'
            cpus: 1
            mem_limit: '1g'
            environment:
              - GOMAXPROCS=1
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - JAEGER_AGENT_HOST=jaeger-agent
            networks:
              - sourcegraph
            restart: always
        
          # Description: LSIF HTTP server for code intelligence.
          #
          # Disk: 200GB / persistent SSD
          # Ports exposed to other Sourcegraph services: 3186/TCP (server) 3187/TCP (worker)
          # Ports exposed to the public internet: none
          #
          lsif-server:
            container_name: lsif-server
            image: 'index.docker.io/sourcegraph/lsif-server:3.12.3@sha256:078ed33647571a21d6ae3393faf46b976a08aa3c1312b6f5272ec6812b1f0f6e'
            cpus: 2
            mem_limit: '2g'
            environment:
              - GOMAXPROCS=2
              - LSIF_STORAGE_ROOT=/lsif-storage
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
            healthcheck:
              test: "wget -q 'http://127.0.0.1:3186/ping' -O /dev/null || exit 1"
              interval: 5s
              timeout: 5s
              retries: 1
              start_period: 60s
            volumes:
              - 'lsif-server:/lsif-storage'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Saved search query runner / notification service.
          #
          # Disk: 1GB / non-persistent SSD (only for read-only config file)
          # Network: 100mbps
          # Liveness probe: n/a
          # Ports exposed to other Sourcegraph services: 3183/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          query-runner:
            container_name: query-runner
            image: 'index.docker.io/sourcegraph/query-runner:3.12.3@sha256:f0fc1c1b8b19937708687eed6bb2456fcc8433083127f9f9f5ce531edcf39e2b'
            cpus: 1
            mem_limit: '1g'
            environment:
              - GOMAXPROCS=1
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - JAEGER_AGENT_HOST=jaeger-agent
            networks:
              - sourcegraph
            restart: always
        
          # Description: Backend for replace operations.
          #
          # Disk: 128GB / non-persistent SSD
          # Ports exposed to other Sourcegraph services: 3185/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          replacer:
            container_name: replacer
            image: 'index.docker.io/sourcegraph/replacer:3.12.3@sha256:f5a40b66614cbb72602e7c31f46232c16919477800a965f5bc5a568c235e7578'
            cpus: 1
            mem_limit: '512m'
            environment:
              - GOMAXPROCS=1
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
            healthcheck:
              test: "wget -q 'http://127.0.0.1:3185/healthz' -O /dev/null || exit 1"
              interval: 1s
              timeout: 10s
              retries: 1
            volumes:
              - 'replacer:/mnt/cache'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Handles repository metadata (not Git data) lookups and updates from external code hosts and other similar services.
          #
          # Disk: 128GB / non-persistent SSD
          # Ports exposed to other Sourcegraph services: 3182/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          repo-updater:
            container_name: repo-updater
            image: 'index.docker.io/sourcegraph/repo-updater:3.12.3@sha256:1a7a29bebd1d142401051c3c0e8d9b0763d0c7d10df776f8ff342cad66af1d88'
            cpus: 4
            mem_limit: '4g'
            environment:
              - GOMAXPROCS=1
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - JAEGER_AGENT_HOST=jaeger-agent
              - 'GITHUB_BASE_URL=http://github-proxy:3180'
            volumes:
              - 'repo-updater:/mnt/cache'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Backend for syntax highlighting operations.
          #
          # Disk: none
          # Ports exposed to other Sourcegraph services: 9238/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          syntect-server:
            container_name: syntect-server
            image: 'index.docker.io/sourcegraph/syntect_server:2b5a3fb@sha256:ef5529cafdc68d5a21edea472ee8ad966878b173044aa5c3db93bc3d84765b1f'
            cpus: 4
            mem_limit: '6g'
            healthcheck:
              test: "wget -q 'http://127.0.0.1:9238/health' -O /dev/null || exit 1"
              interval: 1s
              timeout: 5s
              retries: 1
              start_period: 5s
            networks:
              - sourcegraph
            restart: always
        
          # Description: Backend for symbols operations.
          #
          # Disk: 128GB / non-persistent SSD
          # Ports exposed to other Sourcegraph services: 3184/TCP 6060/TCP
          # Ports exposed to the public internet: none
          #
          symbols-0:
            container_name: symbols-0
            image: 'index.docker.io/sourcegraph/symbols:3.12.3@sha256:29861594850db42724443e4c58a824696314d7566ec508f6ea85e38f693bf0fb'
            cpus: 2
            mem_limit: '4g'
            environment:
              - GOMAXPROCS=2
              - 'SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090'
              - JAEGER_AGENT_HOST=jaeger-agent
            healthcheck:
              test: "wget -q 'http://127.0.0.1:3184/healthz' -O /dev/null || exit 1"
              interval: 5s
              timeout: 5s
              retries: 1
              start_period: 60s
            volumes:
              - 'symbols-0:/mnt/cache'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Prometheus collects metrics and aggregates them into graphs.
          #
          # Disk: 200GB / persistent SSD
          # Ports exposed to other Sourcegraph services: none
          # Ports exposed to the public internet: none (HTTP 9090 should be exposed to admins only)
          #
          prometheus:
            container_name: prometheus
            image: 'index.docker.io/sourcegraph/prometheus:10.0.7@sha256:22d54f27c7df8733a06c7ae8c2e851b61b1ed42f1f5621d493ef58ebd8d815e0'
            cpus: 4
            mem_limit: '8g'
            volumes:
              - 'prometheus-v2:/prometheus'
              - '../prometheus:/sg_prometheus_add_ons'
            ports:
              - '0.0.0.0:9090:9090'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Dashboards and graphs for Prometheus metrics.
          #
          # Disk: 100GB / persistent SSD
          # Ports exposed to other Sourcegraph services: none
          # Ports exposed to the public internet: none (HTTP 3000 should be exposed to admins only)
          #
          # Add the following environment variables if you wish to use an auth proxy with Grafana:
          #
          # 'GF_AUTH_PROXY_ENABLED=true'
          # 'GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User'
          # 'GF_SERVER_ROOT_URL='https://grafana.example.com'
          grafana:
            container_name: grafana
            image: 'index.docker.io/sourcegraph/grafana:10.0.9@sha256:0132e5602030145803753468497a2d17640164b9c34df4ce2532dd93e4b1f6fc'
            cpus: 1
            mem_limit: '1g'
            volumes:
              - 'grafana:/var/lib/grafana'
              - '../grafana/datasources:/sg_config_grafana/provisioning/datasources'
              - '../grafana/dashboards:/sg_grafana_additional_dashboards'
            ports:
              - '0.0.0.0:3370:3370'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Publishes Prometheus metrics about Docker containers.
          #
          # Disk: none
          # Ports exposed to other Sourcegraph services: 8080/TCP
          # Ports exposed to the public internet: none
          #
          cadvisor:
            container_name: cadvisor
            image: 'google/cadvisor:v0.33.0'
            cpus: 1
            mem_limit: '1g'
            volumes:
              - '/:/rootfs:ro'
              - '/var/run:/var/run:ro'
              - '/sys:/sys:ro'
              - '/var/lib/docker/:/var/lib/docker:ro'
              - '/dev/disk/:/dev/disk:ro'
            networks:
              - sourcegraph
            restart: always
        
          # # Description: Jaeger agent which is local to the host machine (containers on
          # # the machine send trace information to it and it relays to the collector).
          # #
          # # Disk: none
          # # Ports exposed to other Sourcegraph services: 5775/UDP 6831/UDP 6832/UDP (on the same host machine)
          # # Ports exposed to the public internet: none
          # #
          # jaeger-agent:
          #   container_name: jaeger-agent
          #   image: 'index.docker.io/jaegertracing/jaeger-agent@sha256:7ad33c19fd66307f2a3c07c95eb07c335ddce1b487f6b6128faa75d042c496cb'
          #   cpus: 1
          #   mem_limit: '1g'
          #   environment:
          #     - "COLLECTOR_HOST_PORT='jaeger-collector:14267'"
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Jaeger's Cassandra database for storing traces.
          # #
          # # Disk: 128GB / persistent SSD
          # # Ports exposed to other Sourcegraph services: 9042/TCP
          # # Ports exposed to the public internet: none
          # #
          # jaeger-cassandra:
          #   container_name: jaeger-cassandra
          #   image: 'index.docker.io/library/cassandra:3.11.4@sha256:9f1d47fd23261c49f226546fe0134e6d4ad0570b7ea3a169c521005cb8369a32'
          #   cpus: 4
          #   mem_limit: '8g'
          #   environment:
          #     - 'HEAP_NEWSIZE=1G'
          #     - 'MAX_HEAP_SIZE=6G'
          #     - 'CASSANDRA_DC=sourcegraph'
          #     - 'CASSANDRA_RACK=rack1'
          #     - 'CASSANDRA_ENDPOINT_SNITCH=GossipingPropertyFileSnitch'
          #   volumes:
          #     - 'jaeger-cassandra:/var/lib/cassandra'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Receives traces from Jaeger agents.
          # #
          # # Disk: none
          # # Ports exposed to other Sourcegraph services: 14267/TCP
          # # Ports exposed to the public internet: none
          # #
          # jaeger-collector:
          #   container_name: jaeger-collector
          #   image: 'index.docker.io/jaegertracing/jaeger-collector:1.11@sha256:0b6d28bb52410f7b50c0f0fc16d7ee391e2e3eca47b713ac88d0891ca8a63cb9'
          #   cpus: 1
          #   mem_limit: '1g'
          #   environment:
          #     - SPAN_STORAGE_TYPE=cassandra
          #     - CASSANDRA_SERVERS=jaeger-cassandra
          #     - CASSANDRA_KEYSPACE=jaeger_v1_sourcegraph
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # # Description: Jaeger frontend for querying traces.
          # #
          # # Disk: none
          # # Ports exposed to other Sourcegraph services: none
          # # Ports exposed to the public internet: none (HTTP 16686 should be exposed to admins only)
          # #
          # jaeger-query:
          #   container_name: jaeger-query
          #   image: 'index.docker.io/jaegertracing/jaeger-query:1.11@sha256:cddc521d0166c868931282685a863368ae2c14c4de0c1be38e388ece3080439e'
          #   cpus: 1
          #   mem_limit: '1g'
          #   environment:
          #     - SPAN_STORAGE_TYPE=cassandra
          #     - CASSANDRA_SERVERS=jaeger-cassandra
          #     - CASSANDRA_KEYSPACE=jaeger_v1_sourcegraph
          #     - CASSANDRA_LOCAL_DC=sourcegraph
          #   ports:
          #     - '0.0.0.0:16686:16686'
          #   networks:
          #     - sourcegraph
          #   restart: always
        
          # Description: PostgreSQL database for various data.
          #
          # Disk: 128GB / persistent SSD
          # Ports exposed to other Sourcegraph services: 5432/TCP 9187/TCP
          # Ports exposed to the public internet: none
          #
          pgsql:
            container_name: pgsql
            image: 'index.docker.io/sourcegraph/postgres-11.4:19-11-14_b084311b@sha256:072481559d559cfd9a53ad77c3688b5cf583117457fd452ae238a20405923297'
            cpus: 4
            mem_limit: '2g'
            healthcheck:
              test: '/liveness.sh'
              interval: 10s
              timeout: 1s
              retries: 3
              start_period: 15s
            volumes:
              - 'pgsql:/data/'
            networks:
              - sourcegraph
            restart: always
        
          # Description: Redis for storing short-lived caches.
          #
          # Disk: 128GB / persistent SSD
          # Ports exposed to other Sourcegraph services: 6379/TCP 9121/TCP
          # Ports exposed to the public internet: none
          #
          redis-cache:
            container_name: redis-cache
            image: index.docker.io/sourcegraph/redis-cache:19-04-16_6891de82@sha256:4cbfac8af0abb673899250d4fd859cc477d6426de519e9deb71e454e18322499
            cpus: 1
            mem_limit: '6g'
            volumes:
              - 'redis-cache:/redis-data'
            networks:
              - sourcegraph
            restart: always
          # Description: Redis for storing semi-persistent data like user sessions.
          #
          # Disk: 128GB / persistent SSD
          # Ports exposed to other Sourcegraph services: 6379/TCP 9121/TCP
          # Ports exposed to the public internet: none
          #
          redis-store:
            container_name: redis-store
            image: 'index.docker.io/sourcegraph/redis-store:19-04-16_6891de821@sha256:56426d601ce1f6d63088fea1cefa61f69a2e809c7d90fc1d157cca63cf81b277'
            cpus: 1
            mem_limit: '6g'
            volumes:
              - 'redis-store:/redis-data'
            networks:
              - sourcegraph
            restart: always
        
        volumes:
          gitserver-0:
          grafana:
          #jaeger-cassandra:
          lsif-server:
          pgsql:
          prometheus-v2:
          redis-cache:
          redis-store:
          replacer:
          repo-updater:
          searcher-0:
          sourcegraph-frontend-0:
          sourcegraph-frontend-internal-0:
          symbols-0:
          zoekt-0-shared:
        networks:
          sourcegraph:

1. Bring up all the other services 

        > docker-compose up -d
