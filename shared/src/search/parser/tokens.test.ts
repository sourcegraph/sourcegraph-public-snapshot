import { getMonacoTokens } from './tokens'
import { parseSearchQuery, ParseSuccess, Sequence } from './parser'

describe('getMonacoTokens()', () => {
    test('returns the tokens for a parsed search query', () => {
        expect(
            getMonacoTokens(
                (parseSearchQuery('r:^github.com/sourcegraph f:code_intelligence trackViews') as ParseSuccess<Sequence>)
                    .token
            )
        ).toStrictEqual([
            {
                scopes: 'keyword',
                startIndex: 0,
            },
            {
                scopes: 'identifier',
                startIndex: 2,
            },
            {
                scopes: 'whitespace',
                startIndex: 25,
            },
            {
                scopes: 'keyword',
                startIndex: 26,
            },
            {
                scopes: 'identifier',
                startIndex: 28,
            },
            {
                scopes: 'whitespace',
                startIndex: 45,
            },
            {
                scopes: 'identifier',
                startIndex: 46,
            },
        ])
    })
})
