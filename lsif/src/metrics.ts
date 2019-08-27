import promClient from 'prom-client'

export const ConnectionCacheHitCounter = new promClient.Counter({
    name: 'lsif_connection_cache_hit',
    help: 'The number of connection cache hits and misses.',
    labelNames: ['type'],
})

export const ConnectionCacheEvictionCounter = new promClient.Counter({
    name: 'lsif_connection_cache_eviction',
    help: 'The number of connection cache entry evictions and failed evictions due to locked entries.',
    labelNames: ['type'],
})

export const ConnectionCacheSizeGauge = new promClient.Gauge({
    name: 'lsif_connection_cache_size',
    help: 'The current number of open SQLite handles.',
})

export const DocumentCacheHitCounter = new promClient.Counter({
    name: 'lsif_document_cache_hit',
    help: 'The number of document cache hits and misses.',
    labelNames: ['type'],
})

export const DocumentCacheEvictionCounter = new promClient.Counter({
    name: 'lsif_document_cache_eviction',
    help: 'The number of document cache entry evictions and failed evictions due to locked entries.',
    labelNames: ['type'],
})

export const DocumentCacheSizeGauge = new promClient.Gauge({
    name: 'lsif_document_cache_size',
    help: 'The current number of bytes reserved for document blos.',
})

export const databaseQueryDurationHistogram = new promClient.Histogram({
    name: 'lsif_database_query_duration_seconds',
    help: 'Total time spent on database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const xrepoQueryDurationHistogram = new promClient.Histogram({
    name: 'lsif_xrepo_query_duration_seconds',
    help: 'Total time spent on cross-repo database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const databaseInsertionCounter = new promClient.Counter({
    name: 'lsif_database_insertions',
    help: 'The number of insertions into a database.',
})

export const databaseInsertionDurationHistogram = new promClient.Histogram({
    name: 'lsif_database_insertion_duration_seconds',
    help: 'Total time spent on database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const xrepoInsertionCounter = new promClient.Counter({
    name: 'lsif_xrepo_insertions',
    help: 'The number of insertions into a cross-repo database.',
})

export const xrepoInsertionDurationHistogram = new promClient.Histogram({
    name: 'lsif_xrepo_insertion_duration_seconds',
    help: 'Total time spent on cross-repo database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const bloomFilterHitCounter = new promClient.Counter({
    name: 'lsif_bloom_filter_hit',
    help: 'The number of bloom filter hits and misses.',
    labelNames: ['type'],
})
