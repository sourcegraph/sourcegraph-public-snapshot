import { AbortError } from '@sourcegraph/common'
import { HTTPStatusError } from '@sourcegraph/http-client'

import { shouldErrorBeReported } from './shouldErrorBeReported'

const createHTTPStatusError = ({ status = 500 }) => {
    const errorResponse = new Response(null, { status })
    return new HTTPStatusError(errorResponse)
}

describe('shouldErrorBeReported', () => {
    test('should capture HttpStatusError except for Server response errors (5xx)', () => {
        expect(shouldErrorBeReported(createHTTPStatusError({ status: 500 }))).toBe(false)
        expect(shouldErrorBeReported(createHTTPStatusError({ status: 400 }))).toBe(true)
    })

    test('should not capture AbortError', () => {
        expect(shouldErrorBeReported(new AbortError())).toBe(false)
    })

    test('should not capture ChunkLoadError', () => {
        const ChunkError = new Error('Loading chunk 123 failed.')
        ChunkError.name = 'ChunkLoadError'

        expect(shouldErrorBeReported(ChunkError)).toBe(false)
    })

    test('should not capture not authenticated error', () => {
        expect(shouldErrorBeReported(new Error('not authenticated'))).toBe(false)
    })
})
