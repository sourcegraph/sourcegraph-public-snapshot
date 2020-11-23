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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('r:^github.com/sourcegraph f:code_intelligence trackViews'))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 2,
                "scopes": "identifier"
              },
              {
                "startIndex": 25,
                "scopes": "whitespace"
              },
              {
                "startIndex": 26,
                "scopes": "field"
              },
              {
                "startIndex": 28,
                "scopes": "identifier"
              },
              {
                "startIndex": 45,
                "scopes": "whitespace"
              },
              {
                "startIndex": 46,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('search query containing parenthesized parameters', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('r:a (f:b and c)')))).toMatchInlineSnapshot(
            `
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 2,
                "scopes": "identifier"
              },
              {
                "startIndex": 3,
                "scopes": "whitespace"
              },
              {
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "field"
              },
              {
                "startIndex": 7,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "whitespace"
              },
              {
                "startIndex": 9,
                "scopes": "keyword"
              },
              {
                "startIndex": 12,
                "scopes": "whitespace"
              },
              {
                "startIndex": 13,
                "scopes": "identifier"
              },
              {
                "startIndex": 14,
                "scopes": "identifier"
              }
            ]
        `
        )
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
                "scopes": "regexpMetaEscapedCharacter"
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
                "scopes": "regexpMetaRangeQuantifier"
              },
              {
                "startIndex": 2,
                "scopes": "regexpMetaLazyQuantifier"
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
                "scopes": "regexpMetaRangeQuantifier"
              }
            ]
        `)
    })

    test('decorate range quantifier', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('b{1} c{1,2} d{3,}?', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "identifier"
              },
              {
                "startIndex": 1,
                "scopes": "regexpMetaRangeQuantifier"
              },
              {
                "startIndex": 4,
                "scopes": "whitespace"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 6,
                "scopes": "regexpMetaRangeQuantifier"
              },
              {
                "startIndex": 11,
                "scopes": "whitespace"
              },
              {
                "startIndex": 12,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "regexpMetaRangeQuantifier"
              },
              {
                "startIndex": 17,
                "scopes": "regexpMetaLazyQuantifier"
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

    test('decorate non-capturing paren groups', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('(?:a(?:b))', false, SearchPatternType.regexp)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 3,
                "scopes": "identifier"
              },
              {
                "startIndex": 4,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 7,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 9,
                "scopes": "regexpMetaDelimited"
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
            getMonacoTokens(
                toSuccess(scanSearchQuery('repo:^foo$ count:10 file:.* fork:yes', false, SearchPatternType.regexp)),
                true
            )
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
                "scopes": "field"
              },
              {
                "startIndex": 17,
                "scopes": "identifier"
              },
              {
                "startIndex": 19,
                "scopes": "whitespace"
              },
              {
                "startIndex": 20,
                "scopes": "field"
              },
              {
                "startIndex": 25,
                "scopes": "regexpMetaCharacterSet"
              },
              {
                "startIndex": 26,
                "scopes": "regexpMetaRangeQuantifier"
              },
              {
                "startIndex": 27,
                "scopes": "whitespace"
              },
              {
                "startIndex": 28,
                "scopes": "field"
              },
              {
                "startIndex": 33,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate regexp | operator, single pattern', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('[|]\\|((a|b)|d)|e', false, SearchPatternType.regexp)), true))
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
                "scopes": "regexpMetaEscapedCharacter"
              },
              {
                "startIndex": 5,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 6,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 7,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "regexpMetaAlternative"
              },
              {
                "startIndex": 9,
                "scopes": "identifier"
              },
              {
                "startIndex": 10,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 11,
                "scopes": "regexpMetaAlternative"
              },
              {
                "startIndex": 12,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 14,
                "scopes": "regexpMetaAlternative"
              },
              {
                "startIndex": 15,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate regexp | operator, multiple patterns', () => {
        expect(
            getMonacoTokens(toSuccess(scanSearchQuery('repo:(a|b) (c|d) (e|f)', false, SearchPatternType.regexp)), true)
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 5,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 7,
                "scopes": "regexpMetaAlternative"
              },
              {
                "startIndex": 8,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 10,
                "scopes": "whitespace"
              },
              {
                "startIndex": 11,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 12,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "regexpMetaAlternative"
              },
              {
                "startIndex": 14,
                "scopes": "identifier"
              },
              {
                "startIndex": 15,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 16,
                "scopes": "whitespace"
              },
              {
                "startIndex": 17,
                "scopes": "regexpMetaDelimited"
              },
              {
                "startIndex": 18,
                "scopes": "identifier"
              },
              {
                "startIndex": 19,
                "scopes": "regexpMetaAlternative"
              },
              {
                "startIndex": 20,
                "scopes": "identifier"
              },
              {
                "startIndex": 21,
                "scopes": "regexpMetaDelimited"
              }
            ]
        `)
    })

    test('decorate escaped characters', () => {
        expect(
            getMonacoTokens(
                toSuccess(scanSearchQuery('[\\--\\abc] \\|\\.|\\(\\)', false, SearchPatternType.regexp)),
                true
            )
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 1,
                "scopes": "regexpMetaEscapedCharacter"
              },
              {
                "startIndex": 3,
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 4,
                "scopes": "regexpMetaEscapedCharacter"
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
                "scopes": "regexpMetaCharacterClass"
              },
              {
                "startIndex": 9,
                "scopes": "whitespace"
              },
              {
                "startIndex": 10,
                "scopes": "regexpMetaEscapedCharacter"
              },
              {
                "startIndex": 12,
                "scopes": "regexpMetaEscapedCharacter"
              },
              {
                "startIndex": 14,
                "scopes": "regexpMetaAlternative"
              },
              {
                "startIndex": 15,
                "scopes": "regexpMetaEscapedCharacter"
              },
              {
                "startIndex": 17,
                "scopes": "regexpMetaEscapedCharacter"
              }
            ]
        `)
    })

    test('decorate structural holes', () => {
        expect(
            getMonacoTokens(
                toSuccess(scanSearchQuery('r:foo Search(thing::[x], :[y])', false, SearchPatternType.structural)),
                true
            )
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "whitespace"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 19,
                "scopes": "structuralMetaHole"
              },
              {
                "startIndex": 23,
                "scopes": "identifier"
              },
              {
                "startIndex": 25,
                "scopes": "structuralMetaHole"
              },
              {
                "startIndex": 29,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate structural holes with valid inlined regexp', () => {
        expect(
            getMonacoTokens(toSuccess(scanSearchQuery('r:foo a:[x~[\\]]]b', false, SearchPatternType.structural)), true)
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "whitespace"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 7,
                "scopes": "structuralMetaHole"
              },
              {
                "startIndex": 16,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate structural hole ... alias', () => {
        expect(
            getMonacoTokens(
                toSuccess(scanSearchQuery('r:foo a...b...c....', false, SearchPatternType.structural)),
                true
            )
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "whitespace"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 7,
                "scopes": "structuralMetaHole"
              },
              {
                "startIndex": 10,
                "scopes": "identifier"
              },
              {
                "startIndex": 11,
                "scopes": "structuralMetaHole"
              },
              {
                "startIndex": 14,
                "scopes": "identifier"
              },
              {
                "startIndex": 15,
                "scopes": "structuralMetaHole"
              },
              {
                "startIndex": 18,
                "scopes": "identifier"
              }
            ]
        `)
    })
    test('decorate structural hole ... alias', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('r:foo ...:...', false, SearchPatternType.structural)), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "whitespace"
              },
              {
                "startIndex": 6,
                "scopes": "structuralMetaHole"
              },
              {
                "startIndex": 9,
                "scopes": "identifier"
              },
              {
                "startIndex": 10,
                "scopes": "structuralMetaHole"
              }
            ]
        `)
    })
})
