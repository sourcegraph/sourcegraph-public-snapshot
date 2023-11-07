import { describe, expect, it } from '@jest/globals'

import { shortenPath } from './utils'

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
