import { Range } from '@sourcegraph/extension-api-types'
import { MarkupKind } from 'sourcegraph'
import { HoverMerged } from './hover'

const FIXTURE_RANGE: Range = { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } }

describe('HoverMerged', () => {
    describe('from', () => {
        test('0 hovers', () => expect(HoverMerged.from([])).toBeNull())
        test('empty hovers', () => expect(HoverMerged.from([null, undefined])).toBeNull())
        test('empty string hovers', () => expect(HoverMerged.from([{ contents: { value: '' } }])).toBeNull())
        test('{language, value}', () =>
            expect(HoverMerged.from([{ contents: { language: 'l', value: 'x' } }])).toEqual({
                contents: [{ kind: MarkupKind.PlainText, value: 'x' }],
            }))
        test('backcompat {language, value}', () =>
            expect(
                HoverMerged.from([{ contents: 'z' as any, __backcompatContents: [{ language: 'l', value: 'x' }] }])
            ).toEqual({
                contents: [{ kind: MarkupKind.Markdown, value: '```l\nx\n```\n' }],
            }))
        test('1 MarkupContent', () =>
            expect(HoverMerged.from([{ contents: { kind: MarkupKind.Markdown, value: 'x' } }])).toEqual({
                contents: [{ kind: MarkupKind.Markdown, value: 'x' }],
            }))
        test('2 MarkupContents', () =>
            expect(
                HoverMerged.from([
                    { contents: { kind: MarkupKind.Markdown, value: 'x' }, range: FIXTURE_RANGE },
                    { contents: { kind: MarkupKind.Markdown, value: 'y' } },
                ])
            ).toEqual({
                contents: [{ kind: MarkupKind.Markdown, value: 'x' }, { kind: MarkupKind.Markdown, value: 'y' }],
                range: FIXTURE_RANGE,
            }))
    })
})
