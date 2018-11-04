import * as assert from 'assert'
import { DocumentSelector } from 'sourcegraph'
import { match, score, TextDocumentItem } from './textDocument'

const FIXTURE_TEXT_DOCUMENT: TextDocumentItem = { uri: 'file:///f', languageId: 'l', text: '' }

describe('match', () => {
    it('reports true if any selectors match', () => {
        assert.ok(match([{ language: 'x' }, '*'], FIXTURE_TEXT_DOCUMENT))
    })

    it('reports false if no selectors match', () => {
        assert.strictEqual(match(['x'], FIXTURE_TEXT_DOCUMENT), false)
        assert.strictEqual(match([], FIXTURE_TEXT_DOCUMENT), false)
    })

    it('supports iterators for the document selectors', () => {
        let i = 0
        const iterator = {
            [Symbol.iterator](): IterableIterator<DocumentSelector> {
                return { [Symbol.iterator]: iterator[Symbol.iterator], next: () => ({ value: ['*'], done: i++ === 1 }) }
            },
        }
        assert.ok(match(iterator as IterableIterator<DocumentSelector>, FIXTURE_TEXT_DOCUMENT))
    })
})

describe('score', () => {
    it('matches', () => {
        assert.strictEqual(score(['l'], 'file:///f', 'l'), 10)
        assert.strictEqual(score(['*'], 'file:///f', 'l'), 5)
        assert.strictEqual(score(['x'], 'file:///f', 'l'), 0)
        assert.strictEqual(score([{ scheme: 'file' }], 'file:///f', 'l'), 10)
        assert.strictEqual(score([{ scheme: '*' }], 'file:///f', 'l'), 5)
        assert.strictEqual(score([{ scheme: 'x' }], 'file:///f', 'l'), 0)
        assert.strictEqual(score([{ pattern: 'file:///*.txt' }], 'file:///f.txt', 'l'), 10)
        assert.strictEqual(score([{ pattern: '**/*.txt' }], 'file:///f.txt', 'l'), 10)
        assert.strictEqual(score([{ pattern: '*.txt' }], 'file:///f.txt', 'l'), 5)
        assert.strictEqual(score([{ pattern: 'f.txt' }], 'file:///f.txt', 'l'), 10)
        assert.strictEqual(score([{ pattern: 'x' }], 'file:///f.txt', 'l'), 0)
        assert.strictEqual(score([{ pattern: 'f.txt', language: 'x' }], 'file:///f.txt', 'l'), 0)
        assert.strictEqual(score([{ language: 'x' }], 'file:///f.txt', 'l'), 0)
        assert.strictEqual(score([{ language: 'l' }], 'file:///f.txt', 'l'), 10)
        assert.strictEqual(score([{ language: '*' }], 'file:///f.txt', 'l'), 5)
    })
})
