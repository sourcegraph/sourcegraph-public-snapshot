import { describe, expect, it } from '@jest/globals'

import { getBatchCount } from './utils'

describe('getBatchCount', () => {
    const tests: {
        name: string
        screenHeight: number
        expected: number
    }[] = [
        {
            name: 'works for tiny windows',
            screenHeight: 499,
            expected: 5,
        },
        {
            name: 'works for small windows',
            screenHeight: 999,
            expected: 10,
        },
        {
            name: 'works for medium windows',
            screenHeight: 1499,
            expected: 15,
        },
        {
            name: 'works for large windows',
            screenHeight: 1999,
            expected: 25,
        },
        {
            name: 'works for extra large windows',
            screenHeight: 2000,
            expected: 30,
        },
    ]
    for (const t of tests) {
        it(t.name, () => {
            expect(getBatchCount(t.screenHeight)).toBe(t.expected)
        })
    }
})
