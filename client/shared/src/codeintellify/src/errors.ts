export interface ErrorLike {
    message: string
    name?: string
}

export const isErrorLike = (value: any): value is ErrorLike =>
    typeof value === 'object' && value !== null && ('message' in value || 'stack' in value) && !('__typename' in value)

/**
 * Ensures a value is a proper Error, copying all properties if needed
 */
export const asError = (error: any): Error => {
    if (error instanceof Error) {
        return error
    }
    if (typeof error === 'object' && error !== null) {
        return Object.assign(new Error(error.message), error)
    }
    return new Error(error)
}
