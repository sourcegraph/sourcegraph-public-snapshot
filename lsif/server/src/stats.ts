/**
 * TODO
 */
export interface EncodingStats {
    // The time it took to perform the encoding.
    elapsedMs: number
    // The amount of space this dump occupies on disk.
    diskKb: number
}

/**
 * TODO
 */
export interface HandleStats {
    // The time it took to create a handle to the target database.
    elapsedMs: number
}

/**
 * TODO
 */
export interface CacheStats {
    // The time it took to get a reference to the target database.
    elapsedMs: number
    // Whether or not the database handle was already in memory.
    cacheHit: boolean
}

/**
 * TODO
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
