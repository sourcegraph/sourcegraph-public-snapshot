# Precise code intelligence worker

The docker image for this part of the application wraps a one or more workers [goreman](https://github.com/mattn/goreman) supervisor. By default, there is are two worker processes. The number of workers can be tuned with the environment variable `NUM_WORKERS`.

### Prometheus metrics

The precise-code-intel-worker exposes a metrics server (but nothing else interesting) on port 3188. It's possible to run multiple workers, but impossible for them all to serve metrics from the same port. Therefore, this container also includes a minimally-configured Prometheus process that will scrape metrics from all of the processes. It is suggested that you use [federation](https://prometheus.io/docs/prometheus/latest/federation/) to scrape all of the process metrics at once instead of scraping the individual ports directly. Doing so will ensure that scaling up or down the number of workers will not change the the required Prometheus configuration.
