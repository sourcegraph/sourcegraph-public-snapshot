import promClient from 'prom-client'

//
// Gitserver Metrics

export const gitserverDurationHistogram = new promClient.Histogram({
    name: 'lsif_gitserver_duration_seconds',
    help: 'Total time spent on gitserver exec queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const gitserverErrorsCounter = new promClient.Counter({
    name: 'lsif_gitserver_errors_total',
    help: 'The number of errors that occurred during a gitserver exec query.',
})
