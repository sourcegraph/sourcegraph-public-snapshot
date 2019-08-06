/**
 * Runtime statistics around the `withDB` cache method.
 */
export interface CacheStats {
    // The time it took to get a reference to the target database.
    elapsedMs: number
    // Whether or not the database handle was already in memory.
    cacheHit: boolean
}

/**
 * Runtime statistics around the `insertDump` backend method.
 */
export interface InsertStats {
    // The time it took to perform the encoding.
    elapsedMs: number
    // The amount of space this dump occupies on disk.
    diskKb: number
}

/**
 * Runtime statistics around the `createRunner` backend method.
 */
export interface CreateRunnerStats {
    // The time it took to create a handle to the target database.
    elapsedMs: number
}

/**
 * Runtime statistics around `query` query runner method.
 */
export interface QueryStats {
    // The time it took to perform the query.
    elapsedMs: number
}

/**
 * Time an operation and return the result as well as the time it took to execute
 * in milliseconds (using a high-resolution timer).
 */
export async function timeit<T>(fn: () => Promise<T>): Promise<{ result: T; elapsed: number }> {
    const start = process.hrtime()
    const result = await fn()
    const elapsed = process.hrtime(start)[1] / 1000000
    return { result, elapsed }
}
