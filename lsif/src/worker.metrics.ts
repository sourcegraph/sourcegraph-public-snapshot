import promClient from 'prom-client'

export const jobEventsCounter = new promClient.Counter({
    name: 'lsif_job_events_total',
    help: 'The total number of jobs success and failures.',
    labelNames: ['class', 'type'],
})

export const jobDurationHistogram = new promClient.Histogram({
    name: 'lsif_job_duration_seconds',
    help: 'Total time spent on jobs.',
    labelNames: ['class'],
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})
