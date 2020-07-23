# Using external databases with Sourcegraph

> NOTE: Using Sourcegraph with an external database is a [paid feature](https://about.sourcegraph.com/pricing). [Contact us](https://about.sourcegraph.com/contact/sales) to get a trial license.

Sourcegraph by default provides its own PostgreSQL and Redis databases for data storage:

- PostgreSQL for storing long-term information, such as user information when using Sourcegraph's built-in authentication provider instead of an external one.
- Redis for storing short-term information, such as session information and cache data.

## Using your own PostgreSQL server

You can use your own PostgreSQL v9.6+ server with Sourcegraph if you wish. For example, you may prefer this if you already have existing backup infrastructure around your own PostgreSQL server, wish to use Amazon RDS, etc.

Simply add the standard PostgreSQL environment variables to your Sourcegraph deployment files and Sourcegraph will use that PostgreSQL server instead of its built-in one.

### sourcegraph/server

Add the following to your `docker run` command:

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->
<pre class="pre-wrap start-sourcegraph-command"><code>docker run [...]<span class="virtual-br"></span> -e PGHOST=psql.mycompany.org<span class="virtual-br"></span> -e PGUSER=sourcegraph<span class="virtual-br"></span> -e PGPASSWORD=secret<span class="virtual-br"></span> -e PGDATABASE=sourcegraph<span class="virtual-br"></span> -e PGSSLMODE=require<span class="virtual-br"></span> sourcegraph/server:3.18.0</code></pre>

### Docker Compose

1. Add/modify the following environment variables to all of the `sourcegraph-frontend-*` services and the `sourcegraph-frontend-internal` service in ` [docker-compose.yaml](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/v3.14.2/docker-compose/docker-compose.yaml): 

    ```
    sourcegraph-frontend-0:
      # ...
      environment:
        # ...
        - 'PGHOST=psql.mycompany.org'
        - 'PGUSER=sourcegraph'
        - 'PGPASSWORD=secret'
        - 'PGDATABASE=sourcegraph'
        - 'PGSSLMODE=require'
      # ...
    ```

    See ["Environment variables in Compose"](https://docs.docker.com/compose/environment-variables/) for other ways to pass these environment variables to the relevant services (including from the command line, a `.env` file, etc.).

1. Comment out / remove the internal `pgsql` service in [docker-compose.yaml](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/v3.14.2/docker-compose/docker-compose.yaml) since Sourcegraph is using the external one now.

    ```
    # # Description: PostgreSQL database for various data.
    # #
    # # Disk: 128GB / persistent SSD
    # # Ports exposed to other Sourcegraph services: 5432/TCP 9187/TCP
    # # Ports exposed to the public internet: none
    # #
    # pgsql:
    # container_name: pgsql
    # image: 'index.docker.io/sourcegraph/postgres-11.4:19-11-14_b084311b@sha256:072481559d559cfd9a53ad77c3688b5cf583117457fd452ae238a20405923297'
    # cpus: 4
    # mem_limit: '2g'
    # healthcheck:
    #    test: '/liveness.sh'
    #    interval: 10s
    #    timeout: 1s
    #    retries: 3
    #    start_period: 15s
    # volumes:
    #    - 'pgsql:/data/'
    # networks:
    #     - sourcegraph
    # restart: always
    ```

### Kubernetes

Update the `PG*` environment variables in the `sourcegraph-frontend` deployment YAML file to point to the external PostgreSQL instance.

### Version requirements

Please refer to our [Postgres](https://docs.sourcegraph.com/admin/postgres) documentation to learn about version requirements.

### Caveats

> NOTE: If your PostgreSQL server does not support SSL, set `PGSSLMODE=disable` instead of `PGSSLMODE=require`. Note that this is potentially insecure.

Most standard PostgreSQL environment variables may be specified (`PGPORT`, etc). See http://www.postgresql.org/docs/current/static/libpq-envars.html for a full list.

> NOTE: On Mac/Windows, if trying to connect to a PostgreSQL server on the same host machine, remember that Sourcegraph is running inside a Docker container inside of the Docker virtual machine. You may need to specify your actual machine IP address and not `localhost` or `127.0.0.1` as that refers to the Docker VM itself.

## Using your own Redis server

**Version requirements**: We support any version *starting from 5.0*.

Generally, there is no reason to do this as Sourcegraph only stores ephemeral cache and session data in Redis. However, if you want to use an external Redis server with Sourcegraph, you can do the following:

Add the `REDIS_ENDPOINT` environment variable to your `docker run` command and Sourcegraph will use that Redis server instead of its built-in one. The string must either have the format `$HOST:PORT`
or follow the [IANA specification for Redis URLs](https://www.iana.org/assignments/uri-schemes/prov/redis) (e.g., `redis://:mypassword@host:6379/2`). For example:

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->
<pre class="pre-wrap"><code>docker run [...]<span class="virtual-br"></span>   -e REDIS_ENDPOINT=redis.mycompany.org:6379<span class="virtual-br"></span>   sourcegraph/server:3.18.0</code></pre>

> NOTE: On Mac/Windows, if trying to connect to a Redis server on the same host machine, remember that Sourcegraph is running inside a Docker container inside of the Docker virtual machine. You may need to specify your actual machine IP address and not `localhost` or `127.0.0.1` as that refers to the Docker VM itself.

If using Docker for Desktop, `host.docker.internal` will resolve to the host IP address.
