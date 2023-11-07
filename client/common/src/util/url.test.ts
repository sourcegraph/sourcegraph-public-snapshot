import { describe, expect, it, test } from '@jest/globals'

import { lprToSelectionsZeroIndexed, encodeURIPathComponent, appendLineRangeQueryParameter } from './url'

/**
 * Asserts deep object equality using node's assert.deepEqual, except it (1) ignores differences in the
 * prototype (because that causes 2 object literals to fail the test) and (2) treats undefined properties as
 * missing.
 */
function assertDeepStrictEqual(actual: any, expected: any): void {
    actual = JSON.parse(JSON.stringify(actual))
    expected = JSON.parse(JSON.stringify(expected))
    expect(actual).toEqual(expected)
}

describe('encodeURIPathComponent', () => {
    it('encodes all special characters except slashes and the plus sign', () => {
        expect(encodeURIPathComponent('hello world+/+some_special_characters_:_#_?_%_@')).toBe(
            'hello%20world+/+some_special_characters_%3A_%23_%3F_%25_%40'
        )
    })
})

describe('lprToSelectionsZeroIndexed', () => {
    test('converts an LPR with only a start line', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 5,
            }),
            [
                {
                    start: {
                        line: 4,
                        character: 0,
                    },
                    end: {
                        line: 4,
                        character: 0,
                    },
                    anchor: {
                        line: 4,
                        character: 0,
                    },
                    active: {
                        line: 4,
                        character: 0,
                    },
                    isReversed: false,
                },
            ]
        )
    })

    test('converts an LPR with a line and a character', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 5,
                character: 45,
            }),
            [
                {
                    start: {
                        line: 4,
                        character: 44,
                    },
                    end: {
                        line: 4,
                        character: 44,
                    },
                    anchor: {
                        line: 4,
                        character: 44,
                    },
                    active: {
                        line: 4,
                        character: 44,
                    },
                    isReversed: false,
                },
            ]
        )
    })

    test('converts an LPR with a start and end line', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 12,
                endLine: 15,
            }),
            [
                {
                    start: {
                        line: 11,
                        character: 0,
                    },
                    end: {
                        line: 14,
                        character: 0,
                    },
                    anchor: {
                        line: 11,
                        character: 0,
                    },
                    active: {
                        line: 14,
                        character: 0,
                    },
                    isReversed: false,
                },
            ]
        )
    })

    test('converts an LPR with a start and end line and characters', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 12,
                character: 30,
                endLine: 15,
                endCharacter: 60,
            }),
            [
                {
                    start: {
                        line: 11,
                        character: 29,
                    },
                    end: {
                        line: 14,
                        character: 59,
                    },
                    anchor: {
                        line: 11,
                        character: 29,
                    },
                    active: {
                        line: 14,
                        character: 59,
                    },
                    isReversed: false,
                },
            ]
        )
    })
})

describe('appendLineRangeQueryParameter', () => {
    it('appends line range to the start of query with existing parameters', () => {
        expect(
            appendLineRangeQueryParameter(
                '/github.com/sourcegraph/sourcegraph/-/blob/.gitattributes?test=test',
                'L24:24'
            )
        ).toBe('/github.com/sourcegraph/sourcegraph/-/blob/.gitattributes?L24:24&test=test')
    })
})
