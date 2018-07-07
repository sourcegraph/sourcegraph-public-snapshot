import * as assert from 'assert'
import { score } from './textDocument'

describe('score', () => {
    it('matches', () => {
        assert.strictEqual(score(['*'], 'file:///f', 'l'), 5)
        assert.strictEqual(score([{ scheme: 'file' }], 'file:///f', 'l'), 10)
        assert.strictEqual(score(['l'], 'file:///f', 'l'), 10)
        assert.strictEqual(score([{ pattern: 'file:///*.txt' }], 'file:///f.txt', 'l'), 10)
        assert.strictEqual(score([{ pattern: '**/*.txt' }], 'file:///f.txt', 'l'), 10)
        assert.strictEqual(score([{ pattern: '*.txt' }], 'file:///f.txt', 'l'), 5)
        assert.strictEqual(score([{ pattern: 'f.txt' }], 'file:///f.txt', 'l'), 10)
        assert.strictEqual(score([{ pattern: 'f.txt', language: 'x' }], 'file:///f.txt', 'l'), 0)
    })
})
