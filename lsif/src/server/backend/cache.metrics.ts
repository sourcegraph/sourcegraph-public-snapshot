import promClient from 'prom-client'

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
