import { Text } from '@codemirror/state'
import { describe, expect, it } from '@jest/globals'

import {
    isValidLineRange,
    offsetToUIPosition,
    positionToOffset,
    sortRangeValuesByStart,
    trimTrailingNewline,
    uiPositionToOffset,
    zeroToOneBasedPosition,
    zeroToOneBasedRange,
} from './utils'

describe('blob/codemirror/utils', () => {
    describe('zeroToOneBased...', () => {
        it('converts zero-based to one-based positions', () => {
            expect(zeroToOneBasedPosition({ line: 5, character: 10 })).toEqual({
                line: 6,
                character: 11,
            })
        })
        it('converts zero-based to one-based ranges', () => {
            expect(zeroToOneBasedRange({ start: { line: 5, character: 10 }, end: { line: 7, character: 3 } })).toEqual({
                start: {
                    line: 6,
                    character: 11,
                },
                end: {
                    line: 8,
                    character: 4,
                },
            })
        })
    })

    describe('positionToOffset', () => {
        const textDocument = Text.of([
            // Don't forget to account for line breaks between lines!
            '0123',
            '5678',
        ])

        it('returns a valid offset', () => {
            expect(positionToOffset(textDocument, { line: 1, character: 2 })).toBe(7)
        })

        it('null when the position is not valid inside the document', () => {
            // Out-of-range character
            expect(positionToOffset(textDocument, { line: 1, character: 20 })).toBe(null)
            // Out-of-range line
            expect(positionToOffset(textDocument, { line: 2, character: 2 })).toBe(null)
        })
    })

    describe('uiPositionToOffset', () => {
        const textDocument = Text.of([
            // Don't forget to account for line breaks between lines!
            '0123',
            '5678',
        ])

        it('returns a valid offset', () => {
            expect(uiPositionToOffset(textDocument, { line: 1, character: 2 })).toBe(1)
        })

        it('null when the position is not valid inside the document', () => {
            // Out-of-range character
            expect(uiPositionToOffset(textDocument, { line: 1, character: 20 })).toBe(null)
            // Out-of-range line
            expect(uiPositionToOffset(textDocument, { line: 3, character: 2 })).toBe(null)
        })
    })

    describe('offsetToUIPosition', () => {
        const textDocument = Text.of([
            // Don't forget to account for line breaks between lines!
            '0123',
            '5678',
        ])

        it('returns a valid position', () => {
            expect(offsetToUIPosition(textDocument, 6)).toEqual({ line: 2, character: 2 })
        })

        it('returns a valid range', () => {
            expect(offsetToUIPosition(textDocument, 2, 6)).toEqual({
                start: { line: 1, character: 3 },
                end: { line: 2, character: 2 },
            })
        })
    })

    describe('sortRangeValuesByStart', () => {
        it('sort ranges by start line and character', () => {
            expect(
                sortRangeValuesByStart([
                    { range: { start: { line: 4, character: 5 } } },
                    { range: { start: { line: 7, character: 1 } } },
                    { range: { start: { line: 4, character: 1 } } },
                ])
            ).toEqual([
                { range: { start: { line: 4, character: 1 } } },
                { range: { start: { line: 4, character: 5 } } },
                { range: { start: { line: 7, character: 1 } } },
            ])
        })
    })

    describe('isValidLineRange', () => {
        const textDocument = Text.of(['', '', '32345', '42345'])

        it('returns true if the line range is inside the document', () => {
            // Empty line
            expect(isValidLineRange({ line: 2 }, textDocument)).toBe(true)
            // Empty line + first character
            expect(isValidLineRange({ line: 1, character: 1 }, textDocument)).toBe(true)
            expect(isValidLineRange({ line: 3, character: 3 }, textDocument)).toBe(true)
            expect(isValidLineRange({ line: 3, character: 2, endLine: 4, endCharacter: 5 }, textDocument)).toBe(true)
        })

        it('returns false if the line range is outside the document', () => {
            expect(isValidLineRange({ line: 10 }, textDocument)).toBe(false)
            expect(isValidLineRange({ line: 1, character: 2 }, textDocument)).toBe(false)
            expect(isValidLineRange({ line: 3, character: 2, endLine: 4, endCharacter: 8 }, textDocument)).toBe(false)
            expect(isValidLineRange({ line: 3, character: 8, endLine: 4, endCharacter: 1 }, textDocument)).toBe(false)
        })
    })

    describe('trimTrailingNewLine', () => {
        const tests: {
            name: string
            content: string
            expected: string
        }[] = [
                {
                    name: 'should remove linebreak from file containing only line breaks',
                    content: '\n\n\n\n\n',
                    expected: '',
                },
                {
                    name: 'no change',
                    content: 'line 1\nline 2\nline 3',
                    expected: 'line 1\nline 2\nline 3',
                },
                {
                    name: 'should remove trailing new line (windows - "\\r\\n")',
                    content: 'line 1\nline 2\r\n',
                    expected: 'line 1\nline 2',
                },
                {
                    name: 'should remove trailing new line (linux/unix - "\\n")',
                    content: 'let x = 12\nlet y = 4\nconsole.log(x * y) // prints 48\n',
                    expected: 'let x = 12\nlet y = 4\nconsole.log(x * y) // prints 48',
                },
                {
                    name: 'should remove one of the two trailing new lines',
                    content: "var test = 'test'\n\n\n",
                    expected: "var test = 'test'\n\n",
                },
            ]

        for (let i = 0; i < tests.length; i++) {
            let t = tests[i]
            it(t.name, () => {
                expect(trimTrailingNewline(t.content)).toBe(t.expected)
            })
        }
    })
})
