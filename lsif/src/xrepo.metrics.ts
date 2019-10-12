import promClient from 'prom-client'

//
// Query Metrics

export const xrepoQueryDurationHistogram = new promClient.Histogram({
    name: 'lsif_xrepo_query_duration_seconds',
    help: 'Total time spent on cross-repo database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})
export const xrepoQueryErrorsCounter = new promClient.Counter({
    name: 'lsif_xrepo_query_errors_total',
    help: 'The number of errors that occurred during a cross-repo database query.',
})

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
