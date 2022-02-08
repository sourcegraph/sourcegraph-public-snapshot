import { encodeURIPathComponent } from '@sourcegraph/common'

import { parseLineRange, serializeBlockInput, serializeBlocks, serializeLineRange } from './serialize'

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
                        lineRange: null,
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

    it('should serialize single line range', () =>
        expect(serializeLineRange({ startLine: 123, endLine: 124 })).toStrictEqual('124'))

    it('should serialize multi line range', () =>
        expect(serializeLineRange({ startLine: 123, endLine: 321 })).toStrictEqual('124-321'))

    it('should parse single line range', () => expect(parseLineRange('124')).toEqual({ startLine: 123, endLine: 124 }))

    it('should parse multi line range', () =>
        expect(parseLineRange('124-321')).toEqual({ startLine: 123, endLine: 321 }))

    it('should serialize multiple blocks', () => {
        expect(
            serializeBlocks(
                [
                    { type: 'md', input: '# Title' },
                    { type: 'query', input: 'repo:a b' },
                ],
                SOURCEGRAPH_URL
            )
        ).toStrictEqual(`md:${encodeURIPathComponent('# Title')},query:${encodeURIComponent('repo:a b')}`)
    })
})
