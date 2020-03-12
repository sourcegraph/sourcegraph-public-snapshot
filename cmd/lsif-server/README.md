# lsif-server

A small wrapper around the TypeScript/node processes that serve precise LSIF code intelligence data.

- [lsif-server](../../lsif/src/server/server.ts)
- [lsif-worker](../../lsif/src/worker/worker.ts)

These processes are run in a [goreman](https://github.com/mattn/goreman) supervisor. By default, there will be one server process and one worker process. The number of replicas per process can be tuned with the environment variables `LSIF_NUM_SERVERS` (zero or one) and `LSIF_NUM_WORKERS` (zero or more).

## Prometheus metrics

The lsif-server process exposes an HTTP API on port 3186. This API contains a `/metrics` endpoint to be scraped by prometheus. The lsif-worker process exposes a metrics server (but nothing else interesting) on port 3187. It's possible to run multiple worker processes, but impossible for them all to serve metrics from the same port. Therefore, this container also contains a minimally-configured Prometheus process that will scrape metrics from the server and worker processes. It is suggested that you use [federation](https://prometheus.io/docs/prometheus/latest/federation/) to scrape all of the process metrics at once instead of scraping the server and worker(s) ports directly. Doing so will ensure that scaling up or down the number of workers will not change the the required Prometheus configuration.
