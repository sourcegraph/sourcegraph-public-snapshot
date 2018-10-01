import pRetry from 'p-retry'
import { OperationOptions } from 'retry'

/**
 * Retry function with more sensible defaults for e2e test assertions
 *
 * @param fn The async assertion function to retry
 * @param options Option overrides passed to pRetry
 */
export const retry = (fn: (attempt: number) => Promise<void>, options: OperationOptions = {}) =>
    pRetry(fn, { factor: 1, ...options })
