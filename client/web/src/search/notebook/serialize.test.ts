import { serializeBlockInput } from './serialize'

const SOURCEGRAPH_URL = 'https://sourcegraph.com'

describe('serialize', () => {
    it('should serialize empty markdown text', () => {
        expect(serializeBlockInput({ type: 'md', input: '' }, SOURCEGRAPH_URL)).toStrictEqual('')
    })

    it('should serialize markdown text', () => {
        expect(serializeBlockInput({ type: 'md', input: '# Title' }, SOURCEGRAPH_URL)).toStrictEqual('# Title')
    })

    it('should serialize empty query', () => {
        expect(serializeBlockInput({ type: 'query', input: '' }, SOURCEGRAPH_URL)).toStrictEqual('')
    })

    it('should serialize a query', () => {
        expect(serializeBlockInput({ type: 'query', input: 'repo:a b' }, SOURCEGRAPH_URL)).toStrictEqual('repo:a b')
    })

    it('should serialize a file without range', () => {
        expect(
            serializeBlockInput(
                {
                    type: 'file',
                    input: {
                        repositoryName: 'github.com/sourcegraph/sourcegraph',
                        revision: 'feature',
                        filePath: 'client/web/index.ts',
                    },
                },
                SOURCEGRAPH_URL
            )
        ).toStrictEqual(`${SOURCEGRAPH_URL}/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts`)
    })

    it('should serialize a file with range', () => {
        expect(
            serializeBlockInput(
                {
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
                },
                SOURCEGRAPH_URL
            )
        ).toStrictEqual(
            `${SOURCEGRAPH_URL}/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts?L101-123`
        )
    })
})
