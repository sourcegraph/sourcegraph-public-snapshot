import { describe, expect, test } from '@jest/globals'

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

    test('should not capture dynamic import Error', () => {
        expect(shouldErrorBeReported(new TypeError('Failed to fetch dynamically imported module'))).toBe(false)
    })

    test('should not capture not authenticated error', () => {
        expect(shouldErrorBeReported(new Error('not authenticated'))).toBe(false)
    })
})
