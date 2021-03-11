import { Token } from './token'
import { getMonacoTokens } from './decoratedToken'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('r:a (f:b and c)')))).toMatchInlineSnapshot(`
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
        `)
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('^oh\\.hai$', false, SearchPatternType.regexp)), true))
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('a*?(b)+', false, SearchPatternType.regexp)), true))
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('b{1} c{1,2} d{3,}?', false, SearchPatternType.regexp)), true))
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('((a) or b)', false, SearchPatternType.regexp)), true))
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('(?:a(?:b))', false, SearchPatternType.regexp)), true))
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('([a-z])', false, SearchPatternType.regexp)), true))
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('[a-z][--z][--z]', false, SearchPatternType.regexp)), true))
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
        expect(
            getMonacoTokens(
                toSuccess(scanSearchQuery('[--\\\\abc] \\|\\.|\\(\\)', false, SearchPatternType.regexp)),
                true
            )
        ).toMatchInlineSnapshot(`
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
        expect(
            getMonacoTokens(toSuccess(scanSearchQuery('repo:foo :[~a?|b*]', false, SearchPatternType.structural)), true)
        ).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('repo:foo@HEAD:v1.2:3')), true)).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('repo:foo rev:HEAD:v1.2:3')), true)).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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

    test('decorate repo revision syntax, path with wildcard and negation', () => {
        expect(getMonacoTokens(toSuccess(scanSearchQuery('repo:foo@*refs/heads/*:*!refs/heads/release*')), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('context:@user')), true)).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('r:^github.com/sourcegraph@HEAD f:a/b/')), true))
            .toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
        expect(getMonacoTokens(toSuccess(scanSearchQuery('r:^github.com(/)sourcegraph')), true)).toMatchInlineSnapshot(`
            [
              {
                "startIndex": 0,
                "scopes": "field"
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
})
