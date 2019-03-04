import { DocumentSelector, TextDocument } from 'sourcegraph'
import { match, score } from './textDocument'

const FIXTURE_TEXT_DOCUMENT: TextDocument = { uri: 'file:///f', languageId: 'l', text: '' }

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
