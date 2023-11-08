import { describe, expect, it } from 'vitest'

import { shortenPath, getSpans } from './utils'

describe('shortenPath', () => {
    it('returns the original path when it is shorter than the desired length', () => {
        expect(shortenPath('a/b/c/d/e/f/g/h', 20)).toBe('a/b/c/d/e/f/g/h')
    })

    it('returns the original path when it does not have sufficient segments', () => {
        expect(shortenPath('thispathonlyhas/twosegements', 5)).toBe('thispathonlyhas/twosegements')
    })

    it('preserves the first and last two segements', () => {
        expect(shortenPath('one/two/three/four/five/six/seven', 5)).toBe('one/two/.../six/seven')
    })
})

describe('getSpans', () => {
    it('returns appropriate spans for the provided input', () => {
        expect(getSpans(new Set([0, 1, 2, 3]), 10)).toEqual([
            [0, 3, true],
            [4, 9, false],
        ])
        expect(getSpans(new Set([0, 1, 2, 5, 9]), 10)).toEqual([
            [0, 2, true],
            [3, 4, false],
            [5, 5, true],
            [6, 8, false],
            [9, 9, true],
        ])
    })

    it('returns appropriate spans for the provided input shifted by offset', () => {
        expect(getSpans(new Set([4, 5, 6]), 10, 4)).toEqual([
            [0, 2, true],
            [3, 9, false],
        ])
    })
})
