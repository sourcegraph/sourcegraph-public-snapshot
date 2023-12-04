import type { AGGREGATE_ERROR_NAME } from './constants'

export interface ErrorLike {
    message: string
    name?: string
}

/**
 * An Error that aggregates multiple errors
 */
export interface AggregateError extends Error {
    name: typeof AGGREGATE_ERROR_NAME
    errors: Error[]
}
