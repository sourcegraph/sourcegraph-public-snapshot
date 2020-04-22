import promClient from 'prom-client'

//
// Query Metrics

export const postgresQueryDurationHistogram = new promClient.Histogram({
    name: 'lsif_xrepo_query_duration_seconds',
    help: 'Total time spent on Postgres database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const postgresQueryErrorsCounter = new promClient.Counter({
    name: 'lsif_xrepo_query_errors_total',
    help: 'The number of errors that occurred during a Postgres database query.',
})

//
// Insertion Metrics

export const postgresInsertionDurationHistogram = new promClient.Histogram({
    name: 'lsif_xrepo_insertion_duration_seconds',
    help: 'Total time spent on Postgres database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const postgresInsertionErrorsCounter = new promClient.Counter({
    name: 'lsif_xrepo_insertion_errors_total',
    help: 'The number of errors that occurred during a Postgres database insertion.',
})
