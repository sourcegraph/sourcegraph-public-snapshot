import { reachableMonikers, normalizeHover, DefaultMap } from './importer'
import { Id } from 'lsif-protocol'

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

        expect(map.get('foo').get('bar')).toEqual(['baz', 'bonk'])
    })
})

describe('normalizeHover', () => {
    it('should handle all lsp.Hover types', () => {
        expect(normalizeHover({ contents: 'foo' })).toEqual('foo')
        expect(normalizeHover({ contents: { language: 'typescript', value: 'bar' } })).toEqual(
            '```typescript\nbar\n```'
        )
        expect(normalizeHover({ contents: { kind: 'markdown', value: 'baz' } })).toEqual('baz')
        expect(
            normalizeHover({
                contents: ['foo', { language: 'typescript', value: 'bar' }],
            })
        ).toEqual('foo\n\n---\n\n```typescript\nbar\n```')
    })
})

describe('reachableMonikers', () => {
    it('should traverse moniker relation graph', () => {
        const monikerSets = new Map<Id, Set<Id>>()
        monikerSets.set(1, new Set<Id>([2]))
        monikerSets.set(2, new Set<Id>([1, 4]))
        monikerSets.set(3, new Set<Id>([4]))
        monikerSets.set(4, new Set<Id>([2, 3]))
        monikerSets.set(5, new Set<Id>([6]))
        monikerSets.set(6, new Set<Id>([5]))

        expect(reachableMonikers(monikerSets, 1)).toEqual(new Set<Id>([1, 2, 3, 4]))
    })
})
