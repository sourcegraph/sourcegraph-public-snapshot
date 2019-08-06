/**
 * Runtime statistics around the `createDB` backend method.
 */
export interface EncodingStats {
    // The time it took to perform the encoding.
    elapsedMs: number
    // The amount of space this dump occupies on disk.
    diskKb: number
}

/**
 * Runtime statistics around the `loadDB` backend method.
 */
export interface HandleStats {
    // The time it took to create a handle to the target database.
    elapsedMs: number
}

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
 * Runtime statistics around backend query methods.
 */
export interface QueryStats {
    // TODO
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
