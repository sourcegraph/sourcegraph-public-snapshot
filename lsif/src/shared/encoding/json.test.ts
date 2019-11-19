import { gunzipJSON, gzipJSON } from './json'

describe('gzipJSON', () => {
    it('should preserve maps', async () => {
        const m = new Map<string, number>([
            ['a', 1],
            ['b', 2],
            ['c', 3],
        ])

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
