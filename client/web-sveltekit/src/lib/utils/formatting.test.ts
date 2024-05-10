import { describe, test, expect } from 'vitest'

import { formatBytes } from './formatting'

describe('formatBytes', () => {
    test.each`
        bytes        | expected
        ${0}         | ${'0 B'}
        ${512}       | ${'512 B'}
        ${1000}      | ${'1.00 KB'}
        ${1000 ** 2} | ${'1.00 MB'}
        ${1000 ** 3} | ${'1.00 GB'}
    `('$bytes bytes => $expected', ({ bytes, expected }) => {
        expect(formatBytes(bytes)).toBe(expected)
    })
})
