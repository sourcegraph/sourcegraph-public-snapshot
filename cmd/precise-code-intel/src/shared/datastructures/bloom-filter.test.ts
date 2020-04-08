import { createFilter, testFilter } from './bloom-filter'

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
