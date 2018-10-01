export interface ErrorLike {
    message: string
    code?: string
}

export const isErrorLike = (val: any): val is ErrorLike =>
    !!val && typeof val === 'object' && (!!val.stack || ('message' in val || 'code' in val)) && !('__typename' in val)

/**
 * Ensures a value is a proper Error, copying all properties if needed
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
        name: 'AggregateError' as 'AggregateError',
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
