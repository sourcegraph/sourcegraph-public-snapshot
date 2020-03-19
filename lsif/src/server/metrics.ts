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
// Unconverted Upload Metrics

export const unconvertedUploadSizeGauge = new promClient.Gauge({
    name: 'lsif_unconverted_upload_size',
    help: 'The current number of uploads that have are pending conversion.',
})
