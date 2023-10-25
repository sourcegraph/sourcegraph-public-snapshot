import { describe, expect, test } from '@jest/globals'

import { SearchPatternType } from '../../graphql-operations'

import { type DecoratedToken, decorate } from './decoratedToken'
import { scanSearchQuery, type ScanSuccess, type ScanResult } from './scanner'
import type { Token } from './token'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

const getTokens = (tokens: Token[]): { startIndex: number; scopes: string }[] =>
    tokens.flatMap(token =>
        decorate(token).map((token: DecoratedToken): { startIndex: number; scopes: string } => {
            switch (token.type) {
                case 'field':
                case 'whitespace':
                case 'keyword':
                case 'comment':
                case 'openingParen':
                case 'closingParen':
                case 'metaFilterSeparator':
                case 'metaRepoRevisionSeparator':
                case 'metaContextPrefix': {
                    return {
                        startIndex: token.range.start,
                        scopes: token.type,
                    }
                }
                case 'metaPath':
                case 'metaRevision':
                case 'metaRegexp':
                case 'metaStructural':
                case 'metaPredicate': {
                    // The scopes value is derived from the token type and its kind.
                    // E.g., regexpMetaDelimited derives from {@link RegexpMeta} and {@link RegexpMetaKind}.
                    return {
                        startIndex: token.range.start,
                        scopes: `${token.type}${token.kind}`,
                    }
                }

                default: {
                    return {
                        startIndex: token.range.start,
                        scopes: 'identifier',
                    }
                }
            }
        })
    )

