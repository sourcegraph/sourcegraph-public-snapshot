import { DefaultMap } from './default-map'

describe('DefaultMap', () => {
    it('should leave get unchanged', () => {
        const map = new DefaultMap<string, string>(() => 'bar')
        expect(map.get('foo')).toBeUndefined()
    })

    it('should create values on access', () => {
        const map = new DefaultMap<string, string>(() => 'bar')
        expect(map.getOrDefault('foo')).toEqual('bar')
    })

    it('should respect explicit set', () => {
        const map = new DefaultMap<string, string>(() => 'bar')
        map.set('foo', 'baz')
        expect(map.getOrDefault('foo')).toEqual('baz')
    })

    it('should support nested gets', () => {
        const map = new DefaultMap<string, DefaultMap<string, string[]>>(
            () => new DefaultMap<string, string[]>(() => [])
        )

        map.getOrDefault('foo')
            .getOrDefault('bar')
            .push('baz')

        map.getOrDefault('foo')
            .getOrDefault('bar')
            .push('bonk')

        const inner = map.get('foo')
        expect(inner?.get('bar')).toEqual(['baz', 'bonk'])
    })
})
