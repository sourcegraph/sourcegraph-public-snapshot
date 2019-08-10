/**
 * Process statistics around a critical section.
 */
export interface ProcessStats {
    // The wall time it took to perfrom an action.
    elapsedMs: number
    // The cpu time it took to perform an action.
    elapsedCpuMs: number
    // The difference in the size of the resident set after performing the action.
    rssDiff: number
    // The difference in the size of the heap after performing the action.
    heapTotalDiff: number
    // The difference in the size of the used portion of the heap after performing the action.
    heapUsedDiff: number
    // The difference in the size of external allocations after performing the action.
    externalDiff: number
}

/**
 * Runtime statistics around the `withDB` cache method.
 */
export interface CacheStats {
    // Process stats around getting a reference to the target database.
    processStats: ProcessStats
    // Whether or not the database handle was already in memory.
    cacheHit: boolean
}

/**
 * Runtime statistics around the `insertDump` backend method.
 */
export interface InsertStats {
    // Process stats around performing LSIF dump encoding.
    processStats: ProcessStats
    // The amount of space this dump occupies on disk.
    diskKb: number
}

/**
 * Runtime statistics around the `createRunner` backend method.
 */
export interface CreateRunnerStats {
    // Process stats around creating a handle to the target database.
    processStats: ProcessStats
}

/**
 * Runtime statistics around `query` query runner method.
 */
export interface QueryStats {
    // Process stats around performing the query.
    processStats: ProcessStats
}

/**
 * Run an operation and return the result, along with statistics about the wall
 * time, cpu time, and memory that it required to execute.
 */
export async function instrument<T>(fn: () => Promise<T>): Promise<{ result: T; processStats: ProcessStats }> {
    const startTime = process.hrtime()
    const startCpuUsage = process.cpuUsage()
    const startMemUsage = process.memoryUsage()

    const result = await fn()

    const [seconds, nanoseconds] = process.hrtime(startTime)
    const cpuUsage = process.cpuUsage(startCpuUsage)
    const endMemUsage = process.memoryUsage()

    const processStats = {
        elapsedMs: seconds * 1e3 + nanoseconds / 1e6,
        elapsedCpuMs: cpuUsage.system / 1e3 + cpuUsage.user / 1e3,
        rssDiff: endMemUsage.rss - startMemUsage.rss,
        heapTotalDiff: endMemUsage.heapTotal - startMemUsage.heapTotal,
        heapUsedDiff: endMemUsage.heapUsed - startMemUsage.heapUsed,
        externalDiff: endMemUsage.external - startMemUsage.external,
    }

    return { result, processStats }
}
