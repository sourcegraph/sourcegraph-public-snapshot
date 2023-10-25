import { describe, expect, test } from 'vitest'

import { stringHuman } from './printer'
import { type ScanResult, scanSearchQuery, type ScanSuccess } from './scanner'
import type { Token } from './token'

expect.addSnapshotSerializer({
    serialize: value => value as string,
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

describe('stringHuman', () => {
    test('complex query', () => {
        const tokens = toSuccess(
            scanSearchQuery('(a or b) (-repo:foo    AND file:bar) content:"count:5000" /yowza/ "a\'b" \\d+')
        )
        expect(stringHuman(tokens)).toMatchInlineSnapshot(
            '(a or b) (-repo:foo AND file:bar) content:"count:5000" /yowza/ "a\'b" \\d+'
        )
    })

    test('render delimited syntax', () => {
        const tokens = toSuccess(scanSearchQuery('patterntype:standard /test\\ .*me/ "and me"'))
        expect(stringHuman(tokens)).toMatchInlineSnapshot('patterntype:standard /test\\ .*me/ "and me"')
    })
})
