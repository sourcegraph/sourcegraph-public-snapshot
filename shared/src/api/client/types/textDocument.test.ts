import { Position } from '@sourcegraph/extension-api-types'
import { DocumentSelector, TextDocument } from 'sourcegraph'
import { match, offsetToPosition, positionToOffset, score } from './textDocument'

const FIXTURE_TEXT_DOCUMENT: Pick<TextDocument, 'uri' | 'languageId'> = { uri: 'file:///f', languageId: 'l' }

describe('match', () => {
    test('reports true if any selectors match', () => {
        expect(match([{ language: 'x' }, '*'], FIXTURE_TEXT_DOCUMENT)).toBeTruthy()
    })

    test('reports false if no selectors match', () => {
        expect(match(['x'], FIXTURE_TEXT_DOCUMENT)).toBe(false)
        expect(match([], FIXTURE_TEXT_DOCUMENT)).toBe(false)
    })

    test('supports iterators for the document selectors', () => {
        let i = 0
        const iterator = {
            [Symbol.iterator](): IterableIterator<DocumentSelector> {
                return { [Symbol.iterator]: iterator[Symbol.iterator], next: () => ({ value: ['*'], done: i++ === 1 }) }
            },
        }
        expect(match(iterator as IterableIterator<DocumentSelector>, FIXTURE_TEXT_DOCUMENT)).toBeTruthy()
    })
})

describe('score', () => {
    test('matches', () => {
        expect(score(['l'], 'file:///f', 'l')).toBe(10)
        expect(score(['*'], 'file:///f', 'l')).toBe(5)
        expect(score(['x'], 'file:///f', 'l')).toBe(0)
        expect(score([{ scheme: 'file' }], 'file:///f', 'l')).toBe(10)
        expect(score([{ scheme: '*' }], 'file:///f', 'l')).toBe(5)
        expect(score([{ scheme: 'x' }], 'file:///f', 'l')).toBe(0)
        expect(score([{ pattern: 'file:///*.txt' }], 'file:///f.txt', 'l')).toBe(10)
        expect(score([{ pattern: '**/*.txt' }], 'file:///f.txt', 'l')).toBe(10)
        expect(score([{ pattern: '*.txt' }], 'file:///f.txt', 'l')).toBe(5)
        expect(score([{ pattern: 'f.txt' }], 'file:///f.txt', 'l')).toBe(10)
        expect(score([{ pattern: 'x' }], 'file:///f.txt', 'l')).toBe(0)
        expect(score([{ pattern: 'f.txt', language: 'x' }], 'file:///f.txt', 'l')).toBe(0)
        expect(score([{ language: 'x' }], 'file:///f.txt', 'l')).toBe(0)
        expect(score([{ language: 'l' }], 'file:///f.txt', 'l')).toBe(10)
        expect(score([{ language: '*' }], 'file:///f.txt', 'l')).toBe(5)
        expect(score([{}], 'file:///f.txt', 'l')).toBe(5)
    })
})

interface OffsetPositionTestCase {
    text: string
    offset: number
    pos: Position
}

const OFFSET_POSITION_COMMON_TESTS: OffsetPositionTestCase[] = [
    { text: 'ab\nc\nd', offset: 0, pos: { line: 0, character: 0 } },
    { text: 'ab\nc\nd', offset: 1, pos: { line: 0, character: 1 } },
    { text: 'ab\nc\nd', offset: 2, pos: { line: 0, character: 2 } },
    { text: 'ab\nc\nd', offset: 3, pos: { line: 1, character: 0 } },
    { text: 'ab\nc\nd', offset: 4, pos: { line: 1, character: 1 } },
    { text: 'ab\nc\nd', offset: 5, pos: { line: 2, character: 0 } },
    { text: 'ab\nc\nd', offset: 6, pos: { line: 2, character: 1 } },
    { text: 'ab\nc\nd\n', offset: 7, pos: { line: 3, character: 0 } },
    { text: '\n', offset: 0, pos: { line: 0, character: 0 } },
    { text: '\n', offset: 1, pos: { line: 1, character: 0 } },
    { text: '', offset: 0, pos: { line: 0, character: 0 } },
]

/** Shared test suite among multiple implementations of offset-to-position conversion functions. */
export const OFFSET_TO_POSITION_TESTS: OffsetPositionTestCase[] = [
    ...OFFSET_POSITION_COMMON_TESTS,
    { text: '', offset: 100, pos: { line: 0, character: 0 } },
    { text: 'a', offset: 100, pos: { line: 0, character: 1 } },
    { text: 'a\nb', offset: 100, pos: { line: 1, character: 1 } },
]

/** Shared test suite among multiple implementations of position-to-offset conversion functions. */
export const POSITION_TO_OFFSET_TESTS: OffsetPositionTestCase[] = [
    ...OFFSET_POSITION_COMMON_TESTS,
    { text: '', pos: { line: 100, character: 0 }, offset: 0 },
    { text: 'a', pos: { line: 100, character: 0 }, offset: 1 },
    { text: 'a\nb', pos: { line: 100, character: 0 }, offset: 3 },
]

describe('offsetToPosition', () => {
    for (const [i, { text, ...c }] of OFFSET_TO_POSITION_TESTS.entries()) {
        if (!('offset' in c)) {
            continue
        }
        test(i.toString(), () => expect(offsetToPosition(text, c.offset)).toEqual(c.pos))
    }
})

describe('positionToOffset', () => {
    for (const [i, { text, ...c }] of POSITION_TO_OFFSET_TESTS.entries()) {
        test(i.toString(), () => expect(positionToOffset(text, c.pos)).toEqual(c.offset))
    }
})
