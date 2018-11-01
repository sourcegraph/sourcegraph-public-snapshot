import assert from 'assert'
import { isErrorLike } from './errors'
import { parseJSONCOrError } from './util'

describe('parseJSONCOrError', () => {
    it('parses valid JSON', () => assert.deepStrictEqual(parseJSONCOrError('{"a":1}'), { a: 1 }))
    it('parses valid JSONC', () => assert.deepStrictEqual(parseJSONCOrError('{/*x*/"a":1,}'), { a: 1 }))
    it('returns an error value for invalid input', () => {
        const value = parseJSONCOrError('.')
        assert.ok(isErrorLike(value))
    })
})
