import type { ErrorLike } from './types'

/**
 * Run the passed function and return `undefined` if it throws an error.
 */
export function tryCatch<T>(function_: () => T): T | undefined {
    try {
        return function_()
    } catch {
        return undefined
    }
}

export const isErrorLike = (value: unknown): value is ErrorLike =>
    typeof value === 'object' && value !== null && ('stack' in value || 'message' in value) && !('__typename' in value)
