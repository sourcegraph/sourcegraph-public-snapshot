import { getSuggestionQuery } from './providers'
import { ScanResult, scanSearchQuery, ScanSuccess } from './scanner'
import { Token } from './token'

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

function getTokens(query: string, tokenIndex: number): [Token[], Token] {
    const tokens = toSuccess(scanSearchQuery(query))
    return [tokens, tokens[tokenIndex]]
}

describe('getSuggestionQuery', () => {
    test('should generate suggestion query for repo filter', () => {
        expect(getSuggestionQuery(...getTokens('repo:a', 0))).toEqual('repo:a type:repo count:50')
    })

    test('should not add non-relevant filters and patterns to repo suggestion query', () => {
        expect(getSuggestionQuery(...getTokens('repo:a count:10 pattern', 0))).toEqual('repo:a type:repo count:50')
    })

    test('should archived, visibility, and fork filters to repo suggestion query', () => {
        expect(getSuggestionQuery(...getTokens('visibility:private archived:no fork:only repo:a', 6))).toEqual(
            'visibility:private archived:no fork:only repo:a type:repo count:50'
        )
    })

    test('should not include relevant filters for repo suggestion query if inside an `or` expression', () => {
        expect(getSuggestionQuery(...getTokens('(visibility:private repo:a) or repo:b', 3))).toEqual(
            'repo:a type:repo count:50'
        )
    })

    test('should not generate a file suggestion query if inside an `or` expression', () => {
        expect(getSuggestionQuery(...getTokens('(file:a) or pattern', 1))).toEqual('')
    })

    test('should generate a file suggestion query with relevant filters', () => {
        expect(getSuggestionQuery(...getTokens('file:a archived:yes lang:Go', 0))).toEqual(
            'archived:yes lang:Go file:a type:path count:50'
        )
    })

    test('should generate a symbol suggestion query with relevant filters', () => {
        expect(getSuggestionQuery(...getTokens('pattern fork:only repo:a lang:Go', 0))).toEqual(
            'fork:only repo:a lang:Go pattern type:symbol count:50'
        )
    })
})
