import { RangeSet, Text } from '@codemirror/state'
import { Decoration } from '@codemirror/view'
import { describe, expect, it } from 'vitest'

import type { DocumentHighlight } from '@sourcegraph/codeintellify'

import { documentHighlightsToRangeSet } from './document-highlights'

describe('document-highlights', () => {
    describe('documentHighlightsToRangeSet', () => {
        const textDocument = Text.of([
            // Don't forget to account for line breaks between lines!
            '012345',
            '789',
        ])
        const decoration = Decoration.mark({})

        it('converts document highlights to a range set', () => {
            const highlights: DocumentHighlight[] = [
                { range: { start: { line: 0, character: 0 }, end: { line: 0, character: 5 } } },
                { range: { start: { line: 1, character: 1 }, end: { line: 1, character: 2 } } },
            ]
            const rangeSet = documentHighlightsToRangeSet(textDocument, highlights, decoration)

            expect(rangeSet).toEqual(
                RangeSet.of([
                    { from: 0, to: 5, value: decoration },
                    { from: 8, to: 9, value: decoration },
                ])
            )
        })

        it('ignores highlights outside the document', () => {
            const highlights: DocumentHighlight[] = [
                { range: { start: { line: 0, character: 0 }, end: { line: 0, character: 5 } } },
                { range: { start: { line: 1, character: 1 }, end: { line: 1, character: 2 } } },
                { range: { start: { line: 10, character: 1 }, end: { line: 10, character: 2 } } },
            ]
            const rangeSet = documentHighlightsToRangeSet(textDocument, highlights, decoration)

            expect(rangeSet).toEqual(
                RangeSet.of([
                    { from: 0, to: 5, value: decoration },
                    { from: 8, to: 9, value: decoration },
                ])
            )
        })

        it('sorts highlights by start position', () => {
            const highlights: DocumentHighlight[] = [
                { range: { start: { line: 1, character: 1 }, end: { line: 1, character: 2 } } },
                { range: { start: { line: 0, character: 0 }, end: { line: 0, character: 5 } } },
            ]
            const rangeSet = documentHighlightsToRangeSet(textDocument, highlights, decoration)

            expect(rangeSet).toEqual(
                RangeSet.of([
                    { from: 0, to: 5, value: decoration },
                    { from: 8, to: 9, value: decoration },
                ])
            )
        })
    })
})
