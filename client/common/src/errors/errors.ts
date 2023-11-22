import { AGGREGATE_ERROR_NAME } from './constants'
import type { ErrorLike } from './types'
import { isErrorLike } from './utils'

/**
 * Converts an ErrorLike to a proper Error if needed, copying all properties
 *
 * @param value An Error, object with ErrorLike properties, or other value.
 */
export const asError = (value: unknown): Error => {
    if (value instanceof Error) {
        return value
    }
    if (isErrorLike(value)) {
        return Object.assign(new Error(value.message), value)
    }
    return new Error(String(value))
}

/**
 * DEPRECATED: use dataOrThrowErrors instead
 * Creates an aggregate error out of multiple provided error likes
 *
 * @param errors The errors or ErrorLikes to aggregate
 */
export const createAggregateError = (errors: readonly ErrorLike[] = []): Error =>
    errors.length === 1
        ? asError(errors[0])
        : Object.assign(new Error(errors.map(error => error.message).join('\n')), {
              name: AGGREGATE_ERROR_NAME,
              errors: errors.map(asError),
          })

export class AbortError extends Error {
    public readonly name = 'AbortError'
    public readonly message = 'Aborted'
}
