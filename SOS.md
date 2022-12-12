In one tab, run this to proxy access to the codeintel-db clone on your local machine.

```
./google-cloud-sdk/cloud_sql_proxy -instances=sourcegraph-dev:us-central1:eric-fritz-codeintel-db-lsif-scip-conversion-test=tcp:5333
```

Then build and run the migrator to see it churn through records.

```
go build ./enterprise/cmd/migrator

env CODEINTEL_PGHOST=127.0.0.1 \
    CODEINTEL_PGPORT=5333 \
    CODEINTEL_PGDATABASE=sg \
    CODEINTEL_PGUSER=dev \
    CODEINTEL_PGPASSWORD='hunter2' \
    SRC_LOG_LEVEL=debug \
    ./migrator run-out-of-band-migrations -id 19
```
