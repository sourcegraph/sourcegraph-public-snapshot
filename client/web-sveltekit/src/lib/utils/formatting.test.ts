import { describe, test, expect } from 'vitest'

import { formatBytes } from './formatting'

describe('formatBytes', () => {
    test.each`
        bytes        | expected
        ${0}         | ${'0 B'}
        ${512}       | ${'512 B'}
        ${1024}      | ${'1.00 KB'}
        ${1024 ** 2} | ${'1.00 MB'}
        ${1024 ** 3} | ${'1.00 GB'}
    `('$bytes bytes => $expected', ({ bytes, expected }) => {
        expect(formatBytes(bytes)).toBe(expected)
    })
})
