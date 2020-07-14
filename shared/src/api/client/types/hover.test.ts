import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { Range } from '@sourcegraph/extension-api-types'
import { fromHoverMerged, HoverMerged } from './hover'

const FIXTURE_RANGE: Range = { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } }
const EMPTY_HOVER: HoverMerged = {alerts: [], contents: []}

describe('HoverMerged', () => {
    describe('from', () => {
        test('0 hovers', () => expect(fromHoverMerged([])).toEqual(EMPTY_HOVER))
        test('empty hovers', () => expect(fromHoverMerged([null, undefined])).toEqual(EMPTY_HOVER))
        test('empty string hovers', () => expect(fromHoverMerged([{ contents: { value: '' } }])).toEqual(EMPTY_HOVER))
        test('1 MarkupContent', () =>
            expect(fromHoverMerged([{ contents: { kind: MarkupKind.Markdown, value: 'x' } }])).toEqual({
                contents: [{ kind: MarkupKind.Markdown, value: 'x' }],
                alerts: [],
            }))
        test('2 MarkupContents', () =>
            expect(
                fromHoverMerged([
                    { contents: { kind: MarkupKind.Markdown, value: 'x' }, range: FIXTURE_RANGE },
                    { contents: { kind: MarkupKind.Markdown, value: 'y' } },
                ])
            ).toEqual({
                contents: [
                    { kind: MarkupKind.Markdown, value: 'x' },
                    { kind: MarkupKind.Markdown, value: 'y' },
                ],
                range: FIXTURE_RANGE,
                alerts: [],
            }))
        test('1 Alert', () =>
            expect(
                fromHoverMerged([{
                    contents: { kind: MarkupKind.Markdown, value: 'x' },
                    alerts: [{ summary: { kind: MarkupKind.PlainText, value: 'x' }}],
                }])
            ).toEqual({
                contents: [
                    { kind: MarkupKind.Markdown, value: 'x' },
                ],
                alerts: [
                    { summary: { kind: MarkupKind.PlainText, value: 'x' }},
                ],
            }))
        test('2 Alerts', () =>
            expect(
                fromHoverMerged([{
                    contents: { kind: MarkupKind.Markdown, value: 'x' },
                    alerts: [
                        { summary: { kind: MarkupKind.PlainText, value: 'x' }},
                        { summary: { kind: MarkupKind.PlainText, value: 'y' }},
                    ],
                }])
            ).toEqual({
                contents: [
                    { kind: MarkupKind.Markdown, value: 'x' },
                ],
                alerts: [
                    { summary: { kind: MarkupKind.PlainText, value: 'x' }},
                    { summary: { kind: MarkupKind.PlainText, value: 'y' }},
                ],
            }))
    })
})
