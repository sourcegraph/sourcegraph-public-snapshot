import { convertMarkdownToBlocks } from './convertMarkdownToBlocks'

describe('convertMarkdownToBlocks', () => {
    it('should handle empty markdown', () => {
        expect(convertMarkdownToBlocks('')).toStrictEqual([])
    })

    it('should handle only markdown text', () => {
        const markdown = `# Title

## Second title

* L1
* L2

Paragraph`
        expect(convertMarkdownToBlocks(markdown)).toStrictEqual([{ type: 'md', input: markdown }])
    })

    it('should handle a single file link', () => {
        const markdown = 'https://sourcegraph.com/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts'

        expect(convertMarkdownToBlocks(markdown)).toStrictEqual([
            {
                type: 'file',
                input: {
                    repositoryName: 'github.com/sourcegraph/sourcegraph',
                    revision: 'feature',
                    filePath: 'client/web/index.ts',
                    lineRange: undefined,
                },
            },
        ])
    })

    it('should handle interleaved markdown, query, and file blocks', () => {
        const markdown = `# Title

\`\`\`sourcegraph
my query
\`\`\`

## Second title

Paragraph with list:

* 1
* 2
* 3

\`\`\`sourcegraph
my second query
\`\`\`

https://sourcegraph.com/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts

## Second title v2

Link to a file is inside text https://sourcegraph.com/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts

https://sourcegraph.com/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts?L101-123

### Third title

https://example.com/a/b

`

        expect(convertMarkdownToBlocks(markdown)).toStrictEqual([
            { type: 'md', input: '# Title\n\n' },
            { type: 'query', input: 'my query' },
            { type: 'md', input: '## Second title\n\nParagraph with list:\n\n* 1\n* 2\n* 3\n\n' },
            { type: 'query', input: 'my second query' },
            {
                type: 'file',
                input: {
                    repositoryName: 'github.com/sourcegraph/sourcegraph',
                    revision: 'feature',
                    filePath: 'client/web/index.ts',
                    lineRange: undefined,
                },
            },
            {
                type: 'md',
                input:
                    '## Second title v2\n\nLink to a file is inside text https://sourcegraph.com/github.com/sourcegraph/sourcegraph@feature/-/blob/client/web/index.ts\n\n',
            },
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
            { type: 'md', input: '### Third title\n\nhttps://example.com/a/b\n\n' },
        ])
    })
})
