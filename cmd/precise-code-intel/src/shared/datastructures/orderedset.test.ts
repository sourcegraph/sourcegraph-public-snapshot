import { OrderedSet } from './orderedset'

describe('OrderedSet', () => {
    it('should not contain duplicates', () => {
        const set = new OrderedSet<string>(value => value)
        set.push('foo')
        set.push('foo')
        set.push('bar')
        set.push('bar')

        expect(set.values).toEqual(['foo', 'bar'])
    })

    it('should retain insertion order', () => {
        const set = new OrderedSet<string>(value => value)
        set.push('bonk')
        set.push('baz')
        set.push('foo')
        set.push('bar')
        set.push('bar')
        set.push('baz')
        set.push('foo')
        set.push('bonk')

        expect(set.values).toEqual(['bonk', 'baz', 'foo', 'bar'])
    })
})
