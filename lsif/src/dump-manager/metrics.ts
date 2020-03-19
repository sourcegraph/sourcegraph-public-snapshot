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
// Database Metrics

export const databaseQueryDurationHistogram = new promClient.Histogram({
    name: 'lsif_database_query_duration_seconds',
    help: 'Total time spent on database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const databaseQueryErrorsCounter = new promClient.Counter({
    name: 'lsif_database_query_errors_total',
    help: 'The number of errors that occurred during a database query.',
})

//
// Unconverted Upload Metrics

export const unconvertedUploadSizeGauge = new promClient.Gauge({
    name: 'lsif_unconverted_upload_size',
    help: 'The current number of uploads that have are pending conversion.',
})

//
// Cache Metrics

export const connectionCacheCapacityGauge = new promClient.Gauge({
    name: 'lsif_connection_cache_capacity',
    help: 'The maximum number of open SQLite handles.',
})

export const connectionCacheSizeGauge = new promClient.Gauge({
    name: 'lsif_connection_cache_size',
    help: 'The current number of open SQLite handles.',
})

export const connectionCacheEventsCounter = new promClient.Counter({
    name: 'lsif_connection_cache_events_total',
    help: 'The number of connection cache hits, misses, and evictions.',
    labelNames: ['type'],
})

export const documentCacheCapacityGauge = new promClient.Gauge({
    name: 'lsif_document_cache_capacity',
    help: 'The maximum number of documents loaded in memory.',
})

export const documentCacheSizeGauge = new promClient.Gauge({
    name: 'lsif_document_cache_size',
    help: 'The current number of documents loaded in memory.',
})

export const documentCacheEventsCounter = new promClient.Counter({
    name: 'lsif_document_cache_events_total',
    help: 'The number of document cache hits, misses, and evictions.',
    labelNames: ['type'],
})

export const resultChunkCacheCapacityGauge = new promClient.Gauge({
    name: 'lsif_results_chunk_cache_capacity',
    help: 'The maximum number of result chunks loaded in memory.',
})

export const resultChunkCacheSizeGauge = new promClient.Gauge({
    name: 'lsif_results_chunk_cache_size',
    help: 'The current number of result chunks loaded in memory.',
})

export const resultChunkCacheEventsCounter = new promClient.Counter({
    name: 'lsif_results_chunk_cache_events_total',
    help: 'The number of result chunk cache hits, misses, and evictions.',
    labelNames: ['type'],
})
