import { threadsQueryMatches, threadsQueryWithValues } from './url'

describe('threadsQueryWithValues', () => {
    test('constructs threads query', () => {
        expect(threadsQueryWithValues('', { a: ['1'] })).toBe('a:1')
        expect(threadsQueryWithValues('b', { a: ['1'] })).toBe('b a:1')
        expect(threadsQueryWithValues('b', { a: ['1'], b: ['2', '3'] })).toBe('b a:1 b:2 b:3')
        expect(threadsQueryWithValues('a:1', { a: ['2'] })).toBe('a:2')
        expect(threadsQueryWithValues('a:1', { a: ['1', '2'] })).toBe('a:1 a:2')
        expect(threadsQueryWithValues('a:1', { a: null })).toBe('')
        expect(threadsQueryWithValues('a:1 c', { a: null })).toBe('c')
        expect(threadsQueryWithValues('b:2 a:1 ', { a: ['3'], b: ['4'] })).toBe('a:3 b:4')
    })
})

describe('threadsQueryMatches', () => {
    test('reports whether threads query matches', () => {
        expect(threadsQueryMatches('', { a: '1' })).toBe(false)
        expect(threadsQueryMatches('a:1', { a: '1' })).toBe(true)
        expect(threadsQueryMatches('a:1 b', { a: '1', b: '' })).toBe(false)
        expect(threadsQueryMatches('a:1 b:2', { a: '1', b: '2' })).toBe(true)
        expect(threadsQueryMatches('a:1 b:2', { a: '1', b: '3' })).toBe(false)
        expect(threadsQueryMatches('a:1 b:2', {})).toBe(true)
    })
})
