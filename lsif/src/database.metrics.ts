import promClient from 'prom-client'

export const databaseQueryDurationHistogram = new promClient.Histogram({
    name: 'lsif_database_query_duration_seconds',
    help: 'Total time spent on database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const databaseQueryErrorsCounter = new promClient.Counter({
    name: 'lsif_database_query_errors_total',
    help: 'The number of errors that occurred during a database query.',
})
