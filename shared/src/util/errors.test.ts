import { asError, isErrorLike } from './errors'

describe('isErrorLike', () => {
    test('reports true for Error values', () => expect(isErrorLike(new Error('m'))).toBeTruthy())
    test('reports true for error-like values', () => expect(isErrorLike({ message: 'm' })).toBeTruthy())
    test('reports false for non-error-like values', () => expect(!isErrorLike('m')).toBeTruthy())
})

describe('asError', () => {
    test('preserves Error values', () => {
        const err = new Error('m')
        expect(asError(err)).toBe(err)
    })

    test('creates Error values from error-like values', () => {
        const err = asError({ message: 'm' })
        expect(isErrorLike(err)).toBeTruthy()
        expect(err.message).toBe('m')
    })

    test('creates Error values from strings', () => {
        const err = asError('m')
        expect(isErrorLike(err)).toBeTruthy()
        expect(err.message).toBe('m')
    })
})