describe('scanSearchQuery() and decorate()', () => {
    test('returns the tokens for a parsed search query', () => {
        expect(getTokens(toSuccess(scanSearchQuery('r:^github.com/sourcegraph f:code_intelligence trackViews'))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpAssertion"
              },
              {
                "startIndex": 3,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 10,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "metaPathSeparator"
              },
              {
                "startIndex": 14,
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
                "startIndex": 27,
                "scopes": "metaFilterSeparator"
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
        expect(getTokens(toSuccess(scanSearchQuery('r:a (f:b and c)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
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
                "scopes": "openingParen"
              },
              {
                "startIndex": 5,
                "scopes": "field"
              },
              {
                "startIndex": 6,
                "scopes": "metaFilterSeparator"
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
                "scopes": "closingParen"
              }
            ]
        `)
    })

    test('no decoration for literal', () => {
        expect(getTokens(toSuccess(scanSearchQuery('(a\\sb)', false, SearchPatternType.standard))))
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
        expect(getTokens(toSuccess(scanSearchQuery('(a\\sb)', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 1,
                "scopes": "identifier"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpDelimited"
              }
            ]
        `)
    })

    test('decorate regexp assertion', () => {
        expect(getTokens(toSuccess(scanSearchQuery('^oh\\.hai$', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpAssertion"
              },
              {
                "startIndex": 1,
                "scopes": "identifier"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpEscapedCharacter"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "metaRegexpAssertion"
              }
            ]
        `)
    })

    test('decorate regexp quantifiers', () => {
        expect(getTokens(toSuccess(scanSearchQuery('a*?(b)+', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "identifier"
              },
              {
                "startIndex": 1,
                "scopes": "metaRegexpRangeQuantifier"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpLazyQuantifier"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 4,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 6,
                "scopes": "metaRegexpRangeQuantifier"
              }
            ]
        `)
    })

    test('decorate range quantifier', () => {
        expect(getTokens(toSuccess(scanSearchQuery('b{1} c{1,2} d{3,}?', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "identifier"
              },
              {
                "startIndex": 1,
                "scopes": "metaRegexpRangeQuantifier"
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
                "scopes": "metaRegexpRangeQuantifier"
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
                "scopes": "metaRegexpRangeQuantifier"
              },
              {
                "startIndex": 17,
                "scopes": "metaRegexpLazyQuantifier"
              }
            ]
        `)
    })

    test('decorate paren groups', () => {
        expect(getTokens(toSuccess(scanSearchQuery('((a) or b)', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "openingParen"
              },
              {
                "startIndex": 1,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 2,
                "scopes": "identifier"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpDelimited"
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
        expect(getTokens(toSuccess(scanSearchQuery('(?:a(?:b))', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 3,
                "scopes": "identifier"
              },
              {
                "startIndex": 4,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 7,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 9,
                "scopes": "metaRegexpDelimited"
              }
            ]
        `)
    })

    test('decorate character classes', () => {
        expect(getTokens(toSuccess(scanSearchQuery('([a-z])', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 1,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpCharacterClassRangeHyphen"
              },
              {
                "startIndex": 4,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 6,
                "scopes": "metaRegexpDelimited"
              }
            ]
        `)
    })

    test('decorate character classes', () => {
        expect(getTokens(toSuccess(scanSearchQuery('[a-z][--z][--z]', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 1,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpCharacterClassRangeHyphen"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 4,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 6,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 7,
                "scopes": "metaRegexpCharacterClassRangeHyphen"
              },
              {
                "startIndex": 8,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 9,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 10,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 11,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 12,
                "scopes": "metaRegexpCharacterClassRangeHyphen"
              },
              {
                "startIndex": 13,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 14,
                "scopes": "metaRegexpCharacterClass"
              }
            ]
        `)
    })

    test('decorate regexp field values', () => {
        expect(
            getTokens(
                toSuccess(scanSearchQuery('repo:^foo$ count:10 file:.* fork:yes', false, SearchPatternType.regexp))
            )
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpAssertion"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "metaRegexpAssertion"
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
                "startIndex": 16,
                "scopes": "metaFilterSeparator"
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
                "startIndex": 24,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 25,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 26,
                "scopes": "metaRegexpRangeQuantifier"
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
                "startIndex": 32,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 33,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate regexp | operator, single pattern', () => {
        expect(getTokens(toSuccess(scanSearchQuery('[|]\\|((a|b)|d)|e', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 1,
                "scopes": "metaRegexpCharacterClassMember"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpEscapedCharacter"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 6,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 7,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 9,
                "scopes": "identifier"
              },
              {
                "startIndex": 10,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 11,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 12,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 14,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 15,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate regexp | operator, multiple patterns', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:(a|b) (c|d) (e|f)', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              },
              {
                "startIndex": 7,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 8,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 10,
                "scopes": "whitespace"
              },
              {
                "startIndex": 11,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 12,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 14,
                "scopes": "identifier"
              },
              {
                "startIndex": 15,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 16,
                "scopes": "whitespace"
              },
              {
                "startIndex": 17,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 18,
                "scopes": "identifier"
              },
              {
                "startIndex": 19,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 20,
                "scopes": "identifier"
              },
              {
                "startIndex": 21,
                "scopes": "metaRegexpDelimited"
              }
            ]
        `)
    })

    test('decorate escaped characters', () => {
        expect(getTokens(toSuccess(scanSearchQuery('[--\\\\abc] \\|\\.|\\(\\)', false, SearchPatternType.regexp))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 1,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpCharacterClassRangeHyphen"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpCharacterClassRange"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpEscapedCharacter"
              },
              {
                "startIndex": 5,
                "scopes": "metaRegexpCharacterClassMember"
              },
              {
                "startIndex": 6,
                "scopes": "metaRegexpCharacterClassMember"
              },
              {
                "startIndex": 7,
                "scopes": "metaRegexpCharacterClassMember"
              },
              {
                "startIndex": 8,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 9,
                "scopes": "whitespace"
              },
              {
                "startIndex": 10,
                "scopes": "metaRegexpEscapedCharacter"
              },
              {
                "startIndex": 12,
                "scopes": "metaRegexpEscapedCharacter"
              },
              {
                "startIndex": 14,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 15,
                "scopes": "metaRegexpEscapedCharacter"
              },
              {
                "startIndex": 17,
                "scopes": "metaRegexpEscapedCharacter"
              }
            ]
        `)
    })

    test('decorate structural holes', () => {
        expect(
            getTokens(toSuccess(scanSearchQuery('r:foo Search(thing::[x], :[y])', false, SearchPatternType.structural)))
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 2,
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
                "scopes": "metaStructuralHole"
              },
              {
                "startIndex": 23,
                "scopes": "identifier"
              },
              {
                "startIndex": 25,
                "scopes": "metaStructuralHole"
              },
              {
                "startIndex": 29,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate structural holes with valid inlined regexp', () => {
        expect(getTokens(toSuccess(scanSearchQuery('r:foo a:[x~[\\]]]b', false, SearchPatternType.structural))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 2,
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
                "scopes": "metaStructuralRegexpHole"
              },
              {
                "startIndex": 9,
                "scopes": "metaStructuralVariable"
              },
              {
                "startIndex": 10,
                "scopes": "metaStructuralRegexpSeparator"
              },
              {
                "startIndex": 11,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 12,
                "scopes": "metaRegexpEscapedCharacter"
              },
              {
                "startIndex": 14,
                "scopes": "metaRegexpCharacterClass"
              },
              {
                "startIndex": 15,
                "scopes": "metaStructuralRegexpHole"
              },
              {
                "startIndex": 16,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate structural holes with valid inlined regexp, no variable', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:foo :[~a?|b*]', false, SearchPatternType.structural))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "whitespace"
              },
              {
                "startIndex": 9,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "metaStructuralRegexpHole"
              },
              {
                "startIndex": 11,
                "scopes": "metaStructuralVariable"
              },
              {
                "startIndex": 11,
                "scopes": "metaStructuralRegexpSeparator"
              },
              {
                "startIndex": 12,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "metaRegexpRangeQuantifier"
              },
              {
                "startIndex": 14,
                "scopes": "metaRegexpAlternative"
              },
              {
                "startIndex": 15,
                "scopes": "identifier"
              },
              {
                "startIndex": 16,
                "scopes": "metaRegexpRangeQuantifier"
              },
              {
                "startIndex": 17,
                "scopes": "metaStructuralRegexpHole"
              }
            ]
        `)
    })

    test('decorate structural hole ... alias', () => {
        expect(getTokens(toSuccess(scanSearchQuery('r:foo a...b...c....', false, SearchPatternType.structural))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 2,
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
                "scopes": "metaStructuralHole"
              },
              {
                "startIndex": 10,
                "scopes": "identifier"
              },
              {
                "startIndex": 11,
                "scopes": "metaStructuralHole"
              },
              {
                "startIndex": 14,
                "scopes": "identifier"
              },
              {
                "startIndex": 15,
                "scopes": "metaStructuralHole"
              },
              {
                "startIndex": 18,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate structural hole ... alias', () => {
        expect(getTokens(toSuccess(scanSearchQuery('r:foo ...:...', false, SearchPatternType.structural))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 2,
                "scopes": "identifier"
              },
              {
                "startIndex": 5,
                "scopes": "whitespace"
              },
              {
                "startIndex": 6,
                "scopes": "metaStructuralHole"
              },
              {
                "startIndex": 9,
                "scopes": "identifier"
              },
              {
                "startIndex": 10,
                "scopes": "metaStructuralHole"
              }
            ]
        `)
    })

    test('decorate repo revision syntax, separate revisions', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:foo@HEAD:v1.2:3')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "metaRepoRevisionSeparator"
              },
              {
                "startIndex": 9,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 13,
                "scopes": "metaRevisionSeparator"
              },
              {
                "startIndex": 14,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 18,
                "scopes": "metaRevisionSeparator"
              },
              {
                "startIndex": 19,
                "scopes": "metaRevisionLabel"
              }
            ]
        `)
    })

    test('decorate revision field syntax, separate revisions', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:foo rev:HEAD:v1.2:3')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "whitespace"
              },
              {
                "startIndex": 9,
                "scopes": "field"
              },
              {
                "startIndex": 12,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 13,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 17,
                "scopes": "metaRevisionSeparator"
              },
              {
                "startIndex": 18,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 22,
                "scopes": "metaRevisionSeparator"
              },
              {
                "startIndex": 23,
                "scopes": "metaRevisionLabel"
              }
            ]
        `)
    })

    test('do not decorate regex syntax when filter value is quoted', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:"^do-not-attempt$" file:\'.*\'')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 23,
                "scopes": "whitespace"
              },
              {
                "startIndex": 24,
                "scopes": "field"
              },
              {
                "startIndex": 28,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 29,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('decorate repo revision syntax, path with wildcard and negation', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:foo@*refs/heads/*:*!refs/heads/release*'))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "identifier"
              },
              {
                "startIndex": 8,
                "scopes": "metaRepoRevisionSeparator"
              },
              {
                "startIndex": 9,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 9,
                "scopes": "metaRevisionIncludeGlobMarker"
              },
              {
                "startIndex": 10,
                "scopes": "metaRevisionReferencePath"
              },
              {
                "startIndex": 21,
                "scopes": "metaRevisionWildcard"
              },
              {
                "startIndex": 22,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 22,
                "scopes": "metaRevisionSeparator"
              },
              {
                "startIndex": 23,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 23,
                "scopes": "metaRevisionExcludeGlobMarker"
              },
              {
                "startIndex": 25,
                "scopes": "metaRevisionReferencePath"
              },
              {
                "startIndex": 43,
                "scopes": "metaRevisionWildcard"
              },
              {
                "startIndex": 44,
                "scopes": "metaRevisionLabel"
              }
            ]
        `)
    })

    test('decorate search context value', () => {
        expect(getTokens(toSuccess(scanSearchQuery('context:@user')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 7,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 8,
                "scopes": "metaContextPrefix"
              },
              {
                "startIndex": 9,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('returns path separator tokens for regexp values', () => {
        expect(getTokens(toSuccess(scanSearchQuery('r:^github.com/sourcegraph@HEAD f:a/b/')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpAssertion"
              },
              {
                "startIndex": 3,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 10,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "metaPathSeparator"
              },
              {
                "startIndex": 14,
                "scopes": "identifier"
              },
              {
                "startIndex": 25,
                "scopes": "metaRepoRevisionSeparator"
              },
              {
                "startIndex": 26,
                "scopes": "metaRevisionLabel"
              },
              {
                "startIndex": 30,
                "scopes": "whitespace"
              },
              {
                "startIndex": 31,
                "scopes": "field"
              },
              {
                "startIndex": 32,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 33,
                "scopes": "identifier"
              },
              {
                "startIndex": 34,
                "scopes": "metaPathSeparator"
              },
              {
                "startIndex": 35,
                "scopes": "identifier"
              },
              {
                "startIndex": 36,
                "scopes": "metaPathSeparator"
              }
            ]
        `)
    })

    test('returns regexp highlighting if path separators cannot be parsed', () => {
        expect(getTokens(toSuccess(scanSearchQuery('r:^github.com(/)sourcegraph')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 1,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpAssertion"
              },
              {
                "startIndex": 3,
                "scopes": "identifier"
              },
              {
                "startIndex": 9,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 10,
                "scopes": "identifier"
              },
              {
                "startIndex": 13,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 14,
                "scopes": "identifier"
              },
              {
                "startIndex": 15,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 16,
                "scopes": "identifier"
              }
            ]
        `)
    })

    test('highlight recognized predicate with body as regexp', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:contains.path(README.md)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 13,
                "scopes": "metaPredicateDot"
              },
              {
                "startIndex": 14,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 18,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 19,
                "scopes": "identifier"
              },
              {
                "startIndex": 25,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 26,
                "scopes": "identifier"
              },
              {
                "startIndex": 28,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight recognized predicate with multiple fields', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:contains.file(path:README.md content:^fix$)'))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 13,
                "scopes": "metaPredicateDot"
              },
              {
                "startIndex": 14,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 18,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 19,
                "scopes": "field"
              },
              {
                "startIndex": 23,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 24,
                "scopes": "identifier"
              },
              {
                "startIndex": 30,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 31,
                "scopes": "identifier"
              },
              {
                "startIndex": 33,
                "scopes": "whitespace"
              },
              {
                "startIndex": 34,
                "scopes": "field"
              },
              {
                "startIndex": 42,
                "scopes": "metaRegexpAssertion"
              },
              {
                "startIndex": 43,
                "scopes": "identifier"
              },
              {
                "startIndex": 46,
                "scopes": "metaRegexpAssertion"
              },
              {
                "startIndex": 47,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight repo:has.file predicate', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:has.file(path:foo content:bar)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 8,
                "scopes": "metaPredicateDot"
              },
              {
                "startIndex": 9,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 13,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 14,
                "scopes": "field"
              },
              {
                "startIndex": 18,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 19,
                "scopes": "identifier"
              },
              {
                "startIndex": 22,
                "scopes": "whitespace"
              },
              {
                "startIndex": 23,
                "scopes": "field"
              },
              {
                "startIndex": 31,
                "scopes": "identifier"
              },
              {
                "startIndex": 34,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight repo:has.topic predicate', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:has.topic(topic1)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 8,
                "scopes": "metaPredicateDot"
              },
              {
                "startIndex": 9,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 14,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 15,
                "scopes": "identifier"
              },
              {
                "startIndex": 21,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight repo:has.description predicate', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:has.description(.*)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 8,
                "scopes": "metaPredicateDot"
              },
              {
                "startIndex": 9,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 20,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 21,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 22,
                "scopes": "metaRegexpRangeQuantifier"
              },
              {
                "startIndex": 23,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight repo:has predicate', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:has(key:value)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 8,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 9,
                "scopes": "identifier"
              },
              {
                "startIndex": 12,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 13,
                "scopes": "identifier"
              },
              {
                "startIndex": 18,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight repo:has.tag predicate', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:has.tag(tag)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 8,
                "scopes": "metaPredicateDot"
              },
              {
                "startIndex": 9,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 12,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 13,
                "scopes": "identifier"
              },
              {
                "startIndex": 16,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight repo:has.key predicate', () => {
        expect(getTokens(toSuccess(scanSearchQuery('repo:has.key(key)')))).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
              },
              {
                "startIndex": 4,
                "scopes": "metaFilterSeparator"
              },
              {
                "startIndex": 5,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 8,
                "scopes": "metaPredicateDot"
              },
              {
                "startIndex": 9,
                "scopes": "metaPredicateNameAccess"
              },
              {
                "startIndex": 12,
                "scopes": "metaPredicateParenthesis"
              },
              {
                "startIndex": 13,
                "scopes": "identifier"
              },
              {
                "startIndex": 16,
                "scopes": "metaPredicateParenthesis"
              }
            ]
        `)
    })

    test('highlight regex delimited pattern for standard search', () => {
        expect(getTokens(toSuccess(scanSearchQuery('/f.*/ x', false, SearchPatternType.standard))))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 1,
                "scopes": "identifier"
              },
              {
                "startIndex": 2,
                "scopes": "metaRegexpCharacterSet"
              },
              {
                "startIndex": 3,
                "scopes": "metaRegexpRangeQuantifier"
              },
              {
                "startIndex": 4,
                "scopes": "metaRegexpDelimited"
              },
              {
                "startIndex": 5,
                "scopes": "whitespace"
              },
              {
                "startIndex": 6,
                "scopes": "identifier"
              }
            ]
        `)
    })
})
