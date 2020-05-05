import promClient from 'prom-client'

//
// Bloom Filter Metrics

export const bloomFilterEventsCounter = new promClient.Counter({
    name: 'lsif_bloom_filter_events_total',
    help: 'The number of bloom filter hits and misses.',
    labelNames: ['type'],
})
