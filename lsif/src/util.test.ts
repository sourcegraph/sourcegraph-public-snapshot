import { assertDefined } from './util'

describe('assertDefined', () => {
    it('should return first defined value', () => {
        const map1 = new Map<string, string>()
        const map2 = new Map<string, string>()

        map2.set('foo', 'baz')
        expect(assertDefined('foo', '', map1, map2)).toEqual('baz')
        map1.set('foo', 'bar')
        expect(assertDefined('foo', '', map1, map2)).toEqual('bar')
    })
})
