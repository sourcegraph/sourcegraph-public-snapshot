import { describe, test, expect } from 'vitest'

import { ViewMode, toViewMode } from './util'

describe('toViewMode', () => {
    test.each`
        input        | expected
        ${undefined} | ${ViewMode.Default}
        ${null}      | ${ViewMode.Default}
        ${'code'}    | ${ViewMode.Code}
        ${'CoDe'}    | ${ViewMode.Code}
        ${'raw'}     | ${ViewMode.Code}
        ${'blame'}   | ${ViewMode.Blame}
        ${'BlAmE'}   | ${ViewMode.Blame}
    `('$input -> $expected', ({ input, expected }) => {
        expect(toViewMode(input)).toBe(expected)
    })
})
