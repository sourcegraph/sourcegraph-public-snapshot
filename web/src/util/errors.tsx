export interface ErrorLike {
    message?: string
    code?: string
}

export const isErrorLike = (val: any): val is ErrorLike =>
    !!val && typeof val === 'object' && (!!val.stack || ('message' in val || 'code' in val))

/**
 * Converts an ErrorLike to a proper Error if needed, copying all properties
 * @param errorLike An Error or object with ErrorLike properties
 */
export const asError = (errorLike: ErrorLike): Error =>
    errorLike instanceof Error ? errorLike : Object.assign(new Error(errorLike.message), errorLike)

/**
 * An Error that aggregates multiple errors
 */
export interface AggregateError extends Error {
    name: 'AggregateError'
    errors: Error[]
}

/**
 * Creates an aggregate error out of multiple provided error likes
 *
 * @param errors The errors or ErrorLikes to aggregate
 */
export const createAggregateError = (errors: ErrorLike[] = []): AggregateError =>
    Object.assign(new Error(errors.map(e => e.message).join('\n')), {
        name: 'AggregateError' as 'AggregateError',
        errors: errors.map(asError),
    })
