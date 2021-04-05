import { stringHuman } from './printer'
import { ScanResult, scanSearchQuery, ScanSuccess } from './scanner'
import { Token } from './token'

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
})
