import { describe, test, expect } from 'vitest'

import { CodeViewMode, toCodeViewMode } from './util'

describe('toViewMode', () => {
    test.each`
        input        | expected
        ${undefined} | ${CodeViewMode.Default}
        ${null}      | ${CodeViewMode.Default}
        ${'code'}    | ${CodeViewMode.Code}
        ${'CoDe'}    | ${CodeViewMode.Code}
        ${'raw'}     | ${CodeViewMode.Code}
        ${'blame'}   | ${CodeViewMode.Blame}
        ${'BlAmE'}   | ${CodeViewMode.Blame}
    `('$input -> $expected', ({ input, expected }) => {
        expect(toCodeViewMode(input)).toBe(expected)
    })
})
