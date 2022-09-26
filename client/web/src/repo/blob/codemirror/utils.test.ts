import { Text } from '@codemirror/state'

import {
    isValidLineRange,
    offsetToUIPosition,
    positionToOffset,
    rangesContain,
    sortRangeValuesByStart,
    uiPositionToOffset,
} from './utils'

describe('blob/codemirror/utils', () => {
    describe('rangeContains', () => {
        const ranges = [
            { from: 10, to: 20 },
            { from: 30, to: 40 },
        ]

        it('returns true when the point is within one of the specific ranges (inclusively)', () => {
            expect(rangesContain(ranges, 15)).toBe(true)
            expect(rangesContain(ranges, 20)).toBe(true)
        })

        it('returns false when the point is within one of the specific ranges (inclusively)', () => {
            expect(rangesContain(ranges, 25)).toBe(false)
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
})
