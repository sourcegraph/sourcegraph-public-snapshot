import * as Prometheus from 'prom-client'
import { CacheStats, InsertStats, QueryStats, CreateRunnerStats } from './stats'

/**
 * The default bucket configuration for Prometheus histograms.
 */
const DEFAULT_HISTOGRAM_BUCKETS = [0.1, 5, 15, 50, 100, 200, 300, 400, 500]

/**
 * The set of reporters created by createPrometheusReporters.
 */
export interface PrometheusReporters {
    insertDumpDuration: Prometheus.Histogram
    cacheDuration: Prometheus.Histogram
    createRunnerDuration: Prometheus.Histogram
    queryDuration: Prometheus.Histogram
    cacheHits: Prometheus.Counter
    cacheMisses: Prometheus.Counter
}

/**
 * Create the Prometheus objects that are exposed to the /metrics
 * endpoint. These are made available in http-server.
 */
export function createPrometheusReporters(): PrometheusReporters {
    const insertDumpDuration = new Prometheus.Histogram({
        name: 'insert_dump_duration_ms',
        help: 'Duration of LSIF dump insertions in ms',
        buckets: DEFAULT_HISTOGRAM_BUCKETS,
    })

    const cacheDuration = new Prometheus.Histogram({
        name: 'cache_duration_ms',
        help: 'Duration of cache lookups in ms',
        buckets: DEFAULT_HISTOGRAM_BUCKETS,
    })

    const createRunnerDuration = new Prometheus.Histogram({
        name: 'create_runner_duration_ms',
        help: 'Duration of query runner creation in ms',
        buckets: DEFAULT_HISTOGRAM_BUCKETS,
    })

    const queryDuration = new Prometheus.Histogram({
        name: 'query_duration_ms',
        help: 'Duration of query invocations in ms',
        buckets: DEFAULT_HISTOGRAM_BUCKETS,
    })

    const cacheHits = new Prometheus.Counter({
        name: 'cache_hits',
        help: 'The number of fast-path cache lookups',
    })

    const cacheMisses = new Prometheus.Counter({
        name: 'cache_misses',
        help: 'The number of fast-path cache lookups',
    })

    return { insertDumpDuration, cacheDuration, createRunnerDuration, queryDuration, cacheHits, cacheMisses }
}

/**
 * Emit backend and query runner stat payloads into Prometheus.
 */
export function emit(reporters: PrometheusReporters, stats: CacheStats): void
export function emit(reporters: PrometheusReporters, stats: InsertStats): void
export function emit(reporters: PrometheusReporters, stats: CreateRunnerStats): void
export function emit(reporters: PrometheusReporters, stats: QueryStats): void {
    if (isCacheStats(stats)) {
        reporters.cacheDuration.observe(stats.elapsedMs)

        if (stats.cacheHit) {
            reporters.cacheHits.inc()
        } else {
            reporters.cacheMisses.inc()
        }
    }

    if (isInsertStats(stats)) {
        reporters.insertDumpDuration.observe(stats.elapsedMs)
    }

    if (isCreateRunnerStats(stats)) {
        reporters.createRunnerDuration.observe(stats.elapsedMs)
    }

    if (isQueryStats(stats)) {
        reporters.queryDuration.observe(stats.elapsedMs)
    }
}

//
// Type Guards

function isCacheStats(stats: any): stats is CacheStats {
    return true
}

function isInsertStats(stats: any): stats is InsertStats {
    return true
}

function isCreateRunnerStats(stats: any): stats is CreateRunnerStats {
    return true
}

function isQueryStats(stats: any): stats is QueryStats {
    return true
}
