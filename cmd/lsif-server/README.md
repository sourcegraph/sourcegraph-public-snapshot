# lsif-server

A small wrapper around the TypeScript/node processes that serve precise LSIF code intelligence data.

- [lsif-server](../../lsif/src/server/server.ts)
- [lsif-dump-manager](../../lsif/src/dump-manager/dump-manager.ts)
- [lsif-dump-processor](../../lsif/src/dump-processor/dump-processor.ts)

These processes are run in a [goreman](https://github.com/mattn/goreman) supervisor. By default, there will be one server process and one worker process. The number of replicas per process can be tuned with the environment variables `LSIF_NUM_SERVERS` (zero or one), `LSIF_NUM_DUMP_MANAGERS` (zero or one), and `LSIF_NUM_DUMP_PROCESSORS` (zero or more).

## Prometheus metrics

The lsif-server and lsif-dump-manager expose HTTP APIs on ports 3186 and 3187, respectively. These APIs contain a `/metrics` endpoint to be scraped by Prometheus. The lsif-dump-processor exposes a metrics server (but nothing else interesting) on port 3188. It's possible to run multiple processors, but impossible for them all to serve metrics from the same port. Therefore, this container also includes a minimally-configured Prometheus process that will scrape metrics from all of the processes. It is suggested that you use [federation](https://prometheus.io/docs/prometheus/latest/federation/) to scrape all of the process metrics at once instead of scraping the individual ports directly. Doing so will ensure that scaling up or down the number of processors will not change the the required Prometheus configuration.
