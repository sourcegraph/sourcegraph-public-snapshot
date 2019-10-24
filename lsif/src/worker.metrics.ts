import promClient from 'prom-client'

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
