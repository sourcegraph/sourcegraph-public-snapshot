import { describe, expect, test } from '@jest/globals'

import type { LineOrPositionOrRange } from '@sourcegraph/common'

import { parseHash } from './codeHost'

describe('parseHash', () => {
    const entries: [string, LineOrPositionOrRange][] = [
        ['#lines-1', { line: 1 }],
        ['#lines-1:5', { line: 1, endLine: 5 }],
    ]

    for (const [hash, expectedValue] of entries) {
        test(`given "${hash}" as argument returns ${JSON.stringify(expectedValue)}`, () => {
            expect(parseHash(hash)).toEqual(expectedValue)
        })
    }
})
