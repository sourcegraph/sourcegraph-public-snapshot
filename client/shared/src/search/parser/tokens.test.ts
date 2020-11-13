import { getMonacoTokens } from './tokens'
import { scanSearchQuery, ScanSuccess, Token, ScanResult } from './scanner'
import { SearchPatternType } from '../../graphql-operations'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

describe('getMonacoTokens()', () => {
    test('returns the tokens for a parsed search query', () => {
        expect(
            getMonacoTokens(toSuccess(scanSearchQuery('r:^github.com/sourcegraph f:code_intelligence trackViews')))
        ).toStrictEqual([
            {
                scopes: 'filterKeyword',
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
                scopes: 'filterKeyword',
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

    test('search query containing parenthesized parameters', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('r:a (f:b and c)')))).toStrictEqual([
            {
                scopes: 'filterKeyword',
                startIndex: 0,
            },
            {
                scopes: 'identifier',
                startIndex: 2,
            },
            {
                scopes: 'whitespace',
                startIndex: 3,
            },
            {
                scopes: 'identifier',
                startIndex: 4,
            },
            {
                scopes: 'filterKeyword',
                startIndex: 5,
            },
            {
                scopes: 'identifier',
                startIndex: 7,
            },
            {
                scopes: 'whitespace',
                startIndex: 8,
            },
            {
                scopes: 'keyword',
                startIndex: 9,
            },
            {
                scopes: 'whitespace',
                startIndex: 12,
            },
            {
                scopes: 'identifier',
                startIndex: 13,
            },
            {
                scopes: 'identifier',
                startIndex: 14,
            },
        ])
    })

    test('no decoration for literal', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('(a\\sb)', false, SearchPatternType.literal)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate regexp character set and group', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('(a\\sb)', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 1,
                "scopes": "identifier"
              },
              {
                "startIndex": 2,
                "scopes": "regexpMetaCharacterSet"
              },
              {
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "regexpMetaDelimited"
              }
            ]
        `)
    })

    test('decorate regexp assertion', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('^oh\\.hai$', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "regexpMetaAssertion"
              },
              {
                "startIndex": 1,
                "scopes": "identifier"
              },
              {
                "startIndex": 2,
                "scopes": "identifier"
              },
              {
                "startIndex": 3,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 7,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "regexpMetaAssertion"
              }
            ]
        `)
    })

    test('decorate regexp quantifiers', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('a*?(b)+', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "identifier"
              },
              {
                "startIndex": 1,
                "scopes": "regexpMetaQuantifier"
              },
              {
                "startIndex": 3,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 6,
                "scopes": "regexpMetaQuantifier"
              }
            ]
        `)
    })

    test('decorate paren groups', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('((a) or b)', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "openingParen"
              },
              {
                "startIndex": 1,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 2,
                "scopes": "identifier"
              },
              {
                "startIndex": 3,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 4,
                "scopes": "whitespace"
              },
              {
                "startIndex": 5,
                "scopes": "keyword"
              },
              {
                "startIndex": 7,
                "scopes": "whitespace"
              },
              {
                "startIndex": 8,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "closingParen"
              }
            ]
        `)
    })

    test('decorate character classes', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('([a-z])', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 1,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 2,
                "scopes": "identifier"
              },
              {
                "startIndex": 3,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 6,
                "scopes": "regexpMetaDelimited"
              }
            ]
        `)
    })

    test('decorate character classes', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('[a-z][--z][--z]', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 1,
                "scopes": "identifier"
              },
              {
                "startIndex": 2,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 3,
                "scopes": "identifier"
              },
              {
                "startIndex": 4,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 5,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 7,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 8,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 10,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 11,
                "scopes": "identifier"
              },
              {
                "startIndex": 12,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 13,
                "scopes": "identifier"
              },
              {
                "startIndex": 14,
                "scopes": "regexpMetaCharacterClass"
              }
            ]
        `)
    })

    test('decorate regexp field values', () => {
        expect(
            getMonacoTokens(toSuccess(scanSearchQuery('repo:^foo$ count:.*', false, SearchPatternType.regexp)), true)
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "filterKeyword"
              },
              {
                "startIndex": 5,
                "scopes": "regexpMetaAssertion"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 7,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "regexpMetaAssertion"
              },
              {
                "startIndex": 10,
                "scopes": "whitespace"
              },
              {
                "startIndex": 11,
                "scopes": "filterKeyword"
              },
              {
                "startIndex": 17,
                "scopes": "identifier"
              }
            ]
        `)
    })
})
