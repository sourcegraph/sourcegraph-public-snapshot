import promClient from 'prom-client'

//
// HTTP Metrics

export const httpUploadDurationHistogram = new promClient.Histogram({
    name: 'lsif_http_upload_request_duration_seconds',
    help: 'Total time spent on upload requests.',
    labelNames: ['code'],
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const httpQueryDurationHistogram = new promClient.Histogram({
    name: 'lsif_http_query_request_duration_seconds',
    help: 'Total time spent on query requests.',
    labelNames: ['code'],
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

//
// Job and Queue Metrics

export const queueSizeGauge = new promClient.Gauge({
    name: 'lsif_queue_size',
    help: 'The current number of items in the work-queue.',
})
