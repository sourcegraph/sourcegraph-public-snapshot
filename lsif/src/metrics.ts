import promClient from 'prom-client'

export const ConnectionCacheHitCounter = new promClient.Counter({
    name: 'connection_cache_hit',
    help: 'The number of connection cache hits and misses.',
    labelNames: ['type'],
})

export const ConnectionCacheEvictionCounter = new promClient.Counter({
    name: 'connection_cache_eviction',
    help: 'The number of connection cache entry evictions and failed evictions due to locked entries.',
    labelNames: ['type'],
})

export const ConnectionCacheSizeGauge = new promClient.Gauge({
    name: 'connection_cache_size',
    help: 'The current number of open SQLite handles.',
})

export const DocumentCacheHitCounter = new promClient.Counter({
    name: 'document_cache_hit',
    help: 'The number of document cache hits and misses.',
    labelNames: ['type'],
})

export const DocumentCacheEvictionCounter = new promClient.Counter({
    name: 'document_cache_eviction',
    help: 'The number of document cache entry evictions and failed evictions due to locked entries.',
    labelNames: ['type'],
})

export const DocumentCacheSizeGauge = new promClient.Gauge({
    name: 'document_cache_size',
    help: 'The current number of bytes reserved for document blos.',
})

export const databaseQueryDurationHistogram = new promClient.Histogram({
    name: 'database_query_duration_seconds',
    help: 'Total time spent on database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const xrepoQueryDurationHistogram = new promClient.Histogram({
    name: 'xrepo_query_duration_seconds',
    help: 'Total time spent on cross-repo database queries.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const databaseInsertionCounter = new promClient.Counter({
    name: 'database_insertions',
    help: 'The number of insertions into a database.',
})

export const databaseInsertionDurationHistogram = new promClient.Histogram({
    name: 'database_insertion_duration_seconds',
    help: 'Total time spent on database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const xrepoInsertionCounter = new promClient.Counter({
    name: 'xrepo_insertions',
    help: 'The number of insertions into a cross-repo database.',
})

export const xrepoInsertionDurationHistogram = new promClient.Histogram({
    name: 'xrepo_insertion_duration_seconds',
    help: 'Total time spent on cross-repo database insertions.',
    buckets: [0.2, 0.5, 1, 2, 5, 10, 30],
})

export const bloomFilterHitCounter = new promClient.Counter({
    name: 'bloom_filter_hit',
    help: 'The number of bloom filter hits and misses.',
    labelNames: ['type'],
})
