import { CompletionsCache } from './cache'

describe('CompletionsCache', () => {
    it('returns the cached completion items', () => {
        const cache = new CompletionsCache()
        cache.add([{ prefix: 'foo\n', content: 'bar', messages: [] }])

        expect(cache.get('foo\n')).toEqual([{ prefix: 'foo\n', content: 'bar', messages: [] }])
    })

    it('returns the cached items when the prefix includes characters from the completion', () => {
        const cache = new CompletionsCache()
        cache.add([{ prefix: 'foo\n', content: 'bar', messages: [] }])

        expect(cache.get('foo\nb')).toEqual([{ prefix: 'foo\nb', content: 'ar', messages: [] }])
        expect(cache.get('foo\nba')).toEqual([{ prefix: 'foo\nba', content: 'r', messages: [] }])
    })

    it('returns the cached items when the prefix has less whitespace', () => {
        const cache = new CompletionsCache()
        cache.add([{ prefix: 'foo \n  ', content: 'bar', messages: [] }])

        expect(cache.get('foo \n  ')).toEqual([{ prefix: 'foo \n  ', content: 'bar', messages: [] }])
        expect(cache.get('foo \n ')).toEqual([{ prefix: 'foo \n ', content: 'bar', messages: [] }])
        expect(cache.get('foo \n')).toEqual([{ prefix: 'foo \n', content: 'bar', messages: [] }])
        expect(cache.get('foo ')).toEqual(undefined)
    })

    it('has a lookup function for untrimmed prefixes', () => {
        const cache = new CompletionsCache()
        cache.add([{ prefix: 'foo\n  ', content: 'baz', messages: [] }])

        expect(cache.get('foo\n  ', false)).toEqual([
            {
                prefix: 'foo\n  ',
                content: 'baz',
                messages: [],
            },
        ])
        expect(cache.get('foo\n ', false)).toEqual(undefined)
    })
})
