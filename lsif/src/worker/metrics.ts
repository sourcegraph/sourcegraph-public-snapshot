import promClient from 'prom-client'

//
// Upload Conversion Metrics

export const uploadConversionDurationHistogram = new promClient.Histogram({
    name: 'lsif_upload_conversion_duration_seconds',
    help: 'Total time spent on upload conversions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const uploadConversionDurationErrorsCounter = new promClient.Counter({
    name: 'lsif_upload_conversion_errors_total',
    help: 'The number of errors that occurred while converting an LSIF upload.',
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
