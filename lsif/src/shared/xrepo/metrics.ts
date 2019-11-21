import promClient from 'prom-client'

//
// Insertion Metrics

export const xrepoInsertionDurationHistogram = new promClient.Histogram({
    name: 'lsif_xrepo_insertion_duration_seconds',
    help: 'Total time spent on cross-repo database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const xrepoInsertionErrorsCounter = new promClient.Counter({
    name: 'lsif_xrepo_insertion_errors_total',
    help: 'The number of errors that occurred during a cross-repo database insertion.',
})

//
// Bloom Filter Metrics

export const bloomFilterEventsCounter = new promClient.Counter({
    name: 'lsif_bloom_filter_events_total',
    help: 'The number of bloom filter hits and misses.',
    labelNames: ['type'],
})

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
