import promClient from 'prom-client'

//
// Cache Metrics

export const CONNECTION_CACHE_SIZE_GAUGE = new promClient.Gauge({
    name: 'lsif_connection_cache_size',
    help: 'The current number of open SQLite handles.',
})

export const CONNECTION_CACHE_EVENTS_COUNTER = new promClient.Counter({
    name: 'lsif_connection_cache_events_total',
    help: 'The number of connection cache hits, misses, and evictions.',
    labelNames: ['type'],
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

//
// Bloom Filter Metrics

export const BLOOM_FILTER_EVENTS_COUNTER = new promClient.Counter({
    name: 'lsif_bloom_events_counter',
    help: 'The number of bloom filter hits and misses.',
    labelNames: ['type'],
})

//
// Database Metrics

export const DATABASE_QUERY_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_database_query_duration_seconds',
    help: 'Total time spent on database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const XREPO_DATABASE_QUERY_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_xrepo_query_duration_seconds',
    help: 'Total time spent on cross-repo database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

//
// Database Insertion Metrics

export const DATABASE_INSERTION_COUNTER = new promClient.Counter({
    name: 'lsif_database_insertions',
    help: 'The number of insertions into a database.',
})

export const DATABASE_INSERTION_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_database_insertion_duration_seconds',
    help: 'Total time spent on database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const XREPO_DATABASE_INSERTION_COUNTER = new promClient.Counter({
    name: 'lsif_xrepo_insertions',
    help: 'The number of insertions into a cross-repo database.',
})

export const XREPO_DATABASE_INSERTION_DURATION_HISTOGRAM = new promClient.Histogram({
    name: 'lsif_xrepo_insertion_duration_seconds',
    help: 'Total time spent on cross-repo database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})
