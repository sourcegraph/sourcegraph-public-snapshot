export interface ErrorLike {
    message: string
    code?: string
}

export const isErrorLike = (val: unknown): val is ErrorLike =>
    typeof val === 'object' &&
    !!val &&
    ('stack' in val || ('message' in val || 'code' in val)) &&
    !('__typename' in val)

/**
 * Converts an ErrorLike to a proper Error if needed, copying all properties
 *
 * @param errorLike An Error or object with ErrorLike properties
 */
export const asError = (err: any): Error => {
    if (err instanceof Error) {
        return err
    }
    if (typeof err === 'object' && err !== null) {
        return Object.assign(new Error(err.message), err)
    }
    return new Error(err)
}

/**
 * An Error that aggregates multiple errors
 */
interface AggregateError extends Error {
    name: 'AggregateError'
    errors: Error[]
}

/**
 * DEPRECATED: use dataOrThrowErrors instead
 * Creates an aggregate error out of multiple provided error likes
 *
 * @param errors The errors or ErrorLikes to aggregate
 */
export const createAggregateError = (errors: ErrorLike[] = []): AggregateError =>
    Object.assign(new Error(errors.map(e => e.message).join('\n')), {
        name: 'AggregateError' as const,
        errors: errors.map(asError),
    })

/**
 * Improves error messages in case of ajax errors
 */
export const normalizeAjaxError = (err: any): void => {
    if (!err) {
        return
    }
    if (typeof err.status === 'number') {
        if (err.status === 0) {
            err.message = 'Unable to reach server. Check your network connection and try again in a moment.'
        } else {
            err.message = `Unexpected HTTP error: ${err.status}`
            if (err.xhr && err.xhr.statusText) {
                err.message += ` ${err.xhr.statusText}`
            }
        }
    }
}
