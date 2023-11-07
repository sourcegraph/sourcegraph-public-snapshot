import { describe, expect, test } from '@jest/globals'

import { type FuzzyFileQuery, parseFuzzyFileQuery } from './FuzzyFiles'

function checkFuzzyFileQuery(query: string, expectedValue: FuzzyFileQuery): void {
    test(query, () => {
        expect(parseFuzzyFileQuery(query)).toStrictEqual(expectedValue)
    })
}

describe('parseFuzzyFileQuery', () => {
    checkFuzzyFileQuery('foo.ts', { filename: 'foo.ts' })
    checkFuzzyFileQuery('foo.ts:', { filename: 'foo.ts', line: 0 })
    checkFuzzyFileQuery('foo.ts:1:', { filename: 'foo.ts', line: 1, column: 0 })
    checkFuzzyFileQuery('foo.ts:100', { filename: 'foo.ts', line: 100 })
    checkFuzzyFileQuery('foo.ts:100:99', { filename: 'foo.ts', line: 100, column: 99 })
})
