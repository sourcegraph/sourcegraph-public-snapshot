import promClient from 'prom-client'

//
// Job Metrics

export const jobDurationHistogram = new promClient.Histogram({
    name: 'lsif_job_duration_seconds',
    help: 'Total time spent on jobs.',
    labelNames: ['class'],
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const jobDurationErrorsCounter = new promClient.Counter({
    name: 'lsif_job_errors_total',
    help: 'The number of errors that occurred while processing a job.',
})

//
// Importer Metric

export const databaseInsertionDurationHistogram = new promClient.Histogram({
    name: 'lsif_database_insertion_duration_seconds',
    help: 'Total time spent on database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const databaseInsertionErrorsCounter = new promClient.Counter({
    name: 'lsif_database_insertion_errors_total',
    help: 'The number of errors that occurred during a database insertion.',
})
