import { serializeBlockInput } from './serialize'

describe('serialize', () => {
    it('should serialize empty markdown text', () => {
        expect(serializeBlockInput({ type: 'md', input: '' })).toStrictEqual('')
    })

    it('should serialize markdown text', () => {
        expect(serializeBlockInput({ type: 'md', input: '# Title' })).toStrictEqual('# Title')
    })

    it('should serialize empty query', () => {
        expect(serializeBlockInput({ type: 'query', input: '' })).toStrictEqual('')
    })

    it('should serialize a query', () => {
        expect(serializeBlockInput({ type: 'query', input: 'repo:a b' })).toStrictEqual('repo:a b')
    })

    it('should serialize a file without range', () => {
        expect(
            serializeBlockInput({
                type: 'file',
                input: {
                    repositoryName: 'github.com/sourcegraph/sourcegraph',
                    revision: 'feature',
                    filePath: 'client/web/index.ts',
                },
            })
        ).toStrictEqual('/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts')
    })

    it('should serialize a file with range', () => {
        expect(
            serializeBlockInput({
                type: 'file',
                input: {
                    repositoryName: 'github.com/sourcegraph/sourcegraph',
                    revision: 'feature',
                    filePath: 'client/web/index.ts',
                    lineRange: {
                        startLine: 100,
                        endLine: 123,
                    },
                },
            })
        ).toStrictEqual('/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts?L101-123')
    })
})
