import { isErrorLike } from '@sourcegraph/common'
import { HTTPStatusError } from '@sourcegraph/http-client'

export function shouldErrorBeReported(error: unknown): boolean {
    if (error instanceof HTTPStatusError) {
        // Ignore Server error responses (5xx).
        return error.status < 500
    }

    if (isChunkLoadError(error) || isAbortError(error) || isNotAuthenticatedError(error) || isNetworkError(error)) {
        return false
    }

    return true
}

export function isChunkLoadError(value: unknown): boolean {
    return isDynamicImportError(value)
}

function isDynamicImportError(value: unknown): boolean {
    return (
        isErrorLike(value) &&
        value.name === 'TypeError' &&
        value.message.startsWith('Failed to fetch dynamically imported module')
    )
}

function isAbortError(value: unknown): boolean {
    return isErrorLike(value) && (value.name === 'AbortError' || value.message.startsWith('AbortError'))
}

function isNotAuthenticatedError(value: unknown): boolean {
    return isErrorLike(value) && value.message.includes('not authenticated')
}

function isNetworkError(value: unknown): boolean {
    return isErrorLike(value) && /(networkerror|failed to fetch|load failed)/gi.test(value.message)
}
