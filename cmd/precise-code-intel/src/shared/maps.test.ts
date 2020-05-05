import { mustGetFromEither } from './maps'

describe('mustGetFromEither', () => {
    it('should return first defined value', () => {
        const map1 = new Map<string, string>()
        const map2 = new Map<string, string>()

        map2.set('foo', 'baz')
        expect(mustGetFromEither(map1, map2, 'foo', '')).toEqual('baz')
        map1.set('foo', 'bar')
        expect(mustGetFromEither(map1, map2, 'foo', '')).toEqual('bar')
    })
})
