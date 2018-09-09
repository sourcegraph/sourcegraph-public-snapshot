import * as assert from 'assert'
import { TextDocumentDecoration } from '../protocol/decoration'
import { createValidator } from './validator'

const validator = createValidator()

const FIXTURE = {
    TextDocumentDecoration: {
        range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
        isWholeLine: true,
        backgroundColor: 'blue',
    } as TextDocumentDecoration,
}

describe('Validator', () => {
    it('asTextDocumentDecoration', () => {
        const input: TextDocumentDecoration = FIXTURE.TextDocumentDecoration
        assert.deepStrictEqual(validator.asTextDocumentDecoration(input), input)
    })

    it('asTextDocumentDecorations', () => {
        const input: TextDocumentDecoration[] = [FIXTURE.TextDocumentDecoration, FIXTURE.TextDocumentDecoration]
        assert.deepStrictEqual(validator.asTextDocumentDecorations(input), input)
        assert.deepStrictEqual(validator.asTextDocumentDecorations([]), [])
        assert.strictEqual(validator.asTextDocumentDecorations(undefined), null)
        assert.strictEqual(validator.asTextDocumentDecorations(null), null)
    })
})
