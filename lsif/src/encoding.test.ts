import { createFilter, testFilter, gzipJSON, gunzipJSON } from './encoding'

describe('testFilter', () => {
    it('should test set membership', async () => {
        const filter = await createFilter(['foo', 'bar', 'baz'])
        expect(await testFilter(filter, 'foo')).toBeTruthy()
        expect(await testFilter(filter, 'bar')).toBeTruthy()
        expect(await testFilter(filter, 'baz')).toBeTruthy()
        expect(await testFilter(filter, 'bonk')).toBeFalsy()
        expect(await testFilter(filter, 'quux')).toBeFalsy()
    })
})

describe('gzipJSON', () => {
    it('should preserve maps', async () => {
        const m = new Map<string, number>([['a', 1], ['b', 2], ['c', 3]])

        const value = {
            foo: [1, 2, 3],
            bar: ['abc', 'xyz'],
            baz: m,
        }

        const encoded = await gzipJSON(value)
        const decoded = await gunzipJSON(encoded)
        expect(decoded).toEqual(value)
    })

    it('should preserve sets', async () => {
        const s = new Set<number>([1, 2, 3, 4, 5])

        const value = {
            foo: [1, 2, 3],
            bar: ['abc', 'xyz'],
            baz: s,
        }

        const encoded = await gzipJSON(value)
        const decoded = await gunzipJSON(encoded)
        expect(decoded).toEqual(value)
    })
})
