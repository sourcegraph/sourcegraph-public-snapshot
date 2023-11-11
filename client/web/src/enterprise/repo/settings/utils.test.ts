import { describe, expect, it } from 'vitest'

import { formatDuration } from './utils'

describe('Repo Settings Utils', () => {
    describe('formatDuration', () => {
        it('formats durations over 1 second to seconds', () => {
            const duration = 1.5
            const expected = '1.50s'
            expect(formatDuration(duration)).toEqual(expected)
        })

        it('formats durations under 1 second to milliseconds', () => {
            const duration = 0.25
            const expected = '250.00ms'
            expect(formatDuration(duration)).toEqual(expected)
        })
    })
})
