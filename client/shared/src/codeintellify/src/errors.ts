export interface ErrorLike {
    message: string
    name?: string
}

export const isErrorLike = (val: any): val is ErrorLike =>
    typeof val === 'object' && val !== null && ('message' in val || 'stack' in val) && !('__typename' in val)

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
