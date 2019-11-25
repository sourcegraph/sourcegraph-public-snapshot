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
