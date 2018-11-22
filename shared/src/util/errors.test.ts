import assert from 'assert'
import { asError, isErrorLike } from './errors'

describe('isErrorLike', () => {
    it('reports true for Error values', () => assert.ok(isErrorLike(new Error('m'))))
    it('reports true for error-like values', () => assert.ok(isErrorLike({ message: 'm' })))
    it('reports false for non-error-like values', () => assert.ok(!isErrorLike('m')))
})

describe('asError', () => {
    it('preserves Error values', () => {
        const err = new Error('m')
        assert.strictEqual(asError(err), err)
    })

    it('creates Error values from error-like values', () => {
        const err = asError({ message: 'm' })
        assert.ok(isErrorLike(err))
        assert.strictEqual(err.message, 'm')
    })

    it('creates Error values from strings', () => {
        const err = asError('m')
        assert.ok(isErrorLike(err))
        assert.strictEqual(err.message, 'm')
    })
})
