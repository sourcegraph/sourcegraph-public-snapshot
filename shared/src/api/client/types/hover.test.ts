import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { Range } from '@sourcegraph/extension-api-types'
import { fromHoverMerged } from './hover'

const FIXTURE_RANGE: Range = { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } }

describe('HoverMerged', () => {
    describe('from', () => {
        test('0 hovers', () => expect(fromHoverMerged([])).toBeNull())
        test('empty hovers', () => expect(fromHoverMerged([null, undefined])).toBeNull())
        test('empty string hovers', () => expect(fromHoverMerged([{ contents: { value: '' } }])).toBeNull())
        test('backcompat {language, value}', () =>
            expect(
                fromHoverMerged([{ contents: 'z' as any, __backcompatContents: [{ language: 'l', value: 'x' }] }])
            ).toEqual({
                contents: [{ kind: MarkupKind.Markdown, value: '```l\nx\n```\n' }],
            }))
        test('backcompat string', () =>
            expect(fromHoverMerged([{ contents: 'z' as any, __backcompatContents: ['x'] }])).toEqual({
                contents: [{ kind: MarkupKind.Markdown, value: 'x' }],
            }))
        test('1 MarkupContent', () =>
            expect(fromHoverMerged([{ contents: { kind: MarkupKind.Markdown, value: 'x' } }])).toEqual({
                contents: [{ kind: MarkupKind.Markdown, value: 'x' }],
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
            }))
    })
})
