import { HTTPStatusError } from '@sourcegraph/shared/src/backend/fetch'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

export function shouldErrorBeReported(error: unknown): boolean {
    if (error instanceof HTTPStatusError) {
        // Ignore Server error responses (5xx).
        return error.status < 500
    }

    if (isWebpackChunkError(error) || isAbortError(error) || isNotAuthenticatedError(error)) {
        return false
    }

    return true
}

export function isWebpackChunkError(value: unknown): boolean {
    return isErrorLike(value) && value.name === 'ChunkLoadError'
}

function isAbortError(value: unknown): boolean {
    return isErrorLike(value) && (value.name === 'AbortError' || value.message.startsWith('AbortError'))
}

function isNotAuthenticatedError(value: unknown): boolean {
    return isErrorLike(value) && value.message.includes('not authenticated')
}
