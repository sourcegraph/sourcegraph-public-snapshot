import { normalizeHover } from './correlator'

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
