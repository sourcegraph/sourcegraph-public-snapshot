import promClient from 'prom-client'

//
// Database Metrics

export const DATABASE_QUERY_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_database_query_duration_seconds',
    help: 'Total time spent on database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const XREPO_QUERY_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_xrepo_query_duration_seconds',
    help: 'Total time spent on cross-repo database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const DATABASE_QUERY_ERRORS_COUNTER = new promClient.Counter({
    name: 'lsif_database_query_errors_total',
    help: 'The number of errors that occurred during a database query.',
})

export const XREPO_QUERY_ERRORS_COUNTER = new promClient.Counter({
    name: 'lsif_xrepo_query_errors_total',
    help: 'The number of errors that occurred during a cross-repo database query.',
})

//
// Database Insertion Metrics

export const DATABASE_INSERTION_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_database_insertion_duration_seconds',
    help: 'Total time spent on database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const XREPO_INSERTION_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_xrepo_insertion_duration_seconds',
    help: 'Total time spent on cross-repo database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const DATABASE_INSERTION_ERRORS_COUNTER = new promClient.Counter({
    name: 'lsif_database_insertion_errors_total',
    help: 'The number of errors that occurred during a database insertion.',
})

export const XREPO_INSERTION_ERRORS_COUNTER = new promClient.Counter({
    name: 'lsif_xrepo_insertion_errors_total',
    help: 'The number of errors that occurred during a cross-repo database insertion.',
})

//
// Cache Metrics

export const CONNECTION_CACHE_CAPACITY_GAUGE = new promClient.Gauge({
    name: 'lsif_connection_cache_capacity',
    help: 'The maximum number of open SQLite handles.',
})

export const CONNECTION_CACHE_SIZE_GAUGE = new promClient.Gauge({
    name: 'lsif_connection_cache_size',
    help: 'The current number of open SQLite handles.',
})

export const CONNECTION_CACHE_EVENTS_COUNTER = new promClient.Counter({
    name: 'lsif_connection_cache_events_total',
    help: 'The number of connection cache hits, misses, and evictions.',
    labelNames: ['type'],
})

export const DOCUMENT_CACHE_CAPACITY_GAUGE = new promClient.Gauge({
    name: 'lsif_document_cache_capacity',
    help: 'The maximum number of documents loaded in memory.',
})

export const DOCUMENT_CACHE_SIZE_GAUGE = new promClient.Gauge({
    name: 'lsif_document_cache_size',
    help: 'The current number of documents loaded in memory.',
})

export const DOCUMENT_CACHE_EVENTS_COUNTER = new promClient.Counter({
    name: 'lsif_document_cache_events_total',
    help: 'The number of document cache hits, misses, and evictions.',
    labelNames: ['type'],
})

export const RESULT_CHUNK_CACHE_CAPACITY_GAUGE = new promClient.Gauge({
    name: 'lsif_results_chunk_cache_capacity',
    help: 'The maximum number of result chunks loaded in memory.',
})

export const RESULT_CHUNK_CACHE_SIZE_GAUGE = new promClient.Gauge({
    name: 'lsif_results_chunk_cache_size',
    help: 'The current number of result chunks loaded in memory.',
})

export const RESULT_CHUNK_CACHE_EVENTS_COUNTER = new promClient.Counter({
    name: 'lsif_results_chunk_cache_events_total',
    help: 'The number of result chunk cache hits, misses, and evictions.',
    labelNames: ['type'],
})

//
// Bloom Filter Metrics

export const BLOOM_FILTER_EVENTS_COUNTER = new promClient.Counter({
    name: 'lsif_bloom_filter_events_total',
    help: 'The number of bloom filter hits and misses.',
    labelNames: ['type'],
})

//
// Helpers

export async function instrument<T>(
    durationHistogram: promClient.Histogram,
    errorsCounter: promClient.Counter,
    fn: () => Promise<T>
): Promise<T> {
    const end = durationHistogram.startTimer()
    try {
        return await fn()
    } catch (e) {
        errorsCounter.inc()
        throw e
    } finally {
        end()
    }
}
