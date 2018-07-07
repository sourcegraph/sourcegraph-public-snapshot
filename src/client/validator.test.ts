import * as assert from 'assert'
import { Position, Range } from 'vscode-languageserver-types'
import { TextDocumentDecoration } from '../protocol/decorations'
import { createValidator } from './validator'

const validator = createValidator()

const FIXTURE = {
    TextDocumentDecoration: {
        range: Range.create(Position.create(1, 2), Position.create(3, 4)),
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
