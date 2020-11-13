import { getHoverResult } from './hover'
import { scanSearchQuery, ScanSuccess, Token, ScanResult } from './scanner'

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

describe('getHoverResult()', () => {
    test('returns hover contents for filters', () => {
        const scannedQuery = toSuccess(scanSearchQuery('repo:sourcegraph file:code_intelligence'))
        expect(getHoverResult(scannedQuery, { column: 4 })).toStrictEqual({
            contents: [
                {
                    value: 'Include only results from repositories matching the given search pattern.',
                },
            ],
            range: {
                endColumn: 17,
                endLineNumber: 1,
                startColumn: 1,
                startLineNumber: 1,
            },
        })
        expect(getHoverResult(scannedQuery, { column: 18 })).toStrictEqual({
            contents: [
                {
                    value: 'Include only results from files matching the given search pattern.',
                },
            ],
            range: {
                endColumn: 40,
                endLineNumber: 1,
                startColumn: 18,
                startLineNumber: 1,
            },
        })
        expect(getHoverResult(scannedQuery, { column: 30 })).toStrictEqual({
            contents: [
                {
                    value: 'Include only results from files matching the given search pattern.',
                },
            ],
            range: {
                endColumn: 40,
                endLineNumber: 1,
                startColumn: 18,
                startLineNumber: 1,
            },
        })
    })
})
